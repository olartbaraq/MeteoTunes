package main

import (
	"encoding/json"
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

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"message": "welcome to meteotunes"})
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{"message": "welcome to meteojjbjtunes"})
}

func (s *Server) Start() {

	stack := CreateStack(
		loggingMiddleware,
		enableCors,
		s.ValidateToken,
	)

	handler := stack((s.router))

	s.router.HandleFunc("GET /home", homeHandler)
	s.router.HandleFunc("GET /user/me", userHandler)

	http.ListenAndServe(fmt.Sprintf(":%d", s.config.PORT), handler)
}
