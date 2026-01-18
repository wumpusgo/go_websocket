package main

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"slices"

	"github.com/gorilla/websocket"
)

type Client struct {
	id string
}

// Upgrader is used to upgrade HTTP connections to WebSocket connections.
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[string]*websocket.Conn)

func getConnectedUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	clientsObj := make([]Client, 0, 10)
	w.Header().Set("Content-Type", "application/json")
	list := slices.Collect(maps.Keys(clients))
	for key, value := range list {
		fmt.Println(key, value)
		clientsObj = append(clientsObj, Client{id: value})
	}
	fmt.Println("clientes ", len(clients))
	err := json.NewEncoder(w).Encode(clientsObj)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// fmt.Fprintf(w, "Hello, you made a GET request!")
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	to := r.URL.Query().Get("to")
	conn, err := upgrader.Upgrade(w, r, nil)
	clients[username] = conn
	fmt.Println("User ", username, "connected")
	if err != nil {
		fmt.Println("Error upgrading:", err)
		return
	}
	// defer conn.Close()
	defer closeConnections(username)
	// Listen for incoming messages
	for {
		// Read message from the client
		_, message, err := clients[username].ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			break
		}

		fmt.Printf("from ", username, " Received: %s\\n", message, "to ", to)
		// Echo the message back to the client
		if err := clients[to].WriteMessage(websocket.TextMessage, message); err != nil {
			fmt.Println("Error writing message:", err)
			break
		}
	}
}

func closeConnections(user string) {
	clients[user].Close()
	delete(clients, user)
	fmt.Printf("User ", user, " disconnected")
}

func main() {
	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/clients", getConnectedUsers)
	fmt.Println("WebSocket server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
