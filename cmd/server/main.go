package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/aqws11223344/dark-badmintonteam/internal/config"
	linehandler "github.com/aqws11223344/dark-badmintonteam/internal/line"
	"github.com/aqws11223344/dark-badmintonteam/internal/store"
	"github.com/aqws11223344/dark-badmintonteam/internal/store/dual"
	"github.com/aqws11223344/dark-badmintonteam/internal/store/sheets"
	"github.com/aqws11223344/dark-badmintonteam/internal/store/turso"
)

func main() {
	cfg := config.Load()

	st := buildStore(cfg)

	handler, err := linehandler.New(linehandler.Config{
		ChannelSecret:  cfg.ChannelSecret,
		ChannelToken:   cfg.ChannelToken,
		LIFFID:         cfg.LIFFID,
		AdminUserIDs:   cfg.AdminUserIDs,
		BootstrapToken: cfg.BootstrapToken,
		Store:          st,
	})
	if err != nil {
		log.Fatalf("init line handler: %v", err)
	}

	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	r.POST("/webhook", gin.WrapF(handler.ServeHTTP))
	r.POST("/api/results", handler.SubmitResult)
	r.GET("/api/options", handler.GetOptions)
	r.Static("/liff", "./web/liff")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{Addr: ":" + port, Handler: r}

	go func() {
		log.Printf("listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

func buildStore(cfg config.Config) store.Store {
	ctx := context.Background()

	var sh store.Store
	if cfg.SheetsID != "" {
		s, err := sheets.New(ctx, cfg.SheetsID, cfg.GoogleCredentialsFile)
		if err != nil {
			log.Printf("sheets disabled: %v", err)
		} else {
			sh = s
			log.Println("sheets store: enabled")
		}
	}

	var ts store.Store
	if cfg.TursoURL != "" {
		t, err := turso.New(cfg.TursoURL, cfg.TursoToken)
		if err != nil {
			log.Printf("turso disabled: %v", err)
		} else {
			ts = t
			log.Println("turso store: enabled")
		}
	}

	switch {
	case sh != nil && ts != nil:
		return dual.New(ts, sh) // turso = primary (read), sheets = mirror
	case ts != nil:
		return ts
	case sh != nil:
		return sh
	default:
		log.Println("no store configured — using in-memory (data lost on restart)")
		return store.NewMemory()
	}
}
