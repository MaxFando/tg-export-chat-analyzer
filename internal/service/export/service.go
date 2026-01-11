package export

import (
	"time"

	"github.com/lintenved/tg-exporter/exporter"
)

// Service управляет экспортом результатов
type Service struct {
	// поля для будущего расширения конфигурации
}

// New создаёт новый ExportService
func New() *Service {
	return &Service{}
}

// ChooseFormat выбирает формат вывода на основе количества участников
func (s *Service) ChooseFormat(participantsCount int) exporter.OutputFormat {
	return exporter.ChooseFormat(participantsCount)
}

// ExportToExcel экспортирует результат в Excel
func (s *Service) ExportToExcel(result exporter.ParticipantsResult) ([]byte, error) {
	return exporter.ExportExcel(result, exporter.Options{
		ExportedAt: time.Now(),
	})
}

// FormatForTelegram форматирует список участников для Telegram
func (s *Service) FormatForTelegram(participants []exporter.Participant) []string {
	return exporter.FormatTelegramUserList(participants, exporter.DefaultMaxMessageLen)
}

// Export выполняет экспорт в выбранный формат
func (s *Service) Export(result exporter.ParticipantsResult) (interface{}, error) {
	format := s.ChooseFormat(len(result.Participants))

	switch format {
	case exporter.OutputExcel:
		return s.ExportToExcel(result)
	case exporter.OutputTelegramList:
		return s.FormatForTelegram(result.Participants), nil
	default:
		return s.FormatForTelegram(result.Participants), nil
	}
}
