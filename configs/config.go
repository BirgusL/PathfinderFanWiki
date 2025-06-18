package configs

import "os"

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type Config struct {
	Port string
	DB   DBConfig
}

func Load() Config {
	return Config{
		Port: os.Getenv("PORT"),

		DB: DBConfig{
			Host:     getEnv("DB_HOST", "turntable.proxy.rlwy.net"),
			Port:     getEnv("DB_PORT", "24013"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("PGPASSWORD", os.Getenv("PGPASSWORD")),
			Name:     getEnv("DB_NAME", "railway"),
			SSLMode:  getEnv("DB_SSLMODE", "require"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
