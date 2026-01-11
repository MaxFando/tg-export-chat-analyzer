package telegram

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/MaxFando/tg-export-chat-analyzer/internal/service/export"
	"github.com/MaxFando/tg-export-chat-analyzer/internal/service/participant"
	"github.com/MaxFando/tg-export-chat-analyzer/internal/service/session"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/inqast/fstorage/storage"

	"github.com/MaxFando/tg-export-chat-analyzer/pkg/logger"

	"github.com/Nikalively/telegram-export-parser/parser"
	"github.com/lintenved/tg-exporter/exporter"
)

// Bot управляет Telegram ботом
type Bot struct {
	api               *tgbotapi.BotAPI
	sessionManager    *session.Manager
	tempStorage       storage.TempStorage
	participantSvc    *participant.ParticipantExtractor
	exportSvc         *export.Service
	logger            *logger.Logger
	maxFiles          int
	maxFileSizeMB     int
	maxTotalSizeMB    int
	sessionTimeoutMin int
}

// Config конфигурация для бота
type Config struct {
	Token             string
	MaxFiles          int
	MaxFileSizeMB     int
	MaxTotalSizeMB    int
	SessionTimeoutMin int
	LogLevel          string
	TempDir           string
}

// New создаёт новый бот
func New(cfg Config) (*Bot, error) {
	// Инициализируем API Telegram
	api, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot API: %w", err)
	}

	// Инициализируем логгер
	log := logger.New(cfg.LogLevel)

	// Инициализируем временное хранилище
	tmpStorage, err := storage.NewFileSystemStorage(cfg.TempDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp storage: %w", err)
	}

	// Создаём сессионный менеджер
	sessionMgr := session.NewManager(time.Duration(cfg.SessionTimeoutMin) * time.Minute)

	// Создаём сервисы
	partSvc := participant.New()
	expSvc := export.New()

	bot := &Bot{
		api:               api,
		sessionManager:    sessionMgr,
		tempStorage:       tmpStorage,
		participantSvc:    partSvc,
		exportSvc:         expSvc,
		logger:            log,
		maxFiles:          cfg.MaxFiles,
		maxFileSizeMB:     cfg.MaxFileSizeMB,
		maxTotalSizeMB:    cfg.MaxTotalSizeMB,
		sessionTimeoutMin: cfg.SessionTimeoutMin,
	}

	log.Info("bot initialized", "botname", api.Self.UserName)
	return bot, nil
}

// Start запускает бота и обрабатывает обновления
func (b *Bot) Start() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	b.logger.Info("bot started, listening for updates")

	for update := range updates {
		// Обрабатываем обновление в отдельной горутине
		go b.handleUpdate(update)
	}

	return nil
}

// handleUpdate обрабатывает одно обновление от Telegram
func (b *Bot) handleUpdate(update tgbotapi.Update) {
	defer func() {
		if r := recover(); r != nil {
			b.logger.Error("recovered from panic", "error", r)
		}
	}()

	// Обработка текстовых сообщений
	if update.Message != nil {
		b.handleMessage(update.Message)
	}
}

// handleMessage обрабатывает текстовое сообщение
func (b *Bot) handleMessage(msg *tgbotapi.Message) {
	userID := int64(msg.From.ID)
	chatID := msg.Chat.ID

	b.logger.Debug("received message", "userID", userID, "text", msg.Text)

	// Обработка команд
	if msg.IsCommand() {
		b.handleCommand(userID, chatID, msg.Command())
		return
	}

	// Обработка файлов
	if msg.Document != nil {
		b.handleFile(userID, chatID, msg.Document)
		return
	}

	// Если сообщение не команда и не файл, игнорируем
}

// handleCommand обрабатывает команды бота
func (b *Bot) handleCommand(userID, chatID int64, command string) {
	switch command {
	case "start":
		b.cmdStart(userID, chatID)
	case "help":
		b.cmdHelp(chatID)
	case "upload":
		b.cmdUpload(chatID)
	case "process":
		b.cmdProcess(userID, chatID)
	case "cancel":
		b.cmdCancel(userID, chatID)
	default:
		b.sendMessage(chatID, "Unknown command. Use /help for available commands.")
	}
}

