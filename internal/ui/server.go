package ui

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/frankchan/drift/internal/config"
	"github.com/frankchan/drift/internal/database"
	"github.com/frankchan/drift/internal/diff"
	"github.com/frankchan/drift/internal/engine"
	"github.com/frankchan/drift/internal/lint"
)

// Server serves the drift web UI and API.
type Server struct {
	cfg  *config.Config
	addr string

	// WebSocket subscribers for real-time events
	mu          sync.RWMutex
	subscribers map[chan Event]struct{}
}

// Event is a real-time event sent to WebSocket clients.
type Event struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// NewServer creates a new UI server.
func NewServer(cfg *config.Config, addr string) *Server {
	return &Server{
		cfg:         cfg,
		addr:        addr,
		subscribers: make(map[chan Event]struct{}),
	}
}

// Start starts the HTTP server.
func (s *Server) Start(openBrowser bool) error {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/v1/info", s.cors(s.handleInfo))
	mux.HandleFunc("/api/v1/validate", s.cors(s.handleValidate))
	mux.HandleFunc("/api/v1/migrate", s.cors(s.handleMigrate))
	mux.HandleFunc("/api/v1/undo", s.cors(s.handleUndo))
	mux.HandleFunc("/api/v1/repair", s.cors(s.handleRepair))
	mux.HandleFunc("/api/v1/clean", s.cors(s.handleClean))
	mux.HandleFunc("/api/v1/baseline", s.cors(s.handleBaseline))
	mux.HandleFunc("/api/v1/diff", s.cors(s.handleDiff))
	mux.HandleFunc("/api/v1/snapshot", s.cors(s.handleSnapshot))
	mux.HandleFunc("/api/v1/lint", s.cors(s.handleLint))
	mux.HandleFunc("/api/v1/config", s.cors(s.handleConfig))
	mux.HandleFunc("/api/v1/events", s.handleSSE)

	// Static files (SPA with fallback to index.html)
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		return err
	}
	mux.Handle("/", spaHandler{fs: http.FS(sub)})

	if openBrowser {
		go func() {
			time.Sleep(200 * time.Millisecond)
			openURL("http://" + s.addr)
		}()
	}

	fmt.Printf("drift UI running at http://%s\n", s.addr)
	fmt.Printf("Press Ctrl+C to stop\n")
	return http.ListenAndServe(s.addr, mux)
}

// spaHandler serves static files with SPA fallback to index.html.
type spaHandler struct {
	fs http.FileSystem
}

func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	// Try to serve the file
	f, err := h.fs.Open(path)
	if err != nil {
		// Fallback to index.html for SPA routing
		r.URL.Path = "/"
	} else {
		f.Close()
	}
	http.FileServer(h.fs).ServeHTTP(w, r)
}

func (s *Server) cors(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}
		next(w, r)
	}
}

func (s *Server) openDB(ctx context.Context) (*engine.Engine, database.Database, error) {
	driver := config.DetectDriver(s.cfg.URL)
	if driver == "" {
		return nil, nil, fmt.Errorf("cannot detect driver")
	}
	db, err := database.Open(ctx, driver, s.cfg.URL)
	if err != nil {
		return nil, nil, err
	}
	eng := engine.New(s.cfg, db)
	return eng, db, nil
}

func (s *Server) jsonReply(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (s *Server) jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// --- Handlers ---

func (s *Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	eng, db, err := s.openDB(r.Context())
	if err != nil {
		s.jsonError(w, err.Error(), 500)
		return
	}
	defer db.Close()

	infos, err := eng.Info(r.Context())
	if err != nil {
		s.jsonError(w, err.Error(), 500)
		return
	}

	s.jsonReply(w, infos)
}

func (s *Server) handleValidate(w http.ResponseWriter, r *http.Request) {
	eng, db, err := s.openDB(r.Context())
	if err != nil {
		s.jsonError(w, err.Error(), 500)
		return
	}
	defer db.Close()

	errors, err := eng.Validate(r.Context())
	if err != nil {
		s.jsonError(w, err.Error(), 500)
		return
	}

	s.jsonReply(w, map[string]interface{}{
		"valid":  len(errors) == 0,
		"errors": errors,
	})
}

func (s *Server) handleMigrate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		s.jsonError(w, "method not allowed", 405)
		return
	}

	eng, db, err := s.openDB(r.Context())
	if err != nil {
		s.jsonError(w, err.Error(), 500)
		return
	}
	defer db.Close()

	dryRun := r.URL.Query().Get("dryRun") == "true"

	var buf bytes.Buffer
	eng.SetOutput(&buf)

	s.broadcast(Event{Type: "migrate_start", Payload: map[string]bool{"dryRun": dryRun}})

	results, err := eng.Migrate(r.Context(), engine.PlanOptions{DryRun: dryRun})

	s.broadcast(Event{Type: "migrate_end", Payload: map[string]interface{}{
		"success": err == nil,
		"count":   len(results),
	}})

	if err != nil {
		s.jsonError(w, err.Error(), 500)
		return
	}

	s.jsonReply(w, map[string]interface{}{
		"results": results,
		"output":  buf.String(),
	})
}

func (s *Server) handleUndo(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		s.jsonError(w, "method not allowed", 405)
		return
	}

	eng, db, err := s.openDB(r.Context())
	if err != nil {
		s.jsonError(w, err.Error(), 500)
		return
	}
	defer db.Close()

	count := 1
	if c := r.URL.Query().Get("count"); c != "" {
		count, _ = strconv.Atoi(c)
	}
	dryRun := r.URL.Query().Get("dryRun") == "true"

	var buf bytes.Buffer
	eng.SetOutput(&buf)

	results, err := eng.Undo(r.Context(), count, "", dryRun)
	if err != nil {
		s.jsonError(w, err.Error(), 500)
		return
	}

	s.broadcast(Event{Type: "undo_complete", Payload: map[string]int{"count": len(results)}})

	s.jsonReply(w, map[string]interface{}{
		"results": results,
		"output":  buf.String(),
	})
}

