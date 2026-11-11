package environment

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
)

const encryptedValuePrefix = "cloudx:enc:v1:"

// secretCipher encrypts secret environment values before they reach the database.
// The application secret is domain-separated so the resulting key is not the JWT key.
type secretCipher struct {
	primary cipher.AEAD
	keys    []cipher.AEAD
}

func newSecretCipher(applicationSecrets ...string) (*secretCipher, error) {
	result := &secretCipher{}
	seen := make(map[string]struct{})
	for _, secret := range applicationSecrets {
		if secret == "" {
			continue
		}
		if _, ok := seen[secret]; ok {
			continue
		}
		seen[secret] = struct{}{}
		mac := hmac.New(sha256.New, []byte(secret))
		_, _ = mac.Write([]byte("cloudx/environment-encryption/v1"))
		block, err := aes.NewCipher(mac.Sum(nil))
		if err != nil {
			return nil, fmt.Errorf("create environment encryption key: %w", err)
		}
		aead, err := cipher.NewGCM(block)
		if err != nil {
			return nil, fmt.Errorf("create environment encryption: %w", err)
		}
		if result.primary == nil {
			result.primary = aead
		}
		result.keys = append(result.keys, aead)
	}
	if result.primary == nil {
		return nil, fmt.Errorf("ENVIRONMENT_ENCRYPTION_KEY or JWT_SECRET must be configured before storing secret environment variables")
	}
	return result, nil
}

func (c *secretCipher) encrypt(value string) (string, error) {
	nonce := make([]byte, c.primary.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("generate environment encryption nonce: %w", err)
	}
	ciphertext := c.primary.Seal(nonce, nonce, []byte(value), nil)
	return encryptedValuePrefix + base64.RawStdEncoding.EncodeToString(ciphertext), nil
}

func (c *secretCipher) decrypt(value string) (string, error) {
	if !strings.HasPrefix(value, encryptedValuePrefix) {
		// Read pre-encryption values during rollout; subsequent saves encrypt them.
		return value, nil
	}
	if c == nil || len(c.keys) == 0 {
		return "", fmt.Errorf("environment encryption is unavailable")
	}
	encoded := strings.TrimPrefix(value, encryptedValuePrefix)
	ciphertext, err := base64.RawStdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("decode encrypted environment value: %w", err)
	}
	for _, key := range c.keys {
		if len(ciphertext) < key.NonceSize() {
			continue
		}
		nonce, body := ciphertext[:key.NonceSize()], ciphertext[key.NonceSize():]
		plaintext, err := key.Open(nil, nonce, body, nil)
		if err == nil {
			return string(plaintext), nil
		}
	}
	return "", fmt.Errorf("decrypt environment value: none of the configured encryption keys matched")
}
