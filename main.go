package main

import (
	"embed"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type Tasks struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	IsDone bool   `json:"is_done"`
}

var (
	//go:embed index.html
	front  embed.FS
	tasks  []Tasks
	nextID = 1
	mu     sync.Mutex
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data, err := front.ReadFile("index.html")
		if err != nil {
			http.Error(w, "index.html not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	})

	mux.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			listTodos(w)
		case http.MethodPost:
			createTodo(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/tasks/", func(w http.ResponseWriter, r *http.Request) {
		idStr := r.URL.Path[len("/tasks/"):]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodPut:
			toggleTodo(w, id)
		case http.MethodDelete:
			deleteTodo(w, id)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func createTodo(w http.ResponseWriter, r *http.Request) {
	var newTodo Tasks
	if err := json.NewDecoder(r.Body).Decode(&newTodo); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if newTodo.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	mu.Lock()
	newTodo.ID = nextID
	nextID++
	newTodo.IsDone = false
	tasks = append(tasks, newTodo)
	mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newTodo)
}

func listTodos(w http.ResponseWriter) {
	mu.Lock()
	defer mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func toggleTodo(w http.ResponseWriter, id int) {
	mu.Lock()
	defer mu.Unlock()

	for i, t := range tasks {
		if t.ID == id { // exported field
			tasks[i].IsDone = !tasks[i].IsDone
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(tasks[i])
			return
		}
	}
	http.Error(w, "task not found", http.StatusNotFound)
}

func deleteTodo(w http.ResponseWriter, id int) {
	mu.Lock()
	defer mu.Unlock()

	for i, t := range tasks {
		if t.ID == id { // exported field
			tasks = append(tasks[:i], tasks[i+1:]...)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	http.Error(w, "task not found", http.StatusNotFound)
}
