package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	handler "drive/handler"
)

func main() {
	args := os.Args

	if len(args) < 4 {
		log.Fatal("Please provide the command line argument \n USAGE./bin [client_id] [client_secret] [file_id]")
	}
	clientId := args[1]
	clientSecret := args[2]
	fileId := args[3]

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("templates/")))

	fmt.Println(clientId, clientSecret, fileId)

	service := handler.NewService(clientId, clientSecret, fileId)

	mux.HandleFunc("/login", service.Login)
	mux.HandleFunc("/callback", service.Callback)
	server := &http.Server{
		Addr:    fmt.Sprintf(":3000"),
		Handler: mux,
	}

	log.Printf("Starting HTTP Server. Listening at %q", server.Addr)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Printf("%v", err)
	} else {
		log.Println("Server closed!")
	}
}
