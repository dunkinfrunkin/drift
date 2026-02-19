package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os/exec"
	"runtime"

	"github.com/frankchan/drift/internal/config"
	"github.com/frankchan/drift/internal/database"
	"github.com/frankchan/drift/internal/engine"
)

// Server serves the drift web UI and API.
type Server struct {
	cfg  *config.Config
	addr string
}

// NewServer creates a new UI server.
func NewServer(cfg *config.Config, addr string) *Server {
	return &Server{cfg: cfg, addr: addr}
}

// Start starts the HTTP server.
func (s *Server) Start(openBrowser bool) error {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/v1/info", s.handleInfo)
	mux.HandleFunc("/api/v1/validate", s.handleValidate)

	// Static files
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		return err
	}
	mux.Handle("/", http.FileServer(http.FS(sub)))

	if openBrowser {
		go openURL("http://" + s.addr)
	}

	return http.ListenAndServe(s.addr, mux)
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

func (s *Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	eng, db, err := s.openDB(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer db.Close()

	infos, err := eng.Info(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(infos)
}

func (s *Server) handleValidate(w http.ResponseWriter, r *http.Request) {
	eng, db, err := s.openDB(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer db.Close()

	errors, err := eng.Validate(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":  len(errors) == 0,
		"errors": errors,
	})
}

func openURL(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	}
	if cmd != nil {
		cmd.Run()
	}
}
