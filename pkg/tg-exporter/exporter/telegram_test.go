package exporter_test

import (
	"github.com/lintenved/tg-exporter/exporter"
	"strings"
	"testing"
)

func TestFormatTelegramUserList_Splits(t *testing.T) {
	parts := make([]exporter.Participant, 0, 500)
	for i := 0; i < 500; i++ {
		parts = append(parts, exporter.Participant{Username: "user" + strings.Repeat("x", 10)})
	}

	msgs := exporter.FormatTelegramUserList(parts, 200)
	if len(msgs) < 2 {
		t.Fatalf("expected split into multiple messages, got %d", len(msgs))
	}
	for _, m := range msgs {
		if len(m) > 200 {
			t.Fatalf("message exceeds limit: %d", len(m))
		}
	}
}
