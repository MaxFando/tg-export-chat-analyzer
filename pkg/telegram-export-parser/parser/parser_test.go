package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseSingleFile(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "result.json")
	jsonData := `{
		"name": "Test Chat",
		"type": "personal_chat",
		"id": 100,
		"messages": [
			{
				"id": 1,
				"type": "message",
				"date": "2026-01-01T00:00:00",
				"from_id": "user123456",
				"from": "User One",
				"text": "Hello @user",
				"text_entities": [
					{
						"type": "mention",
						"text": "@user",
						"offset": 6,
						"length": 5
					}
				]
			}
		]
	}`
	if err := os.WriteFile(filePath, []byte(jsonData), 0644); err != nil {
		t.Fatal(err)
	}

	events, err := ParseSingleFile(filePath)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
	if events[0].ID != 1 || events[0].FromID != "user123456" {
		t.Errorf("Event data mismatch: got ID=%d, FromID=%s", events[0].ID, events[0].FromID)
	}
	if events[0].Text != "Hello @user" {
		t.Errorf("Expected 'Hello @user', got '%s'", events[0].Text)
	}
}

func TestParseFile(t *testing.T) {
	jsonData := `{
		"name": "Test Chat",
		"type": "personal_chat", 
		"id": 100,
		"messages": [
			{
				"id": 1,
				"type": "message",
				"date": "2026-01-01T00:00:00",
				"from_id": "user123456",
				"from": "User Two",
				"text": "Hello"
			}
		]
	}`

	reader := strings.NewReader(jsonData)
	events, err := ParseFile(reader, "test.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
}

func TestMergeFromFolders(t *testing.T) {
	tempDir := t.TempDir()
	folder1 := filepath.Join(tempDir, "export1")
	folder2 := filepath.Join(tempDir, "export2")
	os.MkdirAll(folder1, 0755)
	os.MkdirAll(folder2, 0755)

	jsonData1 := `{
		"name": "Test Chat",
		"type": "personal_chat",
		"id": 100,
		"messages": [
			{
				"id": 1,
				"type": "message",
				"date": "2026-01-01T00:00:00",
				"from_id": "user123",
				"from": "User 1",
				"text": "Msg1"
			}
		]
	}`
	jsonData2 := `{
		"name": "Test Chat",
		"type": "personal_chat", 
		"id": 100,
		"messages": [
			{
				"id": 2,
				"type": "message",
				"date": "2026-01-02T00:00:00",
				"from_id": "user456",
				"from": "User 2",
				"text": "Msg2"
			}
		]
	}`

	os.WriteFile(filepath.Join(folder1, "result.json"), []byte(jsonData1), 0644)
	os.WriteFile(filepath.Join(folder2, "result.json"), []byte(jsonData2), 0644)

	merged, err := MergeFromFolders([]string{folder1, folder2})
	if err != nil {
		t.Fatal(err)
	}
	if len(merged.Events) != 2 {
		t.Errorf("Expected 2 events, got %d", len(merged.Events))
	}
	if !merged.Events[0].Date.Before(merged.Events[1].Date) {
		t.Errorf("Events not sorted by date")
	}
}

func TestMergeEvents(t *testing.T) {
	events1 := []Event{
		{ID: 1, ChatID: 100, Text: "First"},
		{ID: 2, ChatID: 100, Text: "Second"},
	}
	events2 := []Event{
		{ID: 1, ChatID: 100, Text: "First"}, // Duplicate
		{ID: 3, ChatID: 100, Text: "Third"},
	}

	merged := MergeEvents([][]Event{events1, events2})
	if len(merged) != 3 {
		t.Errorf("Expected 3 unique events, got %d", len(merged))
	}
}
