package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

const (
	ServerHost = "localhost"
	ServerPort = "1234"
)

type Session struct {
	Conn       net.Conn
	SessionKey string
	CurrentDir string
}

type Response struct {
	Code  int                    `json:"code"`
	Data  map[string]interface{} `json:"data,omitempty"`
	Error string                 `json:"error,omitempty"`
}

func main() {
	fmt.Println("=== Файловый менеджер ===")

	conn, err := net.Dial("tcp", ServerHost+":"+ServerPort)
	if err != nil {
		log.Fatalf("Ошибка подключения: %v", err)
	}
	defer conn.Close()

	session, err := initSession(conn)
	if err != nil {
		log.Fatalf("Ошибка инициализации сессии: %v", err)
	}

	fmt.Printf(">Успешное подключение. Сессия: %s\n", session.SessionKey)
	fmt.Println("Введите 'help' для списка команд")

	runInteractiveShell(session)
}

func initSession(conn net.Conn) (*Session, error) {
	// Читаем приветственное сообщение с session_key
	reader := bufio.NewReader(conn)
	response, err := readServerResponse(reader)
	if err != nil {
		return nil, err
	}

	if response.Code != 200 {
		return nil, errors.New(response.Error)
	}

	sessionKey, ok := response.Data["session_key"].(string)
	if !ok {
		return nil, errors.New("неверный формат session_key")
	}

	currentDir := "/"
	if dir, ok := response.Data["current_dir"].(string); ok {
		currentDir = dir
	}

	return &Session{
		Conn:       conn,
		SessionKey: sessionKey,
		CurrentDir: currentDir,
	}, nil
}

func readServerResponse(reader *bufio.Reader) (*Response, error) {
	// Читаем размер сообщения (4 байта)
	sizeBuf := make([]byte, 4)
	if _, err := io.ReadFull(reader, sizeBuf); err != nil {
		return nil, err
	}

	size := binary.BigEndian.Uint32(sizeBuf)
	// Читаем само сообщение
	jsonData := make([]byte, size)
	if _, err := io.ReadFull(reader, jsonData); err != nil {
		return nil, err
	}

	var response Response
	if err := json.Unmarshal(jsonData, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func runInteractiveShell(session *Session) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf(">Текущая директория: %s\n> ", session.CurrentDir)
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		if input == "exit" || input == "quit" {
			fmt.Println("Завершение сессии...")
			return
		}

		if input == "help" {
			printHelp()
			continue
		}

		parts := strings.Fields(input)
		if len(parts) == 0 {
			continue
		}

		command := parts[0]

		// Отправляем команду на сервер
		if err := sendCommand(session, input); err != nil {
			fmt.Printf(">Ошибка отправки команды: %v\n", err)
			continue
		}

		// Читаем ответ
		response, err := readServerResponse(bufio.NewReader(session.Conn))
		if err != nil {
			fmt.Printf(">Ошибка чтения ответа: %v\n", err)
			continue
		}

		// Обрабатываем ответ
		if err := handleResponse(session, response, command); err != nil {
			fmt.Printf(">Ошибка: %v\n", err)
		}
	}
}

func sendCommand(session *Session, cmd string) error {
	// Убираем все лишние символы
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return nil
	}

	// Преобразуем команду в байты
	cmdBytes := []byte(cmd + "\n") // Добавляем \n как разделитель команд

	// // Получаем размер сообщения (4 байта)
	// sizeBuf := make([]byte, 4)
	// binary.BigEndian.PutUint32(sizeBuf, uint32(len(cmdBytes)))

	// // Отправляем по отдельности
	// if _, err := session.Conn.Write(sizeBuf); err != nil {
	// 	return err
	// }
	if _, err := session.Conn.Write(cmdBytes); err != nil {
		return err
	}

	return nil
}

func handleResponse(session *Session, resp *Response, command string) error {
	switch resp.Code {
	case 200: // Успех
		// Обработка успешных ответов в зависимости от команды
		switch command {
		case "cd":
			return handleCDResponse(session, resp.Data)
		case "ls":
			return handleLSResponse(resp.Data)
		case "pwd":
			return handlePWDResponse(resp.Data)
		case "cut":
			return handleCutResponse(resp.Data)
		default:
			return fmt.Errorf("неизвестный тип команды для обработки ответа")
		}

	case 400, 403, 404, 423, 500: // Ошибки
		return errors.New(resp.Error)

	default:
		return fmt.Errorf("неизвестный код ответа: %d", resp.Code)
	}
}

// Обработчики для конкретных команд
func handleCDResponse(session *Session, data map[string]interface{}) error {
	dir, ok := data["current_dir"].(string)
	if !ok {
		return errors.New("неверный формат ответа для cd")
	}
	session.CurrentDir = dir
	fmt.Printf("> Текущая директория изменена: %s\n", dir)
	return nil
}

func handleLSResponse(data map[string]interface{}) error {
	filesRaw, ok := data["files"]
	if !ok {
		return errors.New("отсутствует поле files в ответе")
	}

	// Обрабатываем случай null и пустого массива
	if filesRaw == nil {
		fmt.Println("> Папка пуста")
		return nil
	}

	files, ok := filesRaw.([]interface{})
	if !ok {
		return fmt.Errorf("неверный формат списка файлов, получен тип: %T", filesRaw)
	}

	if len(files) == 0 {
		fmt.Println("> Папка пуста")
		return nil
	}

	fmt.Println("> Содержимое директории:")
	for i, f := range files {
		fmt.Printf("%3d. %v\n", i+1, f)
	}
	return nil
}

func handlePWDResponse(data map[string]interface{}) error {
	dir, ok := data["current_dir"].(string)
	if !ok {
		return errors.New("неверный формат ответа для pwd")
	}
	fmt.Printf("> Текущая директория: %s\n", dir)
	return nil
}

func handleCutResponse(data map[string]interface{}) error {
	message, ok := data["message"].(string)
	if !ok {
		return errors.New("неверный формат ответа для cut")
	}
	fmt.Printf("> Успех: %s\n", message)
	return nil
}

func printHelp() {
	fmt.Println("\nДоступные команды:")
	fmt.Println("  pwd                  - показать текущую директорию")
	fmt.Println("  ls                   - список файлов")
	fmt.Println("  cd <путь>            - сменить директорию")
	fmt.Println("  cut <откуда> <куда>  - переместить файл/папку")
	fmt.Println("  help                 - эта справка")
	fmt.Println("  exit/quit            - выход\n")
}
