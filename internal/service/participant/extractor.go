package participant

import (
	"regexp"
	"strings"

	"github.com/Nikalively/telegram-export-parser/parser"
	"github.com/lintenved/tg-exporter/exporter"
)

// Extractor интерфейс для извлечения участников из событий
type Extractor interface {
	Extract(events []parser.Event) (exporter.ParticipantsResult, error)
}

// ParticipantExtractor реализует интерфейс Extractor
type ParticipantExtractor struct {
	// поля для будущего расширения конфигурации
}

// New создаёт новый экстрактор участников
func New() *ParticipantExtractor {
	return &ParticipantExtractor{}
}

// Extract извлекает участников и упоминания из событий
func (pe *ParticipantExtractor) Extract(events []parser.Event) (exporter.ParticipantsResult, error) {
	// Карты для дедупликации
	participantMap := make(map[string]*exporter.Participant)
	mentionMap := make(map[string]*exporter.Participant)
	channelSet := make(map[string]bool)

	// Обрабатываем каждое событие
	for _, event := range events {
		// Добавляем автора как участника
		if event.FromID != "" && !strings.HasPrefix(event.FromID, "channel") {
			key := strings.ToLower(event.FromID)
			if _, exists := participantMap[key]; !exists {
				participantMap[key] = &exporter.Participant{
					ID:        event.FromID,
					Username:  extractUsername(event.FromID),
					IsDeleted: false,
				}
			}
		}

		// Извлекаем упоминания из текста сообщения
		mentions := extractMentions(event.Text)
		for _, mention := range mentions {
			key := strings.ToLower(mention)
			if _, exists := mentionMap[key]; !exists {
				mentionMap[key] = &exporter.Participant{
					ID:        mention,
					Username:  mention,
					IsDeleted: false,
				}
			}
		}

		// Извлекаем упоминания из entities
		for _, entity := range event.Entities {
			if entity.Type == "mention" {
				mention := strings.TrimLeft(entity.Text, "@")
				if mention != "" {
					key := strings.ToLower(mention)
					if _, exists := mentionMap[key]; !exists {
						mentionMap[key] = &exporter.Participant{
							ID:        mention,
							Username:  mention,
							IsDeleted: false,
						}
					}
				}
			}

			// Извлекаем каналы
			if entity.Type == "channel" {
				channelSet[entity.Text] = true
			}
		}
	}

	// Фильтруем удалённые аккаунты и пустые значения
	participants := filterParticipants(participantMap)
	mentions := filterParticipants(mentionMap)

	// Преобразуем set каналов в slice
	channels := make([]string, 0, len(channelSet))
	for ch := range channelSet {
		channels = append(channels, ch)
	}

	return exporter.ParticipantsResult{
		Participants: participants,
		Mentions:     mentions,
		Channels:     channels,
	}, nil
}

// extractUsername извлекает username из ID или текста
func extractUsername(id string) string {
	// Если это уже username с @, убираем его
	if strings.HasPrefix(id, "@") {
		return strings.TrimPrefix(id, "@")
	}

	// Если это числовой ID, пропускаем
	if isNumeric(id) {
		return ""
	}

	return id
}

// extractMentions находит все @username в тексте сообщения
func extractMentions(text string) []string {
	if text == "" {
		return nil
	}

	// Регулярное выражение для поиска @mention
	// Поддерживает: буквы, цифры, underscore, dash
	re := regexp.MustCompile(`@([a-zA-Z0-9_]+)`)
	matches := re.FindAllString(text, -1)

	result := make([]string, 0, len(matches))
	seen := make(map[string]bool)

	for _, match := range matches {
		if !seen[match] {
			result = append(result, match)
			seen[match] = true
		}
	}

	return result
}

// filterParticipants удаляет дубли и удалённые аккаунты
func filterParticipants(pMap map[string]*exporter.Participant) []exporter.Participant {
	result := make([]exporter.Participant, 0, len(pMap))

	for _, p := range pMap {
		// Пропускаем удалённые аккаунты
		if p.IsDeleted {
			continue
		}

		// Пропускаем пустые username и имена
		if strings.TrimSpace(p.Username) == "" &&
			strings.TrimSpace(p.FirstName) == "" &&
			strings.TrimSpace(p.LastName) == "" {
			continue
		}

		result = append(result, *p)
	}

	return result
}

// isNumeric проверяет, является ли строка числовым ID
func isNumeric(s string) bool {
	if s == "" {
		return false
	}

	for _, c := range s {
		if c < '0' || c > '9' {
			if c == '-' && s[0] == '-' {
				continue // Разрешаем знак минус в начале
			}
			return false
		}
	}

	return true
}
