package configs

type Config struct {
	AppName string
	Port    string
	Env     string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	JWTSecret                string
	GitHubAppSlug            string
	GitHubAppID              string
	GitHubAppPrivateKey      string
	GitHubClientID           string
	GitHubClientSecret       string
	PublicAppURL             string
	EnvironmentEncryptionKey string
}

var AppConfig Config
