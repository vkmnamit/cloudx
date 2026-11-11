package githubapp

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/vkmnamit/cloudx/configs"
)

type Handler struct {
	db     *pgx.Conn
	config configs.Config
	client *http.Client
}

type installation struct {
	ID      int64 `json:"id"`
	Account struct {
		Login string `json:"login"`
	} `json:"account"`
}

type repository struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	HTMLURL       string `json:"html_url"`
	CloneURL      string `json:"clone_url"`
	DefaultBranch string `json:"default_branch"`
	Private       bool   `json:"private"`
	Description   string `json:"description"`
}

func NewHandler(db *pgx.Conn, config configs.Config) *Handler {
	return &Handler{db: db, config: config, client: &http.Client{Timeout: 15 * time.Second}}
}

func (h *Handler) Connect(c *gin.Context) {
	if h.config.GitHubAppSlug == "" || h.config.GitHubAppID == "" || h.config.GitHubAppPrivateKey == "" || h.config.GitHubClientID == "" || h.config.GitHubClientSecret == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "GitHub App and OAuth credentials are not fully configured on this CloudX server"})
		return
	}
	organizationID := c.Query("organization_id")
	userID, _ := c.Get("user_id")
	var owned bool
	if err := h.db.QueryRow(c.Request.Context(), `SELECT EXISTS(SELECT 1 FROM organizations WHERE id=$1 AND owner_id=$2)`, organizationID, userID).Scan(&owned); err != nil || !owned {
		c.JSON(http.StatusNotFound, gin.H{"error": "organization not found"})
		return
	}
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not start GitHub connection"})
		return
	}
	state := base64.RawURLEncoding.EncodeToString(stateBytes)
	stateHash := fmt.Sprintf("%x", sha256.Sum256([]byte(state)))
	_, err := h.db.Exec(c.Request.Context(), `INSERT INTO github_app_installation_states (state_hash,user_id,organization_id,expires_at) VALUES ($1,$2,$3,NOW()+INTERVAL '10 minutes')`, stateHash, userID, organizationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save GitHub connection state"})
		return
	}
	installURL := "https://github.com/apps/" + url.PathEscape(h.config.GitHubAppSlug) + "/installations/new?state=" + url.QueryEscape(state)
	c.JSON(http.StatusOK, gin.H{"installation_url": installURL, "expires_in": 600})
}

func (h *Handler) Callback(c *gin.Context) {
	state := c.Query("state")
	installationID, err := strconv.ParseInt(c.Query("installation_id"), 10, 64)
	code := c.Query("code")
	if state == "" || code == "" || err != nil || installationID <= 0 {
		c.String(http.StatusBadRequest, "GitHub did not return a valid installation. Return to CloudX and try again.")
		return
	}
	stateHash := fmt.Sprintf("%x", sha256.Sum256([]byte(state)))
	var userID, organizationID string
	err = h.db.QueryRow(c.Request.Context(), `DELETE FROM github_app_installation_states WHERE state_hash=$1 AND expires_at>NOW() RETURNING user_id::text,organization_id::text`, stateHash).Scan(&userID, &organizationID)
	if err != nil {
		c.String(http.StatusBadRequest, "GitHub connection expired or state was invalid. Return to CloudX and try again.")
		return
	}
	if h.config.GitHubClientID == "" || h.config.GitHubClientSecret == "" {
		c.String(http.StatusServiceUnavailable, "CloudX GitHub user authorization is not configured.")
		return
	}
	userToken, err := h.exchangeUserCode(c.Request.Context(), code)
	if err != nil {
		c.String(http.StatusBadGateway, "CloudX could not verify your GitHub account. Check the GitHub App OAuth configuration.")
		return
	}
	var userInstallations struct {
		Installations []installation `json:"installations"`
	}
	if err := h.githubRequest(c.Request.Context(), http.MethodGet, "https://api.github.com/user/installations?per_page=100", userToken, nil, &userInstallations); err != nil {
		c.String(http.StatusBadGateway, "CloudX could not verify that your GitHub account can access this installation.")
		return
	}
	installationAuthorized := false
	for _, item := range userInstallations.Installations {
		if item.ID == installationID {
			installationAuthorized = true
			break
		}
	}
	if !installationAuthorized {
		c.String(http.StatusForbidden, "The GitHub account you authorized does not have access to this installation.")
		return
	}
	var account installation
	if err := h.githubRequest(c.Request.Context(), http.MethodGet, fmt.Sprintf("https://api.github.com/app/installations/%d", installationID), h.appJWT(), nil, &account); err != nil {
		c.String(http.StatusBadGateway, "CloudX could not read this GitHub installation.")
		return
	}
	_, err = h.db.Exec(c.Request.Context(), `INSERT INTO github_installations (organization_id,installation_id,account_name,access_token,expires_at) VALUES ($1,$2,$3,NULL,NOW()) ON CONFLICT (installation_id) DO UPDATE SET organization_id=EXCLUDED.organization_id,account_name=EXCLUDED.account_name,access_token=NULL,refresh_token=NULL,expires_at=NOW(),updated_at=NOW()`, organizationID, installationID, account.Account.Login)
	if err != nil {
		c.String(http.StatusInternalServerError, "GitHub was connected, but CloudX could not save the installation.")
		return
	}
	if h.config.PublicAppURL == "" {
		c.String(http.StatusOK, "GitHub is connected. Return to CloudX.")
		return
	}
	c.Redirect(http.StatusSeeOther, strings.TrimRight(h.config.PublicAppURL, "/")+"/settings?github=connected")
}

