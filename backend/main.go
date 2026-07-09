package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
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

var mockPoll = Poll{
	ID:       1,
	Question: "Какой стек выбрать для реалтайм пет-проекта?",
	Options: []Option{
		{ID: 1, Text: "Go + Flutter + Postgres", Votes: 10},
		{ID: 2, Text: "Node.js + React + MongoDB", Votes: 3},
		{ID: 3, Text: "Python + Vue + SQLite", Votes: 5},
	},
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("Input request",
		"method", r.Method,
		"path", r.URL.Path,
		"remote_addr", r.RemoteAddr,
	)

	fmt.Fprintf(w, "It is backend for realtime vote room.")
}

func getPollHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("Запрос данных опроса", "method", r.Method, "path", r.URL.Path)

	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(mockPoll)
	if err != nil {
		slog.Error("Ошибка кодирования JSON", "error", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/api/poll", getPollHandler)

	port := ":8080"
	slog.Info("Server successfully started", "port", port)

	err := http.ListenAndServe(port, nil)
	if err != nil {
		slog.Error("Critical server running error", "error", err)
		os.Exit(1)
	}
}