// handleFile обрабатывает загруженный файл
func (b *Bot) handleFile(userID, chatID int64, doc *tgbotapi.Document) {
	// Проверяем лимит файлов
	sess := b.sessionManager.GetOrCreate(userID)
	if len(sess.Files) >= b.maxFiles {
		b.sendMessage(chatID, fmt.Sprintf(MessageFileLimitExceeded, len(sess.Files)))
		return
	}

	// Проверяем размер файла
	fileSizeMB := float64(doc.FileSize) / (1024 * 1024)
	if fileSizeMB > float64(b.maxFileSizeMB) {
		b.sendMessage(chatID, fmt.Sprintf(MessageFileSizeExceeded, fileSizeMB))
		return
	}

	// Скачиваем файл
	fileURL, err := b.api.GetFileDirectURL(doc.FileID)
	if err != nil {
		b.logger.Error("failed to get file URL", "error", err)
		b.sendMessage(chatID, MessageUnexpectedError)
		return
	}

	// Загружаем файл
	resp, err := http.Get(fileURL)
	if err != nil {
		b.logger.Error("failed to download file", "error", err)
		b.sendMessage(chatID, MessageUnexpectedError)
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			b.logger.Error("failed to close response body", "error", err)
		}
	}()

	// Сохраняем в временное хранилище
	filename := doc.FileName
	if filename == "" {
		filename = fmt.Sprintf("export_%d.json", userID)
	}

	filePath, err := b.tempStorage.Save(filename, resp.Body)
	if err != nil {
		b.logger.Error("failed to save temp file", "error", err)
		b.sendMessage(chatID, MessageUnexpectedError)
		return
	}

	// Добавляем в сессию
	b.sessionManager.AddFile(userID, filePath)
	b.sendMessage(chatID, fmt.Sprintf(MessageFileReceived, filename))
	b.sendMessage(chatID, fmt.Sprintf(MessageFilesReady, len(sess.Files), fileSizeMB))
}

// cmdStart обрабатывает команду /start
func (b *Bot) cmdStart(userID, chatID int64) {
	sess := b.sessionManager.Get(userID)
	if sess != nil && len(sess.Files) > 0 {
		b.sendMessage(chatID, MessageWelcomeBack)
	} else {
		b.sendMessage(chatID, MessageStart)
	}
}

// cmdHelp обрабатывает команду /help
func (b *Bot) cmdHelp(chatID int64) {
	b.sendMessage(chatID, MessageHelp)
}

// cmdUpload обрабатывает команду /upload
func (b *Bot) cmdUpload(chatID int64) {
	b.sendMessage(chatID, MessageUploadPrompt)
}

// cmdProcess обрабатывает команду /process
func (b *Bot) cmdProcess(userID, chatID int64) {
	sess := b.sessionManager.Get(userID)
	if sess == nil || len(sess.Files) == 0 {
		b.sendMessage(chatID, MessageNoFiles)
		return
	}

	b.sessionManager.SetState(userID, session.StateProcessing)
	b.sendMessage(chatID, fmt.Sprintf(MessageProcessing, len(sess.Files)))

	// Обрабатываем файлы
	b.processFiles(userID, chatID, sess.Files)

	// Очищаем сессию и удаляем временные файлы
	defer func() {
		b.tempStorage.DeleteAll(sess.Files)
		b.sessionManager.Clear(userID)
	}()
}

// cmdCancel обрабатывает команду /cancel
func (b *Bot) cmdCancel(userID, chatID int64) {
	sess := b.sessionManager.Get(userID)
	if sess == nil || len(sess.Files) == 0 {
		b.sendMessage(chatID, MessageNothingToCancel)
		return
	}

	// Удаляем все временные файлы
	b.tempStorage.DeleteAll(sess.Files)
	b.sessionManager.Clear(userID)

	b.sendMessage(chatID, MessageCancelled)
}

