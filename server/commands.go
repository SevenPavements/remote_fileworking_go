package main

import (
	"net"
	"os"
	"path/filepath"
	"strings"
)

// Перейти в новую директорию
func handleCD(conn net.Conn, args []string, session *Session, sm *SessionManager) {

	// Неизвестная команда
	if len(args) < 1 {
		sendResponse(conn, 400, nil)
		return
	}

	// Запрещен доступ
	absPath, err := sm.ResolvePath(session, args[0])
	if err != nil {
		sendResponse(conn, 403, nil)
		return
	}

	// Не найдено такой папки/такого файла
	if info, err := os.Stat(absPath); err != nil || !info.IsDir() {
		sendResponse(conn, 404, nil)
		return
	}

	// Обновляем CurrentDir (относительно корня)
	relPath, _ := filepath.Rel(sm.rootDir, absPath)
	session.CurrentDir = filepath.Join("/", relPath)
	sendResponse(conn, 200, map[string]string{
		"current_dir": strings.TrimPrefix(absPath, sm.rootDir),
	})
}

// Вывести содержимое текущей директории
func handleLS(conn net.Conn, session *Session, sm *SessionManager) {
	absPath := filepath.Join(sm.rootDir, session.CurrentDir)
	files, err := os.ReadDir(absPath)
	// Ошибка чтения файла (внутренняя ошибка сервера)
	if err != nil {
		sendResponse(conn, 500, nil)
		return
	}

	var fileNames []string
	for _, file := range files {
		fileNames = append(fileNames, file.Name())
	}
	sendResponse(conn, 200, map[string][]string{
		"files": fileNames,
	})
}

// Напечатать относительный путь до текущей директории
func handlePWD(conn net.Conn, session *Session, sm *SessionManager) {
	sendResponse(conn, 200, map[string]string{
		"current_dir": session.CurrentDir,
	})
}

// Переместить содержимое первой директории во вторую
func handleCUT(conn net.Conn, args []string, session *Session, sm *SessionManager) {
	if len(args) < 2 {
		sendResponse(conn, 400, map[string]string{
			"error": "Не указаны исходный и целевой пути",
		})
		return
	}

	srcPath, err := sm.ResolvePath(session, args[0])
	if err != nil {
		sendResponse(conn, 403, map[string]string{
			"error": "Неверный исходный путь",
		})
		return
	}

	dstPath, err := sm.ResolvePath(session, args[1])
	if err != nil {
		sendResponse(conn, 403, map[string]string{
			"error": "Неверный целевой путь",
		})
		return
	}

	// Проверка занятости (рекурсивно для папок)
	if isBusy, _ := sm.IsPathBusy(srcPath); isBusy {
		sendResponse(conn, 423, map[string]string{
			"error": "Ресурс занят",
		})
		return
	}

	// Блокировка путей
	if !sm.LockPath(srcPath) || !sm.LockPath(dstPath) {
		sendResponse(conn, 423, map[string]string{
			"error": "Не удалось заблокировать путь",
		})
		return
	}
	defer sm.UnlockPath(srcPath)
	defer sm.UnlockPath(dstPath)

	// Перемещение
	if err := os.Rename(srcPath, dstPath); err != nil {
		sendResponse(conn, 500, map[string]string{
			"error": "Ошибка перемещения: " + err.Error(),
		})
		return
	}

	sendResponse(conn, 200, "Успешно перемещено")
}

// Рекурсивная проверка на занятость папок
func (sm *SessionManager) IsPathBusy(path string) (bool, []string) {
	var busyItems []string

	// Проверка самого пути
	if _, busy := sm.busyPaths.Load(path); busy {
		busyItems = append(busyItems, path)
	}

	// Рекурсивная проверка для папок
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		filepath.Walk(path, func(subPath string, info os.FileInfo, err error) error {
			if _, busy := sm.busyPaths.Load(subPath); busy {
				busyItems = append(busyItems, subPath)
			}
			return nil
		})
	}

	return len(busyItems) > 0, busyItems
}
