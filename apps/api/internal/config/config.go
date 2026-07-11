package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL     string
	Port            string
	JWTSecret       string
	VAPIDPublicKey  string
	VAPIDPrivateKey string
	PushContact     string
}

func Load() (*Config, error) {
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../.env")
	_ = godotenv.Load("../../.env")

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}
	// Web Push is optional: if unset, push.Sender silently no-ops rather than
	// failing startup, so local dev without VAPID keys still works.
	return &Config{
		DatabaseURL:     dbURL,
		Port:            port,
		JWTSecret:       jwtSecret,
		VAPIDPublicKey:  os.Getenv("VAPID_PUBLIC_KEY"),
		VAPIDPrivateKey: os.Getenv("VAPID_PRIVATE_KEY"),
		PushContact:     os.Getenv("PUSH_CONTACT"),
	}, nil
}
