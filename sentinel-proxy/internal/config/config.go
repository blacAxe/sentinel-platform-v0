package config

import "os"

type Config struct {
	Port        string
	Target      string
	JWTSecret   string
}

func Load() *Config {
	return &Config{
		Port:      getEnv("PORT", "8081"),
		Target:    getEnv("TARGET_URL", "https://self-healing-security-lab.onrender.com"),
		JWTSecret: getEnv("JWT_SECRET", "dev_secret"),
	}
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}