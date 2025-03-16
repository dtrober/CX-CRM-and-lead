package server

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
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
	service   *service.Service
	templates *template.Template
	cfg       *config.Config
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
	srv := &Server{
		Server: &http.Server{
			Addr:         cfg.ServerAddress,
			Handler:      r,
			ReadTimeout:  cfg.ServerReadTimeout,
			WriteTimeout: cfg.ServerWriteTimeout,
			IdleTimeout:  120 * time.Second,
		},
		service: svc,
		cfg:     cfg,
	}

	//static file server
	fileServer := http.FileServer(http.Dir(cfg.StaticDir))
	r.Handle("/static/*", http.StripPrefix("/static", fileServer))

	//Frontend Routes
	r.Get("/", srv.homePage)
	r.Get("/users", srv.usersPage)

	//API Routes
	r.Get("/health", srv.healthCheck)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/users", func(r chi.Router) {
			r.Get("/", srv.getUsers)
			r.Post("/", srv.createUser)
			r.Get("/{id}", srv.getUser)
		})
	})
	return srv
}

// parseTemplates recursively parses all .html templates in the specified directory,
// preserving the relative file paths as template names.
func parsePageTemplates(basePath, pagePath string) (*template.Template, error) {
	return template.ParseFiles(basePath, pagePath)
}

// handle Page handles

// Home page handler
func (s *Server) homePage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := parsePageTemplates(
		filepath.Join(s.cfg.TemplatesDir, "base.html"),
		filepath.Join(s.cfg.TemplatesDir, "pages", "index.html"),
	)
	if err != nil {
		log.Printf("Error parsing home templates: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if err := tmpl.ExecuteTemplate(w, "base", nil); err != nil {
		log.Printf("Error executing home template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Users page handler
func (s *Server) usersPage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := parsePageTemplates(
		filepath.Join(s.cfg.TemplatesDir, "base.html"),
		filepath.Join(s.cfg.TemplatesDir, "pages", "users.html"),
	)
	if err != nil {
		log.Printf("Error parsing users templates: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if err := tmpl.ExecuteTemplate(w, "base", nil); err != nil {
		log.Printf("Error executing users template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	}
	respondJSON(w, http.StatusOK, response)
}

// grabs all users
func (s *Server) getUsers(w http.ResponseWriter, r *http.Request) {
	users, err := s.service.GetUsers(r.Context())
	if err != nil {
		log.Printf("Error getting users: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get users")
		return
	}

	respondJSON(w, http.StatusOK, users)
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
