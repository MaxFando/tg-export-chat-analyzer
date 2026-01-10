package exporter

import (
	"strings"
)

const DefaultMaxMessageLen = 3500

func FormatTelegramUserList(participants []Participant, maxLen int) []string {
	if maxLen <= 0 {
		maxLen = DefaultMaxMessageLen
	}

	lines := make([]string, 0, len(participants))
	for _, p := range participants {
		u := strings.TrimSpace(p.Username)
		if u == "" {
			continue
		}
		if !strings.HasPrefix(u, "@") {
			u = "@" + u
		}
		lines = append(lines, u)
	}

	if len(lines) == 0 {
		return []string{"(нет участников с username)"}
	}

	var out []string
	var b strings.Builder

	flush := func() {
		if b.Len() == 0 {
			return
		}
		out = append(out, b.String())
		b.Reset()
	}

	for _, line := range lines {
		// +1 for '\n' if not first line
		addLen := len(line)
		if b.Len() > 0 {
			addLen += 1
		}
		// if line itself bigger than limit, we still send it alone
		if b.Len() > 0 && b.Len()+addLen > maxLen {
			flush()
		}
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(line)

		// If single line already exceeds maxLen, flush immediately
		if b.Len() >= maxLen {
			flush()
		}
	}

	flush()
	return out
}
