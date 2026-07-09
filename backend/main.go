package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Option struct {
	ID    int    `json:"id"`
	Text  string `json:"text"`
	Votes int    `json:"votes"`
}

type Poll struct {
	ID       int      `json:"id"`
	Question string   `json:"question"`
	Options  []Option `json:"options"`
}

var dbPool *pgxpool.Pool

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var clients = make(map[*websocket.Conn]bool)

var clientsMu sync.Mutex

func broadcastPollState() {
	ctx := context.Background()
	var poll Poll

	err := dbPool.QueryRow(ctx, "SELECT id, question FROM polls WHERE id = $1", 1).Scan(&poll.ID, &poll.Question)
	if err != nil {
		slog.Error("Broadcast: Getting poll error", "error", err)
		return
	}

	rows, err := dbPool.Query(ctx, "SELECT id, text, votes FROM options WHERE poll_id = $1 ORDER BY id", poll.ID)
	if err != nil {
		slog.Error("Broadcast: Getting options error", "error", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var opt Option
		rows.Scan(&opt.ID, &opt.Text, &opt.Votes)
		poll.Options = append(poll.Options, opt)
	}

	jsonData, err := json.Marshal(poll)
	if err != nil {
		slog.Error("Broadcast: Marshaling JSON error", "error", err)
		return
	}

	clientsMu.Lock()
	defer clientsMu.Unlock()

	for client := range clients {
		err := client.WriteMessage(websocket.TextMessage, jsonData)
		if err != nil {
			slog.Warn("Unable to send message to client. Closing connection", "error", err)
			client.Close()
			delete(clients, client)
		}
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Upgrade error in WebSocket", "error", err)
		return
	}

	clientsMu.Lock()
	clients[conn] = true
	clientsMu.Unlock()
	slog.Info("New client connected via WebSocket", "addr", conn.RemoteAddr())

	go broadcastPollState()

	go func() {
		defer func() {
			clientsMu.Lock()
			delete(clients, conn)
			clientsMu.Unlock()
			conn.Close()
			slog.Info("Client disconnected", "addr", conn.RemoteAddr())
		}()

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				break
			}

			var voteReq struct {
				OptionID int `json:"option_id"`
			}

			if err := json.Unmarshal(message, &voteReq); err != nil {
				slog.Warn("Incorrect client message format", "error", err)
				continue
			}

			slog.Info("Got vote", "option_id", voteReq.OptionID)

			_, err = dbPool.Exec(context.Background(),
				"UPDATE options SET votes = votes + 1 WHERE id = $1", voteReq.OptionID)
			if err != nil {
				slog.Error("Update vote database error", "error", err)
				continue
			}

			go broadcastPollState()
		}
	}()
}

func getPollHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := context.Background()

	var poll Poll

	err := dbPool.QueryRow(ctx, "SELECT id, question FROM polls WHERE id = $1", 1).Scan(&poll.ID, &poll.Question)
	if err != nil {
		slog.Error("Getting poll database error", "error", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	rows, err := dbPool.Query(ctx, "SELECT id, text, votes FROM options WHERE poll_id = $1 ORDER BY id", poll.ID)
	if err != nil {
		slog.Error("Getting option database error", "error", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var opt Option
		if err := rows.Scan(&opt.ID, &opt.Text, &opt.Votes); err != nil {
			slog.Error("Scan option row error", "error", err)
			continue
		}
		poll.Options = append(poll.Options, opt)
	}

	json.NewEncoder(w).Encode(poll)
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	connStr := os.Getenv("DB_URL")
	if connStr == "" {
		connStr = "postgres://postgres:mypassword@127.0.0.1:5433/mydb?sslmode=disable"
	}

	ctx := context.Background()
	var err error

	dbPool, err = pgxpool.New(ctx, connStr)
	if err != nil {
		slog.Error("Unable to create connection pool", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	err = dbPool.Ping(ctx)
	if err != nil {
		slog.Error("Database is not available (Ping failed)", "error", err)
		os.Exit(1)
	}
	slog.Info("Successful connection to PostgreSQL")

	http.HandleFunc("/api/poll", getPollHandler)
	http.HandleFunc("/ws", wsHandler)

	port := ":8080"
	slog.Info("Server successfully start", "port", port)

	err = http.ListenAndServe(port, nil)
	if err != nil {
		slog.Error("Critical server start error", "error", err)
		os.Exit(1)
	}
}
