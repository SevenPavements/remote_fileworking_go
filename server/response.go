package main

import (
	"encoding/binary"
	"encoding/json"
	"net"
	"net/http"
)

// type Response struct {
// 	Status  string `json:"status"`
// 	Message string `json:"message,omitempty"`
// 	Data    string `json:"data,omitempty"`
// }

// Отправка ответа с размером данных в начале
func sendResponse(conn net.Conn, code int, data any) error {
	if msg, ok := data.(string); ok {
		if code >= 400 {
			data = map[string]string{"error": msg}
		} else {
			data = map[string]string{"message": msg}
		}
	}

	resp := struct {
		Code int    `json:"code"`
		Data any    `json:"data,omitempty"`
		Err  string `json:"error,omitempty"`
	}{
		Code: code,
		Data: data,
	}
	if code >= 400 {
		resp.Err = http.StatusText(code)
	}

	jsonData, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	// Определяем размер (4 байта, big-endian)
	sizeBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeBuf, uint32(len(jsonData)))

	// Отправляем размер и данные
	if _, err := conn.Write(sizeBuf); err != nil {
		return err
	}

	if _, err := conn.Write(jsonData); err != nil {
		return err
	}

	return nil
}
