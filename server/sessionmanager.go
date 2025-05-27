package main

import (
	"errors"
	"path/filepath"
	"strings"
	"sync"
)

type SessionManager struct {
	sessions  map[string]*Session
	rootDir   string       // Абсолютный путь (например, "C:\For_network_technologies\Working_directory")
	busyPaths sync.Map     // Занятые пути (одна из сессий запустила обновления)
	mutex     sync.RWMutex // Для потокобезопасного доступа к сессиям
}

func NewSessionManager(rootDir string) *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
		rootDir:  filepath.Clean(rootDir),
	}
}

// Возвращает абсолютный путь к userpath.
func (sm *SessionManager) ResolvePath(session *Session, userPath string) (string, error) {
	cleanPath := filepath.Clean(filepath.Join(session.CurrentDir, userPath))
	absPath := filepath.Join(sm.rootDir, cleanPath)

	// Проверка выхода за пределы корня
	if !strings.HasPrefix(absPath, sm.rootDir) {
		return "", errors.New("выход за пределы разрешённой директории")
	}
	return absPath, nil
}

// LockPath блокирует папку для операций
func (sm *SessionManager) LockPath(path string) bool {
	_, loaded := sm.busyPaths.LoadOrStore(path, true)
	return !loaded // !loaded = false => путь уже занят
}

func (sm *SessionManager) UnlockPath(path string) {
	sm.busyPaths.Delete(path)
}
