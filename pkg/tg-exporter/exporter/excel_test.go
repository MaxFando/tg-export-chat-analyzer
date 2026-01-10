package exporter_test

import (
	"github.com/lintenved/tg-exporter/exporter"
	"testing"
	"time"
)

func TestExportExcel_Basic(t *testing.T) {
	now := time.Now()

	res := exporter.ParticipantsResult{
		Participants: []exporter.Participant{
			{Username: "alice", FirstName: "Alice"},
			{Username: "@bob", FirstName: "Bob"},
		},
		Mentions: []exporter.Participant{{Username: "charlie"}},
		Channels: []string{"@ch1", "@ch2"},
	}

	b, err := exporter.ExportExcel(res, exporter.Options{ExportedAt: now})
	if err != nil {
		t.Fatalf("ExportExcel error: %v", err)
	}
	if len(b) < 1000 {
		t.Fatalf("expected non-trivial xlsx size, got %d", len(b))
	}
	if err = exporter.ValidateExcelBytes(b); err != nil {
		t.Fatalf("ValidateExcelBytes: %v", err)
	}
}
