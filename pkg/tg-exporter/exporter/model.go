package exporter

import "time"

type Participant struct {
	ID           string
	Username     string
	FirstName    string
	LastName     string
	Bio          string
	RegisteredAt *time.Time
	HasChannel   bool
	IsDeleted    bool
}

type ParticipantsResult struct {
	Participants []Participant
	Mentions     []Participant
	Channels     []string
}