func (h *Handler) exchangeUserCode(ctx context.Context, code string) (string, error) {
	payload := map[string]string{
		"client_id":     h.config.GitHubClientID,
		"client_secret": h.config.GitHubClientSecret,
		"code":          code,
	}
	var response struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := h.githubRequest(ctx, http.MethodPost, "https://github.com/login/oauth/access_token", "", payload, &response); err != nil {
		return "", err
	}
	if response.Error != "" || response.AccessToken == "" {
		return "", errors.New("GitHub did not issue an OAuth access token")
	}
	return response.AccessToken, nil
}

func (h *Handler) Installations(c *gin.Context) {
	userID, _ := c.Get("user_id")
	rows, err := h.db.Query(c.Request.Context(), `SELECT gi.installation_id,COALESCE(gi.account_name,''),gi.organization_id::text FROM github_installations gi JOIN organizations o ON o.id=gi.organization_id WHERE o.owner_id=$1 ORDER BY gi.created_at DESC`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list GitHub installations"})
		return
	}
	defer rows.Close()
	items := make([]gin.H, 0)
	for rows.Next() {
		var id int64
		var account, orgID string
		if err := rows.Scan(&id, &account, &orgID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list GitHub installations"})
			return
		}
		items = append(items, gin.H{"installation_id": id, "account_name": account, "organization_id": orgID})
	}
	c.JSON(http.StatusOK, items)
}