// processFiles обрабатывает загруженные файлы
func (b *Bot) processFiles(userID, chatID int64, filePaths []string) {
	var allEvents []parser.Event

	// Парсим все файлы
	for _, filePath := range filePaths {
		f, err := b.tempStorage.Read(filePath)
		if err != nil {
			b.logger.Error("failed to read file", "error", err)
			filename := filepath.Base(filePath)
			b.sendMessage(chatID, fmt.Sprintf(MessageFileParseError, filename, "Unable to read file"))
			return
		}
		defer func() {
			if err := f.Close(); err != nil {
				b.logger.Error("failed to close file", "error", err)
			}
		}()

		events, err := parser.ParseFile(f, filePath)
		if err != nil {
			b.logger.Error("failed to parse file", "error", err)
			filename := filepath.Base(filePath)
			b.sendMessage(chatID, fmt.Sprintf(MessageFileParseError, filename, err.Error()))
			return
		}

		allEvents = append(allEvents, events...)
	}

	// Объединяем события
	if len(allEvents) == 0 {
		b.sendMessage(chatID, MessageNoParticipants)
		return
	}

	mergedEvents := parser.MergeEvents([][]parser.Event{allEvents})

	// Извлекаем участников
	result, err := b.participantSvc.Extract(mergedEvents)
	if err != nil {
		b.logger.Error("failed to extract participants", "error", err)
		b.sendMessage(chatID, fmt.Sprintf(MessageProcessingError, err.Error()))
		return
	}

	if len(result.Participants) == 0 {
		b.sendMessage(chatID, MessageNoParticipants)
		return
	}

	// Отправляем статистику
	b.sendMessage(chatID, fmt.Sprintf(MessageResultReady,
		len(result.Participants),
		len(result.Mentions),
		len(result.Channels),
		len(mergedEvents)))

	// Выбираем формат и экспортируем
	format := b.exportSvc.ChooseFormat(len(result.Participants))

	switch format {
	case exporter.OutputExcel:
		b.sendExcelResult(chatID, result)
	case exporter.OutputTelegramList:
		b.sendListResult(chatID, result)
	}

	b.sessionManager.SetState(userID, session.StateComplete)
}

// sendListResult отправляет результат в виде списка в чат
func (b *Bot) sendListResult(chatID int64, result exporter.ParticipantsResult) {
	messages := b.exportSvc.FormatForTelegram(result.Participants)

	for i, msg := range messages {
		if len(messages) > 1 {
			b.sendMessage(chatID, fmt.Sprintf(MessageListTruncated, i+1, len(messages), msg))
		} else {
			b.sendMessage(chatID, fmt.Sprintf(MessageListReady, msg))
		}
	}
}

// sendExcelResult отправляет результат в виде Excel файла
func (b *Bot) sendExcelResult(chatID int64, result exporter.ParticipantsResult) {
	data, err := b.exportSvc.ExportToExcel(result)
	if err != nil {
		b.logger.Error("failed to export to Excel", "error", err)
		b.sendMessage(chatID, fmt.Sprintf(MessageProcessingError, err.Error()))
		return
	}

	// Отправляем файл
	fileBytes := tgbotapi.FileBytes{
		Name:  "export.xlsx",
		Bytes: data,
	}

	msg := tgbotapi.NewDocument(chatID, fileBytes)
	msg.Caption = MessageExcelReady

	if _, err := b.api.Send(msg); err != nil {
		b.logger.Error("failed to send Excel file", "error", err)
		b.sendMessage(chatID, MessageUnexpectedError)
	}
}

// sendMessage отправляет текстовое сообщение
func (b *Bot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"

	if _, err := b.api.Send(msg); err != nil {
		b.logger.Error("failed to send message", "error", err)
	}
}
