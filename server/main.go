package main

import (
	"flag"
	"log"
	"net"
)

func main() {

	portpt := flag.String("p", "1234", "Порт сервера")
	RootDir := "C:\\For_network_technologies\\Working_directory"
	flag.Parse()
	port := *portpt

	sm := NewSessionManager(RootDir)

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Ошибка при запуске сервера: %v", err)
	}
	defer listener.Close()

	log.Println("Сервер запущен и слушает порт " + port)

	// Бесконечный цикл сервера
	for {
		// Приём входящего соединения
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Ошибка при принятии соединения: %v", err)
			continue
		}

		// Обработка в отдельной горутине. Так как пользователей планируется немного, никакого кода сверх этого по идее писать не нужно
		go handleConnection(conn, sm)
	}
}
