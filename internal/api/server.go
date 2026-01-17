package api

import (
	"encoding/json"
	"net/http"

	"github.com/nerdneilsfield/dumper/internal/llm"
	"github.com/nerdneilsfield/dumper/internal/store"
)

type Server struct {
	stores    *store.Manager
	botToken  string
	llmClient *llm.Client
	mux       *http.ServeMux
}

func NewServer(stores *store.Manager, botToken string, llmClient *llm.Client) *Server {
	s := &Server{
		stores:    stores,
		botToken:  botToken,
		llmClient: llmClient,
		mux:       http.NewServeMux(),
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	// API routes (protected)
	api := http.NewServeMux()
	api.HandleFunc("GET /items", s.handleListItems)
	api.HandleFunc("GET /items/{id}", s.handleGetItem)
	api.HandleFunc("DELETE /items/{id}", s.handleDeleteItem)
	api.HandleFunc("GET /search", s.handleSearch)
	api.HandleFunc("GET /tags", s.handleGetTags)
	api.HandleFunc("GET /graph", s.handleGetGraph)
	api.HandleFunc("POST /ask", s.handleAsk)
	api.HandleFunc("GET /export", s.handleExport)
	api.HandleFunc("GET /stats", s.handleStats)

	s.mux.Handle("/api/", http.StripPrefix("/api", s.authMiddleware(api)))

	// Health check
	s.mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Static files for Mini App (if exists)
	s.mux.Handle("/", http.FileServer(http.Dir("mini-app/dist")))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// CORS headers for Mini App
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Telegram-Init-Data")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	s.mux.ServeHTTP(w, r)
}

func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
