package main

import (
	"backend-catalog/graph"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/websocket"
)

const defaultPort = "3000"

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Хранилище сообщений чата
var chatMessages = []string{}
var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan string)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Создание GraphQL сервера
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{}}))
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.GET{})
	srv.Use(extension.Introspection{})

	// Определение пути к папке frontend
	frontendDir := filepath.Join("..", "frontend")

	// Настройка маршрутов
	http.Handle("/graphql", srv)                                          // Маршрут для GraphQL API
	http.Handle("/playground", playground.Handler("GraphQL", "/graphql")) // Playground для разработки
	http.Handle("/", http.FileServer(http.Dir(frontendDir)))              // Статические файлы из папки frontend

	// WebSocket маршрут
	http.HandleFunc("/ws", handleWebSocket)

	// Запуск горутины для отправки сообщений всем клиентам
	go broadcastMessages()

	log.Printf("Server running at http://localhost:%s/", port)
	log.Printf("GraphQL playground at http://localhost:%s/playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// Обработка WebSocket соединений
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Ошибка при обновлении соединения:", err)
		return
	}
	defer conn.Close()

	clients[conn] = true

	// Отправляем историю сообщений новому клиенту
	for _, msg := range chatMessages {
		err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			log.Println("Ошибка при отправке истории сообщений:", err)
			break
		}
	}

	// Таймер для отправки пинга каждые 30 секунд
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Отправляем пинг
			err := conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				log.Println("Ошибка при отправке пинга:", err)
				delete(clients, conn)
				return
			}
		default:
			// Читаем сообщения от клиента
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("Ошибка при чтении сообщения:", err)
				delete(clients, conn)
				return
			}

			// Добавляем сообщение в историю
			chatMessages = append(chatMessages, string(message))
			if len(chatMessages) > 100 { // Ограничение истории до 100 сообщений
				chatMessages = chatMessages[1:]
			}

			// Отправляем сообщение всем клиентам
			broadcast <- string(message)
		}
	}
}

// Бродкаст сообщений всем клиентам
func broadcastMessages() {
	for {
		msg := <-broadcast
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				log.Println("Ошибка при отправке сообщения:", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
