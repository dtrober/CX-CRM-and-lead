package server

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
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
	// Pre-parse templates recursively from the templates directory.
	templates, err := parseTemplates(cfg.TemplatesDir)
	if err != nil {
		log.Printf("Warning: Failed to parse templates: %v", err)
	}
	srv := &Server{
		Server: &http.Server{
			Addr:         cfg.ServerAddress,
			Handler:      r,
			ReadTimeout:  cfg.ServerReadTimeout,
			WriteTimeout: cfg.ServerWriteTimeout,
			IdleTimeout:  120 * time.Second,
		},
		service:   svc,
		templates: templates,
		cfg:       cfg,
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
func parseTemplates(dir string) (*template.Template, error) {
	tmpl := template.New("")
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || filepath.Ext(path) != ".html" {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		// Parse the template with its relative path as the name
		_, err = tmpl.New(rel).Parse(string(content))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}

// Template rend helper
func (s *Server) renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	if s.templates == nil {
		http.Error(w, "Templates not available", http.StatusInternalServerError)
		return
	}
	if err := s.templates.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("Error rendering template %s: %v", name, err)
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
	}
}

// handle Page handles

// Home page handler
func (s *Server) homePage(w http.ResponseWriter, r *http.Request) {
	// Render the home page template.
	// The template name is based on the relative path from the templates directory.
	s.renderTemplate(w, filepath.Join("pages", "index.html"), nil)
}

// Users page handler
func (s *Server) usersPage(w http.ResponseWriter, r *http.Request) {
	// Render the users page template.
	s.renderTemplate(w, filepath.Join("pages", "users.html"), nil)
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
