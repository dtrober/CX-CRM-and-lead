package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/dyrober/AgencyCRM/internal/config"
	"github.com/dyrober/AgencyCRM/internal/domain"
	"github.com/dyrober/AgencyCRM/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// make a server obj
type Server struct {
	*http.Server
	service *service.Service
}

// create a new http server
func NewServer(cfg *config.Config, svc *service.Service) *Server {
	r := chi.NewRouter()

	//Middle ware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	//create a server
	srv := &Server{
		Server: &http.Server{
			Addr:         cfg.ServerAddress,
			Handler:      r,
			ReadTimeout:  cfg.ServerReadTimeout,
			WriteTimeout: cfg.ServerWriteTimeout,
			IdleTimeout:  120 * time.Second,
		},
		service: svc,
	}

	//Register routes
	r.Get("/health", srv.healthCheck)

	//API Routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/users", func(r chi.Router) {
			r.Post("/", srv.createUser)
			r.Get("/{id}", srv.getUser)
		})
	})
	return srv
}

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	}
	respondJSON(w, http.StatusOK, response)
}

// GetUser grabs a user by ID
func (s *Server) getUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	//get the user
	user, err := s.service.GetUser(r.Context(), id)
	if err != nil {
		log.Printf("Error getting user: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	respondJSON(w, http.StatusOK, user)
}

// Create a new User
func (s *Server) createUser(w http.ResponseWriter, r *http.Request) {
	//Parse and check body
	var req domain.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if req.Name == "" || req.Email == "" {
		respondError(w, http.StatusBadRequest, "Name and Email are required")
		return
	}

	//create user
	id, err := s.service.CreateUser(r.Context(), req)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	respondJSON(w, http.StatusCreated, map[string]int{"id": id})
}

// Fun to send Json
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Printf("Error encoding response: %v", err)
		}
	}
}

// fun to send error
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, domain.ErrorResponse{Error: message})
}
