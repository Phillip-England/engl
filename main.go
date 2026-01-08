package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/phillip-england/engl/pkg/filescanner"
	"github.com/phillip-england/engl/pkg/pathutil"
	"github.com/phillip-england/engl/pkg/shell"
)

type Endpoint struct {
	Path        string `json:"path"`
	Method      string `json:"method"`
	Description string `json:"description"`
}

type IndexResponse struct {
	Name        string     `json:"name"`
	Version     string     `json:"version"`
	AllowedRoot string     `json:"allowed_root"`
	Endpoints   []Endpoint `json:"endpoints"`
}

var endpoints = []Endpoint{
	{Path: "/", Method: "GET", Description: "This index - lists all available endpoints"},
	{Path: "/mcp/tool/file_scanner/list", Method: "POST", Description: "List directory contents as a tree structure"},
	{Path: "/mcp/tool/file_scanner/read", Method: "POST", Description: "Read file contents"},
	{Path: "/mcp/tool/file_scanner/write", Method: "POST", Description: "Write content to a file"},
	{Path: "/mcp/tool/file_scanner/delete", Method: "POST", Description: "Delete a file or directory"},
	{Path: "/mcp/tool/shell/list", Method: "GET", Description: "List available shell commands"},
	{Path: "/mcp/tool/shell/exec", Method: "POST", Description: "Execute a whitelisted shell command"},
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(IndexResponse{
		Name:        "MCP File Scanner Server",
		Version:     "1.0.0",
		AllowedRoot: pathutil.GetAllowedRoot(),
		Endpoints:   endpoints,
	})
}

func cors(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func main() {
	http.HandleFunc("/", cors(indexHandler))
	http.HandleFunc("/mcp/tool/file_scanner/list", cors(filescanner.ListHandler))
	http.HandleFunc("/mcp/tool/file_scanner/read", cors(filescanner.ReadHandler))
	http.HandleFunc("/mcp/tool/file_scanner/write", cors(filescanner.WriteHandler))
	http.HandleFunc("/mcp/tool/file_scanner/delete", cors(filescanner.DeleteHandler))
	http.HandleFunc("/mcp/tool/shell/list", cors(shell.ListHandler))
	http.HandleFunc("/mcp/tool/shell/exec", cors(shell.ExecHandler))

	port := ":8080"
	log.Printf("MCP Server listening on %s...", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
