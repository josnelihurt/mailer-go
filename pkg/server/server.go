package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/josnelihurt/mailer-go/pkg/config"
	"github.com/josnelihurt/mailer-go/pkg/mailer"
)

type Server struct {
	config config.Config
	apiKey string
}

func NewServer(cfg config.Config) *Server {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("API_KEY environment variable is required in server mode")
	}

	return &Server{
		config: cfg,
		apiKey: apiKey,
	}
}

func (s *Server) Start() error {
	http.HandleFunc("/v1/sms/enqueue", s.authMiddleware(s.handleEnqueue))
	http.HandleFunc("/health", s.handleHealth)

	addr := ":8080"
	log.Printf("Server listening on %s", addr)
	return http.ListenAndServe(addr, nil)
}

func (s *Server) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey != s.apiKey {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			log.Printf("Unauthorized access attempt from %s", r.RemoteAddr)
			return
		}
		next(w, r)
	}
}

func (s *Server) handleEnqueue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		log.Printf("Failed to read body: %v", err)
		return
	}
	defer r.Body.Close()

	var req mailer.SMSEnqueueRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		log.Printf("Invalid JSON: %v", err)
		return
	}

	if req.FolderName == "" {
		http.Error(w, "Missing folder_name", http.StatusBadRequest)
		return
	}

	// Push to Redis pub/sub channel
	mailer.PushToRedis(s.config, req.FolderName, req.SMSMessage)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))

	log.Printf("SMS enqueued successfully from %s to queue sms:%s", r.RemoteAddr, req.FolderName)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}
