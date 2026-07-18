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
}

var AppConfig Config
