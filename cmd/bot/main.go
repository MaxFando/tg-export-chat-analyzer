package main

import (
	"log"
	"os"
	"strconv"

	"github.com/MaxFando/tg-export-chat-analyzer/internal/delivery/telegram"
)

func main() {
	// Получаем конфигурацию из переменных окружения
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable not set")
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	tempDir := os.Getenv("TEMP_DIR")
	if tempDir == "" {
		tempDir = "/tmp/telegram-bot"
	}

	// Парсим целочисленные параметры
	maxFiles := 10
	if maxFilesStr := os.Getenv("MAX_FILES"); maxFilesStr != "" {
		if v, err := strconv.Atoi(maxFilesStr); err == nil {
			maxFiles = v
		}
	}

	maxFileSizeMB := 10
	if maxFileSizeStr := os.Getenv("MAX_FILE_SIZE_MB"); maxFileSizeStr != "" {
		if v, err := strconv.Atoi(maxFileSizeStr); err == nil {
			maxFileSizeMB = v
		}
	}

	maxTotalSizeMB := 100
	if maxTotalStr := os.Getenv("MAX_TOTAL_SIZE_MB"); maxTotalStr != "" {
		if v, err := strconv.Atoi(maxTotalStr); err == nil {
			maxTotalSizeMB = v
		}
	}

	sessionTimeoutMin := 60
	if timeoutStr := os.Getenv("SESSION_TIMEOUT_MINUTES"); timeoutStr != "" {
		if v, err := strconv.Atoi(timeoutStr); err == nil {
			sessionTimeoutMin = v
		}
	}

	// Создаём бота
	cfg := telegram.Config{
		Token:             token,
		MaxFiles:          maxFiles,
		MaxFileSizeMB:     maxFileSizeMB,
		MaxTotalSizeMB:    maxTotalSizeMB,
		SessionTimeoutMin: sessionTimeoutMin,
		LogLevel:          logLevel,
		TempDir:           tempDir,
	}

	bot, err := telegram.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	// Запускаем бота
	if err := bot.Start(); err != nil {
		log.Fatalf("Failed to start bot: %v", err)
	}
}
