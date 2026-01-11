package session

import (
	"sync"
	"time"
)

// State представляет состояние сессии
type State string

const (
	StateEmpty      State = "empty"
	StateLoading    State = "loading"
	StateProcessing State = "processing"
	StateComplete   State = "complete"
)

// Session хранит информацию о сессии пользователя
type Session struct {
	UserID    int64
	State     State
	Files     []string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Manager управляет сессиями пользователей in-memory
type Manager struct {
	mu       sync.RWMutex
	sessions map[int64]*Session
	timeout  time.Duration
}

// NewManager создаёт новый Manager с timeout для очистки сессий
func NewManager(timeout time.Duration) *Manager {
	sm := &Manager{
		sessions: make(map[int64]*Session),
		timeout:  timeout,
	}

	// Запускаем горутину для очистки старых сессий
	go sm.cleanupExpired()

	return sm
}

// GetOrCreate получает существующую сессию или создаёт новую
func (sm *Manager) GetOrCreate(userID int64) *Session {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if session, exists := sm.sessions[userID]; exists {
		session.UpdatedAt = time.Now()
		return session
	}

	session := &Session{
		UserID:    userID,
		State:     StateEmpty,
		Files:     make([]string, 0),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	sm.sessions[userID] = session
	return session
}

// Get получает сессию по userID, или nil если не существует
func (sm *Manager) Get(userID int64) *Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.sessions[userID]
}

// AddFile добавляет файл в сессию и обновляет состояние
func (sm *Manager) AddFile(userID int64, filePath string) (*Session, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[userID]
	if !exists {
		session = &Session{
			UserID:    userID,
			State:     StateEmpty,
			Files:     make([]string, 0),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		sm.sessions[userID] = session
	}

	session.Files = append(session.Files, filePath)
	session.State = StateLoading
	session.UpdatedAt = time.Now()

	return session, nil
}

// GetFiles возвращает список файлов для пользователя
func (sm *Manager) GetFiles(userID int64) []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[userID]
	if !exists {
		return nil
	}

	// Возвращаем копию, чтобы избежать race conditions
	result := make([]string, len(session.Files))
	copy(result, session.Files)
	return result
}

// SetState обновляет состояние сессии
func (sm *Manager) SetState(userID int64, state State) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if session, exists := sm.sessions[userID]; exists {
		session.State = state
		session.UpdatedAt = time.Now()
	}
}

// Clear очищает сессию пользователя
func (sm *Manager) Clear(userID int64) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.sessions, userID)
}

// FileCount возвращает количество файлов в сессии
func (sm *Manager) FileCount(userID int64) int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[userID]
	if !exists {
		return 0
	}

	return len(session.Files)
}

// cleanupExpired запускается в отдельной горутине и очищает старые сессии
func (sm *Manager) cleanupExpired() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		sm.mu.Lock()

		now := time.Now()
		for userID, session := range sm.sessions {
			if now.Sub(session.UpdatedAt) > sm.timeout {
				delete(sm.sessions, userID)
			}
		}

		sm.mu.Unlock()
	}
}
