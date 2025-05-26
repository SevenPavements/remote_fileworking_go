package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
)

type SessionKeyTransfer struct {
	SessionKey string `json:"session_key"`
}

func processCommand(conn net.Conn, cmd string, token string) {
	// Для тестирования
	response := fmt.Sprintf("%s: Сервер получил команду: '%s'\n", token, cmd)
	log.Print(response)

	// Отправляем ответ клиенту
	conn.Write([]byte(response))
}

func handleConnection(conn net.Conn, sm *SessionManager) {
	defer conn.Close()

	// Создаем новую сессию
	token := sm.CreateSession(conn)

	// Формируем и отправляем ответ
	resp := SessionKeyTransfer{SessionKey: token}
	jsonResp, _ := json.Marshal(resp)
	conn.Write(jsonResp)

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		cmd := scanner.Text()
		processCommand(conn, cmd, token)
	}
}
