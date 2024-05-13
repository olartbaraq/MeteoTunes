package api

import (
	"fmt"
	"net/http"
)

type Server struct {
	router *http.ServeMux
	config *Config
}

func NewServer(envPath string) *Server {
	config, err := LoadConfig(envPath)
	if err != nil {
		panic(fmt.Sprintf("Could not load env.env config: %v", err))
	}

	mux := http.NewServeMux()

	return &Server{
		router: mux,
		config: config,
	}
}

func Routes() http.Handler {
	server := NewServer(".")

	stack := CreateStack(
		loggingMiddleware,
		enableCors,
		server.ValidateToken,
	)

	handler := stack((server.router))

	server.router.HandleFunc("GET /home", homeHandler)
	server.router.HandleFunc("GET /user/me", userHandler)
	server.router.HandleFunc("POST /load", server.fetchMusic)

	return handler
}