// TokenForRepository returns a short-lived installation token only when the
// repository has been installed for an organization owned by this user.
func (h *Handler) TokenForRepository(ctx context.Context, userID, repositoryURL string) (string, error) {
	var installationID int64
	owner, name, ok := repositorySlug(repositoryURL)
	if !ok {
		return "", nil
	}
	err := h.db.QueryRow(ctx, `SELECT gi.installation_id FROM github_installations gi JOIN organizations o ON o.id=gi.organization_id WHERE o.owner_id=$1 AND lower(gi.account_name)=lower($2) ORDER BY gi.created_at DESC LIMIT 1`, userID, owner).Scan(&installationID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	// Restrict the installation token to the selected repository. GitHub's
	// installation token endpoint accepts repository names to enforce this.
	var scoped struct {
		Token string `json:"token"`
	}
	err = h.githubRequest(ctx, http.MethodPost, fmt.Sprintf("https://api.github.com/app/installations/%d/access_tokens", installationID), h.appJWT(), map[string]any{"repositories": []string{name}, "permissions": map[string]string{"contents": "read"}}, &scoped)
	if err != nil {
		return "", err
	}
	if scoped.Token == "" {
		return "", errors.New("GitHub returned an empty repository token")
	}
	return scoped.Token, nil
}

func repositorySlug(repositoryURL string) (owner, name string, ok bool) {
	parsed, err := url.Parse(strings.TrimSpace(repositoryURL))
	if err != nil {
		return "", "", false
	}
	path := strings.Trim(strings.TrimSuffix(parsed.Path, ".git"), "/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 || (parsed.Host != "github.com" && parsed.Host != "www.github.com") {
		return "", "", false
	}
	return parts[0], parts[1], parts[0] != "" && parts[1] != ""
}

func (h *Handler) Repositories(c *gin.Context) {
	userID, _ := c.Get("user_id")
	installationID, err := strconv.ParseInt(c.Query("installation_id"), 10, 64)
	if err != nil || installationID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "valid installation_id is required"})
		return
	}
	var owned bool
	err = h.db.QueryRow(c.Request.Context(), `SELECT EXISTS(SELECT 1 FROM github_installations gi JOIN organizations o ON o.id=gi.organization_id WHERE gi.installation_id=$1 AND o.owner_id=$2)`, installationID, userID).Scan(&owned)
	if err != nil || !owned {
		c.JSON(http.StatusNotFound, gin.H{"error": "GitHub installation not found"})
		return
	}
	token, err := h.installationToken(c.Request.Context(), installationID)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "could not authorize with GitHub; check the GitHub App key configuration"})
		return
	}
	var payload struct {
		Repositories []repository `json:"repositories"`
	}
	if err := h.githubRequest(c.Request.Context(), http.MethodGet, "https://api.github.com/installation/repositories?per_page=100", token, nil, &payload); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "could not load repositories from GitHub"})
		return
	}
	var localInstallationID string
	if err := h.db.QueryRow(c.Request.Context(), `SELECT id::text FROM github_installations WHERE installation_id=$1`, installationID).Scan(&localInstallationID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not resolve GitHub installation"})
		return
	}
	for _, repo := range payload.Repositories {
		_, _ = h.db.Exec(c.Request.Context(), `INSERT INTO github_repositories (installation_id,github_repository_id,repository_name,full_name,default_branch,private,clone_url) VALUES ($1,$2,$3,$4,$5,$6,$7) ON CONFLICT (github_repository_id) DO UPDATE SET installation_id=EXCLUDED.installation_id,repository_name=EXCLUDED.repository_name,full_name=EXCLUDED.full_name,default_branch=EXCLUDED.default_branch,private=EXCLUDED.private,clone_url=EXCLUDED.clone_url,updated_at=NOW()`, localInstallationID, repo.ID, repo.Name, repo.FullName, repo.DefaultBranch, repo.Private, repo.CloneURL)
	}
	c.JSON(http.StatusOK, payload.Repositories)
}

func (h *Handler) installationToken(ctx context.Context, installationID int64) (string, error) {
	var response struct {
		Token string `json:"token"`
	}
	err := h.githubRequest(ctx, http.MethodPost, fmt.Sprintf("https://api.github.com/app/installations/%d/access_tokens", installationID), h.appJWT(), map[string]any{"permissions": map[string]string{"contents": "read"}}, &response)
	if err != nil {
		return "", err
	}
	if response.Token == "" {
		return "", errors.New("GitHub returned an empty installation token")
	}
	return response.Token, nil
}

func (h *Handler) appJWT() string {
	keyData := []byte(strings.ReplaceAll(h.config.GitHubAppPrivateKey, `\n`, "\n"))
	block, _ := pem.Decode(keyData)
	if block == nil {
		return ""
	}
	var key *rsa.PrivateKey
	if parsed, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		key = parsed
	} else if parsedAny, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		key, _ = parsedAny.(*rsa.PrivateKey)
	}
	appID, err := strconv.ParseInt(h.config.GitHubAppID, 10, 64)
	if err != nil || key == nil {
		return ""
	}
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{"iat": now.Add(-time.Minute).Unix(), "exp": now.Add(8 * time.Minute).Unix(), "iss": appID})
	signed, _ := token.SignedString(key)
	return signed
}

func (h *Handler) githubRequest(ctx context.Context, method, endpoint, token string, body any, output any) error {
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(encoded)
	}
	request, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	if strings.Contains(endpoint, "/login/oauth/access_token") {
		request.Header.Set("Accept", "application/json")
	}
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := h.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("GitHub API returned status %d", response.StatusCode)
	}
	if output == nil {
		return nil
	}
	return json.NewDecoder(io.LimitReader(response.Body, 4<<20)).Decode(output)
}
