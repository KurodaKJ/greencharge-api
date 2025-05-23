package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"greencharge-api/handlers"
	"greencharge-api/server" // Assuming setupRouter comes from here or similar

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// helper function to create a test server
// Modified to pass 't' for error reporting
func setupTestServer(t *testing.T) *httptest.Server {
	router := setupRouter() // call to setup the router
	server := httptest.NewServer(router)

	// Initialize Firebase Admin SDK
	credentialPath := "YOUR_CREDENTIALS_PATH" // Path to your credentials file
	conf := &firebase.Config{
		ProjectID: "YOUR_PROJECT_ID",
	}
	ctx := context.Background()
	// Corrected Assignment: Expect only 'error' based on compiler message
	err := handlers.NewFcmClient(ctx, conf, option.WithCredentialsFile(credentialPath))
	if err != nil {
		// Use t.Fatalf for better test failure reporting instead of panic
		t.Fatalf("Failed to initialize FCM client during test setup: %v", err)
	}

	return server
}

// Removed the local NewFcmClient wrapper function as it was causing errors
// and seemed redundant based on the compiler messages.

// test the root endpoint
func TestRootHandler(t *testing.T) {
	server := setupTestServer(t) // Pass t to the helper function
	defer server.Close()
	// make a request to the root endpoint
	resp, err := http.Get(server.URL + "/")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Check the status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
	// check the body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	body := string(bodyBytes)
	expected := "Hello, this is the root endpoint!\n"

	if body != expected {
		t.Errorf("Expected body %q, got %q", expected, body)
	}
}

// Test the REST endpoint
func TestRestHandler(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	// Prepare the JSON payload
	payload := map[string]string{
		"message": "alert",
		"title":   "Test Alert 🚨",
		"body":    "This is a test notification.",
		"token":   "dtPub3DnRQWt5E3fDzvzOZ:APA91bGm8XrMgqBEVZ9wp5v6RAXeH5fDwuP_57h43xtL9f3O8nWNmLwCBfAcBuVsZfKXZ2DRPt7BdWdct2dSF2DP-3WQdrxC6Pp8xCKBz31FtOKKjPY3k7Y",
	}
	payloadBytes, _ := json.Marshal(payload)

	// Create a new POST request with the JSON payload
	resp, err := http.Post(server.URL+"/restyet", "application/json", io.NopCloser(bytes.NewReader(payloadBytes)))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	var msg handlers.Message
	err = json.NewDecoder(resp.Body).Decode(&msg)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	expectedAction := "message_received"
	if msg.Action != expectedAction {
		t.Errorf("Expected action %q, got %q", expectedAction, msg.Action)
	}
}

func TestWsHandler(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	// Establish a WebSocket connection
	wsURL := "ws" + server.URL[4:] + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	// Send a message with custom title and body
	message := handlers.MessageWithToken{
		Message: "alert",
		Title:   "Test Alert 🚨",
		Body:    "This is a test notification.",
		Token:   "YOUR_DEVICE_TOKEN_HERE", // Replace with a valid token
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}

	err = conn.WriteMessage(websocket.TextMessage, messageBytes)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	// Read the echoed message
	_, received, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read message: %v", err)
	}

	// Check if the received message matches what was sent (minus title/body)
	var receivedMessage handlers.Message
	err = json.Unmarshal(received, &receivedMessage)
	if err != nil {
		t.Fatalf("Failed to unmarshal received message: %v", err)
	}

	if receivedMessage.Message != message.Message {
		t.Errorf("Expected message %q, got %q", message.Message, receivedMessage.Message)
	}

	if receivedMessage.Action != "message_received" {
		t.Errorf("Expected action %q, got %q", "message_received", receivedMessage.Action)
	}
}

// setup a new router
func setupRouter() *mux.Router {
	r := server.InitRouter() // Assumes server package provides InitRouter
	handlers.InitRoutes(r)   // Assumes handlers package provides InitRoutes
	return r
}
