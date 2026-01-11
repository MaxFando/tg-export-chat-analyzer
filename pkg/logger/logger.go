package logger

import (
	"crypto/sha256"
	"fmt"
	"log"
	"strings"
)

// Logger предоставляет логирование без PII (личных данных)
type Logger struct {
	level Level
}

type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

func New(level string) *Logger {
	lvl := Level(strings.ToLower(level))
	switch lvl {
	case LevelDebug, LevelInfo, LevelWarn, LevelError:
		return &Logger{level: lvl}
	default:
		return &Logger{level: LevelInfo}
	}
}

func (l *Logger) Debug(msg string, keysAndValues ...interface{}) {
	if l.shouldLog(LevelDebug) {
		l.log("DEBUG", msg, keysAndValues...)
	}
}

func (l *Logger) Info(msg string, keysAndValues ...interface{}) {
	if l.shouldLog(LevelInfo) {
		l.log("INFO", msg, keysAndValues...)
	}
}

func (l *Logger) Warn(msg string, keysAndValues ...interface{}) {
	if l.shouldLog(LevelWarn) {
		l.log("WARN", msg, keysAndValues...)
	}
}

func (l *Logger) Error(msg string, keysAndValues ...interface{}) {
	if l.shouldLog(LevelError) {
		l.log("ERROR", msg, keysAndValues...)
	}
}

func (l *Logger) shouldLog(checkLevel Level) bool {
	levels := map[Level]int{
		LevelDebug: 0,
		LevelInfo:  1,
		LevelWarn:  2,
		LevelError: 3,
	}
	return levels[checkLevel] >= levels[l.level]
}

func (l *Logger) log(level, msg string, keysAndValues ...interface{}) {
	var parts []string
	parts = append(parts, fmt.Sprintf("[%s]", level))
	parts = append(parts, msg)

	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			key := fmt.Sprintf("%v", keysAndValues[i])
			val := fmt.Sprintf("%v", keysAndValues[i+1])
			// Не логируем чувствительные данные
			if isPrivateKey(key) {
				val = hashValue(val)
			}
			parts = append(parts, fmt.Sprintf("%s=%v", key, val))
		}
	}

	log.Println(strings.Join(parts, " "))
}

func isPrivateKey(key string) bool {
	private := map[string]bool{
		"username":  true,
		"userId":    true,
		"user_id":   true,
		"userID":    true,
		"firstName": true,
		"lastName":  true,
		"bio":       true,
		"file":      true,
		"path":      true,
		"token":     true,
		"password":  true,
		"email":     true,
		"phone":     true,
	}
	return private[strings.ToLower(key)]
}

func hashValue(val string) string {
	h := sha256.Sum256([]byte(val))
	return fmt.Sprintf("%x", h[:8])
}