func (s *Server) handleRepair(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		s.jsonError(w, "method not allowed", 405)
		return
	}

	eng, db, err := s.openDB(r.Context())
	if err != nil {
		s.jsonError(w, err.Error(), 500)
		return
	}
	defer db.Close()

	var buf bytes.Buffer
	eng.SetOutput(&buf)

	if err := eng.Repair(r.Context()); err != nil {
		s.jsonError(w, err.Error(), 500)
		return
	}

	s.jsonReply(w, map[string]string{"output": buf.String()})
}

func (s *Server) handleClean(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		s.jsonError(w, "method not allowed", 405)
		return
	}

	eng, db, err := s.openDB(r.Context())
	if err != nil {
		s.jsonError(w, err.Error(), 500)
		return
	}
	defer db.Close()

	var buf bytes.Buffer
	eng.SetOutput(&buf)

	if err := eng.Clean(r.Context()); err != nil {
		s.jsonError(w, err.Error(), 500)
		return
	}

	s.broadcast(Event{Type: "clean_complete", Payload: nil})

	s.jsonReply(w, map[string]string{"output": buf.String()})
}

func (s *Server) handleBaseline(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		s.jsonError(w, "method not allowed", 405)
		return
	}

	eng, db, err := s.openDB(r.Context())
	if err != nil {
		s.jsonError(w, err.Error(), 500)
		return
	}
	defer db.Close()

	version := r.URL.Query().Get("version")
	if version == "" {
		version = "001"
	}

	var buf bytes.Buffer
	eng.SetOutput(&buf)

	if err := eng.Baseline(r.Context(), version); err != nil {
		s.jsonError(w, err.Error(), 500)
		return
	}

	s.jsonReply(w, map[string]string{"output": buf.String()})
}

func (s *Server) handleDiff(w http.ResponseWriter, r *http.Request) {
	_, db, err := s.openDB(r.Context())
	if err != nil {
		s.jsonError(w, err.Error(), 500)
		return
	}
	defer db.Close()

	current, err := diff.CaptureSchema(r.Context(), db)
	if err != nil {
		s.jsonError(w, err.Error(), 500)
		return
	}

	// Compare against empty if no snapshot provided
	var from *diff.SchemaSnapshot
	snapshotPath := r.URL.Query().Get("snapshot")
	if snapshotPath != "" {
		from, err = diff.LoadSnapshot(snapshotPath)
		if err != nil {
			s.jsonError(w, err.Error(), 400)
			return
		}
	} else {
		from = &diff.SchemaSnapshot{}
	}

	changes := diff.Compare(from, current)
	s.jsonReply(w, changes)
}

func (s *Server) handleSnapshot(w http.ResponseWriter, r *http.Request) {
	_, db, err := s.openDB(r.Context())
	if err != nil {
		s.jsonError(w, err.Error(), 500)
		return
	}
	defer db.Close()

	snap, err := diff.CaptureSchema(r.Context(), db)
	if err != nil {
		s.jsonError(w, err.Error(), 500)
		return
	}

	s.jsonReply(w, snap)
}

func (s *Server) handleLint(w http.ResponseWriter, r *http.Request) {
	linter := lint.NewLinter(nil)
	results, err := linter.LintLocations(s.cfg.Locations)
	if err != nil {
		s.jsonError(w, err.Error(), 500)
		return
	}

	s.jsonReply(w, map[string]interface{}{
		"results": results,
		"clean":   len(results) == 0,
	})
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	// Return sanitized config (no password in URL)
	sanitized := map[string]interface{}{
		"driver":    config.DetectDriver(s.cfg.URL),
		"locations": s.cfg.Locations,
		"table":     s.cfg.Table,
		"schemas":   s.cfg.Schemas,
	}
	s.jsonReply(w, sanitized)
}

// --- SSE (Server-Sent Events) for real-time updates ---

func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", 500)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ch := make(chan Event, 16)
	s.mu.Lock()
	s.subscribers[ch] = struct{}{}
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.subscribers, ch)
		s.mu.Unlock()
	}()

	// Send initial ping
	fmt.Fprintf(w, "event: connected\ndata: {}\n\n")
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			return
		case evt := <-ch:
			data, _ := json.Marshal(evt)
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", evt.Type, data)
			flusher.Flush()
		}
	}
}

func (s *Server) broadcast(evt Event) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for ch := range s.subscribers {
		select {
		case ch <- evt:
		default:
			// Drop if subscriber is slow
		}
	}
}

func openURL(url string) {
	var args []string
	switch runtime.GOOS {
	case "darwin":
		args = []string{"open", url}
	case "linux":
		args = []string{"xdg-open", url}
	case "windows":
		args = []string{"rundll32", "url.dll,FileProtocolHandler", url}
	default:
		return
	}
	cmd := exec.Command(args[0], args[1:]...)
	_ = cmd.Run()
}

// SanitizeURL removes password from database URL for display.
func SanitizeURL(url string) string {
	// Simple: hide anything between :// and @
	if idx := strings.Index(url, "://"); idx >= 0 {
		rest := url[idx+3:]
		if atIdx := strings.Index(rest, "@"); atIdx >= 0 {
			return url[:idx+3] + "***@" + rest[atIdx+1:]
		}
	}
	return url
}
