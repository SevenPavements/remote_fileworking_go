package main

import (
	"bufio"
	"log"
	"net"
	"strings"
	"time"
)

func handleConnection(conn net.Conn, sm *SessionManager) {
	defer conn.Close()

	// Создаем сессию
	session := NewSession(conn)
	sm.mutex.Lock()
	sm.sessions[session.ID] = session
	sm.mutex.Unlock()

	// Отправляем только session_key (код 200 + данные)
	err := sendResponse(conn, 200, map[string]interface{}{
		"session_key": session.ID,
	})
	if err != nil {
		log.Println("Ошибка отправки session_key:", err)
		return
	}

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		session.LastActivity = time.Now()

		cmd := strings.TrimSpace(scanner.Text())
		if cmd == "" {
			continue
		}

		// Разбиваем на части по пробелам
		parts := strings.Split(cmd, " ")
		switch parts[0] {
		case "cd":
			if len(parts) < 2 {
				sendResponse(conn, 400, "Укажите путь")
				continue
			}
			handleCD(conn, parts[1:], session, sm)

		case "ls":
			handleLS(conn, session, sm)

		case "pwd":
			handlePWD(conn, session, sm)

		case "cut":
			if len(parts) < 3 {
				sendResponse(conn, 400, "Укажите исходный и целевой пути")
				continue
			}
			handleCUT(conn, parts[1:], session, sm)

		case "help":
			sendHelp(conn)

		default:
			sendResponse(conn, 400, "Неизвестная команда")
		}
	}
}

func sendHelp(conn net.Conn) {
	helpText := map[string]string{
		"help": "Показать эту справку",
		"cd":   "cd <путь> - сменить директорию",
		"ls":   "ls - список файлов",
		"pwd":  "pwd - текущая директория",
		"cut":  "cut <откуда> <куда> - переместить файл",
	}
	sendResponse(conn, 200, helpText)
}
