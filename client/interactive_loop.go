package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

func RunInteractiveLoop(session *Session) {
	defer fmt.Println("Завершение сессии...")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		cmd := scanner.Text()
		if shouldExit(cmd) {
			break
		}

		if err := processCommand(session, cmd); err != nil {
			log.Printf("Ошибка: %v", err)
		}
	}
}

func shouldExit(cmd string) bool {
	return cmd == "exit" || cmd == "quit"
}

func processCommand(session *Session, cmd string) error {
	// Отправка команды
	if err := sendCommand(session.Conn, cmd); err != nil {
		return fmt.Errorf("ошибка отправки: %w", err)
	}

	// Получение ответа
	response, err := readResponse(session.Conn)
	if err != nil {
		return fmt.Errorf("ошибка чтения: %w", err)
	}

	fmt.Println("Сервер:", response)
	return nil
}

func sendCommand(conn net.Conn, cmd string) error {
	conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
	_, err := fmt.Fprintf(conn, "%s\n", cmd)
	return err
}

func readResponse(conn net.Conn) (string, error) {
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	reader := bufio.NewReader(conn)
	return reader.ReadString('\n')
}
