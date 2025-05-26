package main

import (
	"crypto/rand"
	"encoding/hex"
	"net"
	"sync"
	"time"
)

type Session struct {
	ID         string
	ClientAddr string
	CreatedAt  time.Time
}

type SessionManager struct {
	sessions map[string]*Session
	mutex    sync.RWMutex
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
	}
}

func (sm *SessionManager) CreateSession(conn net.Conn) string {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	token := generateUUID()
	sm.sessions[token] = &Session{
		ID:         token,
		ClientAddr: conn.RemoteAddr().String(),
		CreatedAt:  time.Now(),
	}
	return token
}

func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
