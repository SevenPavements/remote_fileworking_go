package main

import (
	"crypto/rand"
	"encoding/hex"
	"net"
	"path/filepath"
	"sync"
	"time"
)

type Session struct {
	ID           string
	ClientAddr   string
	CurrentDir   string // Относительный путь от корневой папки
	CreatedAt    time.Time
	LastActivity time.Time // Обновляется в handleConnection

}

func (s *Session) UpdateActivity() {
	s.LastActivity = time.Now()
}

type SessionManager struct {
	sessions  map[string]*Session
	rootDir   string
	busyPaths sync.Map
	mutex     sync.RWMutex
}

func NewSessionManager(rootDir string) *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
		rootDir:  filepath.Clean(rootDir),
	}
}

func (sm *SessionManager) CreateSession(conn net.Conn) *Session {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	token := generateUUID()
	session := &Session{
		ID:           token,
		ClientAddr:   conn.RemoteAddr().String(),
		CreatedAt:    time.Now(),
		CurrentDir:   "/", // Начинаем с корня
		LastActivity: time.Now(),
	}
	sm.sessions[token] = session
	return session
}

func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
