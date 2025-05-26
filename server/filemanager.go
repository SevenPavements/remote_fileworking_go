package main

import (
	"errors"
	"path/filepath"
	"strings"
)

func (sm *SessionManager) ResolvePath(session *Session, userPath string) (string, error) {

	cleanPath := filepath.Clean(filepath.Join(session.CurrentDir, userPath))
	absPath := filepath.Join(sm.rootDir, cleanPath)

	relPath, err := filepath.Rel(sm.rootDir, absPath)
	if err != nil || strings.HasPrefix(relPath, "..") {
		return "", errors.New("неверный путь")
	}
	return absPath, nil
}

func (sm *SessionManager) LockPath(path string) bool {
	_, loaded := sm.busyPaths.LoadOrStore(path, true)
	return !loaded
}

func (sm *SessionManager) UnlockPath(path string) {
	sm.busyPaths.Delete(path)
}
