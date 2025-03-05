package main

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type Product struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Price       int    `json:"price"`
	Description string `json:"description"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	// Обслуживание статических файлов (admin.html и стили)
	fs := http.FileServer(http.Dir("../frontend"))
	http.Handle("/", fs)

	// REST API для получения товаров с backend-catalog
	http.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		res, err := http.Get("http://localhost:3000/products")
		if err != nil {
			http.Error(w, "Не удалось получить товары", http.StatusInternalServerError)
			return
		}
		defer res.Body.Close()

		// Прочитать тело ответа
		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			http.Error(w, "Ошибка при чтении ответа", http.StatusInternalServerError)
			return
		}

		// Отправить ответ клиенту
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(res.StatusCode)
		w.Write(bodyBytes)
	})

	// WebSocket маршрут для чата
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println("Ошибка при обновлении соединения:", err)
			return
		}
		defer conn.Close()

		// Настройка опций соединения
		conn.SetReadLimit(512)
		conn.SetReadDeadline(time.Now().Add(60 * time.Second)) // Установка таймаута чтения
		conn.SetPongHandler(func(appData string) error {
			conn.SetReadDeadline(time.Now().Add(60 * time.Second)) // Сброс таймаута после получения pong
			return nil
		})

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("Ошибка при чтении сообщения:", err)
				break
			}

			// Выводим полученное сообщение в консоль администратора
			fmt.Println("Получено сообщение:", string(message))

			// Отправляем сообщение обратно клиенту
			err = conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				fmt.Println("Ошибка при отправке сообщения:", err)
				break
			}
		}
	})

	// Подключение к WebSocket серверу backend-catalog
	go connectToCatalogChat()

	fmt.Println("Админ-панель работает на порту 8080")
	http.ListenAndServe(":8080", nil)
}

// Функция для подключения к WebSocket серверу backend-catalog
func connectToCatalogChat() {
	url := "ws://localhost:3000/ws"

	for {
		conn, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			fmt.Println("Не удалось подключиться к WebSocket серверу:", err)
			time.Sleep(5 * time.Second)
			continue
		}

		fmt.Println("Успешно подключились к WebSocket серверу backend-catalog")

		// Настройка опций соединения
		conn.SetReadLimit(512)
		conn.SetReadDeadline(time.Now().Add(60 * time.Second)) // Установка таймаута чтения
		conn.SetPongHandler(func(appData string) error {
			conn.SetReadDeadline(time.Now().Add(60 * time.Second)) // Сброс таймаута после получения pong
			return nil
		})

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("Ошибка при чтении сообщения с WebSocket сервера:", err)
				conn.Close()
				break
			}

			// Выводим сообщение в консоль администратора
			fmt.Println("Получено сообщение от backend-catalog:", string(message))
		}

		// Подождать перед следующей попыткой подключения
		time.Sleep(5 * time.Second)
	}
}
