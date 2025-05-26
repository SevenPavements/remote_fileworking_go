package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
)

type SessionKeyTransfer struct {
	SessionKey string `json:"session_key"`
}

type Session struct {
	Conn       net.Conn
	SessionKey string
}

func main() {

	hostippt := flag.String("h", "localhost", "Хост сервера")
	portpt := flag.String("p", "1234", "Порт сервера")
	flag.Parse()

	hostip := *hostippt
	port := *portpt

	// Подключение к серверу
	conn, err := net.Dial("tcp", hostip+":"+port)
	if err != nil {
		log.Fatalf("Ошибка при подключении к серверу: %v", err)
	}
	defer conn.Close()

	// Читаем ответ
	var resp SessionKeyTransfer
	jsonData := make([]byte, 1024) // Буфер с запасом
	n, err := conn.Read(jsonData)
	if err != nil {
		log.Fatalf("Ошибка принятия: %v", err)
	}
	jsonData = jsonData[:n] // Обрезаем лишнее
	err = json.Unmarshal([]byte(jsonData), &resp)
	if err != nil {
		log.Fatalf("Ошибка чтения: %v", err)
	}
	sessionKey := resp.SessionKey
	session := Session{
		Conn:       conn,
		SessionKey: sessionKey,
	}

	log.Printf("Подключено к серверу, сессионный ключ: %s", sessionKey)
	fmt.Printf("Успешно подключено, Ваш ключ: %s\n", sessionKey)
	fmt.Println("Сохраните его на всякий случай.")

	RunInteractiveLoop(&session)
}
