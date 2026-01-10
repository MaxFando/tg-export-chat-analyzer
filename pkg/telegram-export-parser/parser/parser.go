package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ParseFile parses a single Telegram export file from io.Reader and returns events.
func ParseFile(r io.Reader, filename string) ([]Event, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Determine file type by extension
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".json":
		return parseJSON(data)
	case ".html":
		// TODO: HTML parsing can be added later if needed
		return nil, fmt.Errorf("HTML parsing not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}
}

// ParseSingleFile parses a single result.json file and returns events.
func ParseSingleFile(filePath string) ([]Event, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return ParseFile(file, filePath)
}

// parseJSON parses JSON data and returns events
func parseJSON(data []byte) ([]Event, error) {
	var export TelegramExport
	if err := json.Unmarshal(data, &export); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	var events []Event
	for _, msg := range export.Messages {
		// Skip service messages and non-message types
		if msg.Type != "message" {
			continue
		}

		// Parse date
		date, err := time.Parse("2006-01-02T15:04:05", msg.Date)
		if err != nil {
			// Try alternative format
			date, err = time.Parse(time.RFC3339, msg.Date)
			if err != nil {
				continue // Skip messages with invalid dates
			}
		}

		// Extract text from various formats
		text := extractText(msg.Text)

		event := Event{
			ID:       msg.ID,
			FromID:   msg.FromID,
			Text:     text,
			Date:     date,
			ChatID:   export.ID,
			Entities: msg.Entities,
		}
		events = append(events, event)
	}
	return events, nil
}

// extractText extracts text from various Telegram text formats
func extractText(textField interface{}) string {
	if textField == nil {
		return ""
	}

	switch v := textField.(type) {
	case string:
		return v
	case []interface{}:
		var result strings.Builder
		for _, item := range v {
			switch itemVal := item.(type) {
			case string:
				result.WriteString(itemVal)
			case map[string]interface{}:
				if text, ok := itemVal["text"].(string); ok {
					result.WriteString(text)
				}
			}
		}
		return result.String()
	default:
		return ""
	}
}

// MergeEvents merges multiple event slices, sorts by date, and deduplicates by ID.
func MergeEvents(eventSlices [][]Event) []Event {
	eventMap := make(map[int64]Event)

	for _, events := range eventSlices {
		for _, event := range events {
			// Use combination of ID and ChatID as unique key for better deduplication
			key := event.ID + event.ChatID*1000000
			eventMap[key] = event
		}
	}

	events := make([]Event, 0, len(eventMap))
	for _, event := range eventMap {
		events = append(events, event)
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].Date.Before(events[j].Date)
	})

	return events
}

// MergeFromFolders merges events from multiple export folders, sorts by date, and deduplicates by ID.
func MergeFromFolders(folderPaths []string) (*MergedEvents, error) {
	var eventSlices [][]Event

	for _, folder := range folderPaths {
		files, err := filepath.Glob(filepath.Join(folder, "result.json"))
		if err != nil {
			return nil, fmt.Errorf("failed to glob files in %s: %w", folder, err)
		}

		for _, file := range files {
			events, err := ParseSingleFile(file)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %s: %w", file, err)
			}
			eventSlices = append(eventSlices, events)
		}
	}

	merged := MergeEvents(eventSlices)
	return &MergedEvents{Events: merged}, nil
}
