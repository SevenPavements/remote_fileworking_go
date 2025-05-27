package main

import (
	"crypto/rand"
	"encoding/hex"
	"net"
	"time"
)

type Session struct {
	ID           string
	ClientAddr   string
	CurrentDir   string // Относительный путь от корневой папки
	CreatedAt    time.Time
	LastActivity time.Time // Обновляется в handleConnection
}

func NewSession(conn net.Conn) *Session {
	return &Session{
		ID:           generateUUID(),
		ClientAddr:   conn.RemoteAddr().String(),
		CreatedAt:    time.Now(),
		CurrentDir:   "/", // Начальная директория
		LastActivity: time.Now(),
	}
}

func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
