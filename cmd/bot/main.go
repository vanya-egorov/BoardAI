package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"BoardAI/internal/bot"
	"BoardAI/internal/config"
	"BoardAI/internal/llm"
	"BoardAI/internal/orchestrator"
	"BoardAI/internal/repository"
)

func main() {
	http.DefaultClient.Timeout = 20 * time.Minute
	http.DefaultTransport.(*http.Transport).ResponseHeaderTimeout = 20 * time.Minute
	http.DefaultTransport.(*http.Transport).IdleConnTimeout = 20 * time.Minute
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := repository.NewPostgresDB(cfg.DBURL)
	if err != nil {
		log.Fatalf("failed to init postgres: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("error closing db: %v", err)
		}
	}()

	repo := repository.NewAnalysisRepository(db)

	llmClient := llm.NewClient(cfg.OllamaBaseURL, cfg.OllamaAPIToken)

	orc := orchestrator.NewOrchestrator(llmClient, cfg)

	tgBot, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	if err != nil {
		log.Fatalf("failed to init telegram bot: %v", err)
	}
	tgBot.Debug = false

	log.Printf("Authorized on account %s", tgBot.Self.UserName)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		log.Printf("received signal: %s, shutting down...", sig)
		cancel()
		time.Sleep(2 * time.Second)
	}()

	stateManager := bot.NewStateManager()
	handler := bot.NewHandler(tgBot, repo, orc, stateManager)

	if err := handler.Run(ctx); err != nil {
		log.Fatalf("bot stopped with error: %v", err)
	}
}
