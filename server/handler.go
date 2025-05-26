package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
)

type SessionKeyTransfer struct {
	SessionKey string `json:"session_key"`
}

func processCommand(conn net.Conn, cmd string, token string) {
	// Для тестирования
	response := fmt.Sprintf("%s: Сервер получил команду: '%s'\n", token, cmd)
	log.Print(response)

	parts := strings.Split(cmd, " ")
	switch parts[0] {
	case "cd":
		if len(parts) < 2 {
			conn.Write([]byte("Укажите путь\n"))
			return
		}
		// 	handleCD(conn, parts[1], session, sm)
		// case "upload":
		// 	handleUpload(conn, session, sm)
		// ... другие команды ...
	}

	// Отправляем ответ клиенту
	conn.Write([]byte(response))
}

func handleConnection(conn net.Conn, sm *SessionManager) {
	defer conn.Close()

	// Создаем новую сессию
	session := sm.CreateSession(conn)
	token := session.ID

	// Формируем и отправляем ответ
	resp := SessionKeyTransfer{SessionKey: token}
	jsonResp, _ := json.Marshal(resp)
	conn.Write(jsonResp)

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		session.UpdateActivity()
		cmd := scanner.Text()
		processCommand(conn, cmd, token)
	}
}
