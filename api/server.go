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

func (s *Server) Start() {

	stack := CreateStack(
		loggingMiddleware,
		enableCors,
		s.ValidateToken,
	)

	handler := stack((s.router))

	s.router.HandleFunc("GET /home", homeHandler)
	s.router.HandleFunc("GET /user/me", userHandler)
	s.router.HandleFunc("POST /load", s.fetchMusic)

	http.ListenAndServe(fmt.Sprintf(":%d", s.config.PORT), handler)
}
