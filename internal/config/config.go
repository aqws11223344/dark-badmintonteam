package config

import (
	"log"
	"os"
	"strings"

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
	AdminUserIDs          []string // LINE user IDs 有管理權限；空 = 所有人都是管理員
}

func Load() Config {
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
		AdminUserIDs:          parseList(os.Getenv("ADMIN_USER_IDS")),
	}
}

func parseList(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func required(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("env %s is required", key)
	}
	return v
}
