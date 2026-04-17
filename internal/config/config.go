package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ChannelSecret         string
	ChannelToken          string
	LIFFID                string
	SheetsID              string
	GoogleCredentialsFile string
	TursoURL              string
	TursoToken            string
	PublicBaseURL         string
}

func Load() Config {
	// 本機開發用：有 .env 就載入，沒有就跳過（線上部署靠 Cloud Run 環境變數）。
	if err := godotenv.Load(); err != nil {
		log.Printf(".env not loaded (%v) — using OS environment", err)
	}

	return Config{
		ChannelSecret:         required("LINE_CHANNEL_SECRET"),
		ChannelToken:          required("LINE_CHANNEL_TOKEN"),
		LIFFID:                os.Getenv("LIFF_ID"),
		SheetsID:              os.Getenv("GOOGLE_SHEETS_ID"),
		GoogleCredentialsFile: os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"),
		TursoURL:              os.Getenv("TURSO_DATABASE_URL"),
		TursoToken:            os.Getenv("TURSO_AUTH_TOKEN"),
		PublicBaseURL:         os.Getenv("PUBLIC_BASE_URL"),
	}
}

func required(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("env %s is required", key)
	}
	return v
}
