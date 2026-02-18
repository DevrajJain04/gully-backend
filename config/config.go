package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoURI  string
	JWTSecret string
	Port      string
}

func Load() *Config {
	// best-effort .env load — ignore error if file missing
	_ = godotenv.Load()

	cfg := &Config{
		MongoURI:  os.Getenv("MONGO_URI"),
		JWTSecret: os.Getenv("JWT_SECRET"),
		Port:      os.Getenv("PORT"),
	}

	if cfg.MongoURI == "" {
		log.Fatal("MONGO_URI is required")
	}
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = "default-secret"
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	return cfg
}
