package config

import (
	"log"
	"os"
)

type Config struct {
	DatabaseURL string
}

func Load() Config {
	dbURL := os.Getenv("DATABASE_URL")

	log.Printf("DATABASE_URL=%q\n", dbURL)

	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	return Config{
		DatabaseURL: dbURL,
	}
}
