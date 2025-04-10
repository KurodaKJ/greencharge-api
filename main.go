package main

import (
	"context"
	"log"

	"greencharge-api/handlers"
	"greencharge-api/server"

	// Remove the older messaging import if using v4 consistently
	// "firebase.google.com/go/messaging"
	firebase "firebase.google.com/go/v4"
	// Import the v4 messaging package explicitly if needed, or rely on firebase.App methods
	// Note: Often the client type comes directly from the app object, so a separate v4/messaging import might not be needed here.
	// Let's adjust the return type using the alias 'firebase' for the core v4 package.
	"firebase.google.com/go/v4/messaging" // We need this for the return type signature if not using the alias method
	"google.golang.org/api/option"
)

// Modified function signature to return the v4 messaging client type
func initFirebase() *messaging.Client { // Adjusted return type
	opt := option.WithCredentialsFile("YOUR_CREDENTIAL_PATH") // Replace with the actual path
	// Use the v4 firebase package (aliased)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}

	// This returns *firebase.messaging.Client (v4 type)
	fcmClient, err := app.Messaging(context.Background())
	if err != nil {
		log.Fatalf("error initializing fcm client: %v\n", err)
	}
	// Now the return type matches the function signature
	return fcmClient
}

func main() {
	// Get Firebase config
	credentialPath := "YOUR_CREDENTIAL_PATH"
	opt := option.WithCredentialsFile(credentialPath)
	// Config struct might be defined in firebase/v4 package, ensure correct usage if needed
	// conf := &firebase.Config{} // This might need adjustment based on handlers.NewFcmClient expectations
	// If handlers.NewFcmClient only needs context and options, conf might be nil or empty struct?
	// Let's assume conf is correct for handlers.NewFcmClient for now.
	conf := &firebase.Config{}

	// Initialize Firebase via handlers function
	// Corrected Assignment: Expect only 'error' based on compiler message
	// _, err := handlers.NewFcmClient(context.Background(), conf, opt) // Before
	err := handlers.NewFcmClient(context.Background(), conf, opt) // After
	if err != nil {
		log.Fatalf("error initializing Firebase via handlers: %v\n", err)
		// No need for return here, log.Fatalf exits the program
	}

	// Note: The result of handlers.NewFcmClient (presumably the client itself, if it returned one)
	// is not captured or used here. Ensure this is intended.

	// Initialize the router and routes
	router := server.InitRouter()
	handlers.InitRoutes(router) // Assumes InitRoutes doesn't need the FCM client passed explicitly
	// Start the server
	server.StartServer()
}
