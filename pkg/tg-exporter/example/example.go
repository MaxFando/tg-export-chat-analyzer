package main

import (
	"github.com/lintenved/tg-exporter/exporter"
	"os"
	"time"
)

func main() {
	now := time.Now()
	reg := now.AddDate(-1, 0, 0)

	res := exporter.ParticipantsResult{
		Participants: []exporter.Participant{
			{Username: "alice", FirstName: "Alice", LastName: "A", Bio: "hi", RegisteredAt: &reg, HasChannel: true},
			{Username: "@bob", FirstName: "Bob", LastName: "B", Bio: "yo", HasChannel: false},
		},
		Mentions: []exporter.Participant{
			{Username: "charlie"},
		},
		Channels: []string{"@my_channel"},
	}

	b, err := exporter.ExportExcel(res, exporter.Options{ExportedAt: now})
	if err != nil {
		panic(err)
	}

	_ = os.MkdirAll("examples", 0o755)
	if err = os.WriteFile("examples/export_example.xlsx", b, 0o644); err != nil {
		panic(err)
	}
}
