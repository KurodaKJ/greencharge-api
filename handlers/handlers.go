package handlers

import (
	"context"
	"encoding/json"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
	"firebase.google.com/go/v4"
	"log"
	"net/http"
	"syscall"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// We'll need an upgrader to handle upgrading regular HTTP connections to WebSockets.
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// CheckOrigin allows connections from any origin.  In a production
		// environment, this should be restricted to specific origins.
		return true
	},
}

// Define a simple message structure for our REST endpoint
type Message struct {
	Message string `json:"message"`
	Action  string `json:"action"`
}

type MessageWithToken struct {
	Message string `json:"message"`
	Title   string `json:"title"`
	Body    string `json:"body"`
	Token   string `json:"token"`
}

var fcmService *messaging.Client

// app is a global Firebase app instance.
var app *firebase.App

// NewFcmClient initializes the fcmClient for sending notifications.
func NewFcmClient(ctx context.Context, config *firebase.Config, opt ...option.ClientOption) error {
	var err error
	app, err = firebase.NewApp(ctx, config, opt...)
	if err != nil {
		return err
	}

	fcmService, err = app.Messaging(ctx)
	if err != nil {
		return err
	}
	return nil
}

func InitRoutes(router *mux.Router) {
	// Define a handler for the root path.  This is a simple greeting message.
	router.HandleFunc("/", rootHandler).Methods("GET")

	// Define a handler for WebSocket connections on the "/ws" path.
	router.HandleFunc("/ws", wsHandler)

	// Define a REST endpoint on "/rest" which returns a JSON message.
	router.HandleFunc("/restyet", restHandler).Methods("GET")

	// Add a shutdown endpoint
	router.HandleFunc("/shutdown", shutdownHandler).Methods("GET")
}

// rootHandler handles requests to the root path "/".
func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, this is the root endpoint!\n"))
}

// wsHandler handles WebSocket connections.  It upgrades the HTTP connection and
// then echoes back any messages it receives from the client.
func wsHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade the HTTP connection to a WebSocket connection.
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}
	defer conn.Close() // Ensure the connection is closed when the function exits.

	// Echo messages back to the client until the connection is closed.
	for {
		mt, messageBytes, err := conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		log.Printf("received message: %s", messageBytes)

		var message MessageWithToken
		if err := json.Unmarshal(messageBytes, &message); err != nil {
			log.Printf("error unmarshaling message: %v", err)
			break
		}

		// Here you process the message, and check the condition
		if shouldSendNotification(message.Message, message.Token) { // This is a custom function

			err := SendNotification(prepareMessage(message.Token, message.Title, message.Body))

			if err != nil {
				log.Printf("error sending notification: %v", err)
			}
		}

		response := Message{
			Message: message.Message,
			Action:  "message_received",
		}

		jsonResponse, _ := json.Marshal(response)

		err = conn.WriteMessage(mt, jsonResponse)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

// restHandler handles requests to the "/rest" path. It returns a JSON message.
func restHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(Message{Message: "Hello from REST"})
}

func shutdownHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Server shutting down...\n"))
	// Trigger server shutdown by sending a signal to the stop channel
	syscall.Kill(syscall.Getpid(), syscall.SIGINT)
}

func shouldSendNotification(message string, token string) bool {
	// Example logic: Send a notification if the message contains the word "alert"
	if message == "alert" && token != "" {
		return true
	}
	return false
}

// SetFcmService sets the FCM service client.
func SetFcmService(fcm *messaging.Client) {
	fcmService = fcm
}

func prepareMessage(token, title, body string) *messaging.Message {
	return &messaging.Message{
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Token: token,
	}
}

// SendNotification sends a notification message to the specified device.
func SendNotification(m *messaging.Message) error {
	if _, err := fcmService.Send(context.Background(), m); err != nil {
		return err
	}
	return nil
}
