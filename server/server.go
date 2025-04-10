package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"greencharge-api/handlers"

	"github.com/gorilla/mux"
)

func StartServer() {
	log.Print("Starting server...")
	// Determine the port to listen on.  If the PORT environment variable is set, use that.
	// Otherwise, default to port 3000.
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
		log.Printf("Defaulting to port %s", port)
	}

	// Create a new router using gorilla/mux
	router := InitRouter()

	// Start the HTTP server, listening on the specified port.  Use the router we created.
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Create a channel to listen for OS signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		log.Printf("Listening on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on %s: %v\n", port, err)
		}
	}()

	// Block until a signal is received
	<-stop
	log.Println("Shutting down server...")

	// Create a context with a timeout for the shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Gracefully shut down the server
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Could not gracefully shutdown the server: %v\n", err)
	}
	log.Println("Server stopped")
}

func InitRouter() *mux.Router {
	router := mux.NewRouter()
	handlers.InitRoutes(router)
	return router
}
