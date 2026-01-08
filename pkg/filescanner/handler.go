package filescanner

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/phillip-england/engl/pkg/pathutil"
)

type ListRequest struct {
	Path string `json:"path"`
}

type FileEntry struct {
	Name  string      `json:"name"`
	Path  string      `json:"path"`
	IsDir bool        `json:"is_dir"`
	Files []FileEntry `json:"files,omitempty"`
}

type ListResponse struct {
	Tree  FileEntry `json:"tree"`
	Error string    `json:"error,omitempty"`
}

type ReadRequest struct {
	Path string `json:"path"`
}

type ReadResponse struct {
	Content string `json:"content,omitempty"`
	Error   string `json:"error,omitempty"`
}

type WriteRequest struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type WriteResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type DeleteRequest struct {
	Path string `json:"path"`
}

type DeleteResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

func ListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid JSON body")
		return
	}
	defer r.Body.Close()

	if req.Path == "" {
		writeError(w, "path is required")
		return
	}

	validPath, err := pathutil.ValidatePath(req.Path)
	if err != nil {
		writeError(w, "access denied: "+err.Error())
		return
	}

	log.Printf("HIT: %s | Path: %s", r.URL.Path, validPath)

	tree, err := buildTree(validPath)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ListResponse{Tree: tree})
}

func buildTree(root string) (FileEntry, error) {
	info, err := os.Stat(root)
	if err != nil {
		return FileEntry{}, err
	}

	entry := FileEntry{
		Name:  info.Name(),
		Path:  root,
		IsDir: info.IsDir(),
	}

	if !info.IsDir() {
		return entry, nil
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return FileEntry{}, err
	}

	for _, e := range entries {
		childPath := filepath.Join(root, e.Name())
		child, err := buildTree(childPath)
		if err != nil {
			continue
		}
		entry.Files = append(entry.Files, child)
	}

	return entry, nil
}

func ReadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ReadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeReadError(w, "Invalid JSON body")
		return
	}
	defer r.Body.Close()

	if req.Path == "" {
		writeReadError(w, "path is required")
		return
	}

	validPath, err := pathutil.ValidatePath(req.Path)
	if err != nil {
		writeReadError(w, "access denied: "+err.Error())
		return
	}

	log.Printf("HIT: %s | Path: %s", r.URL.Path, validPath)

	info, err := os.Stat(validPath)
	if err != nil {
		writeReadError(w, err.Error())
		return
	}

	if info.IsDir() {
		writeReadError(w, "path is a directory, not a file")
		return
	}

	content, err := os.ReadFile(validPath)
	if err != nil {
		writeReadError(w, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ReadResponse{Content: string(content)})
}

func WriteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req WriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeWriteError(w, "Invalid JSON body")
		return
	}
	defer r.Body.Close()

	if req.Path == "" {
		writeWriteError(w, "path is required")
		return
	}

	validPath, err := pathutil.ValidatePath(req.Path)
	if err != nil {
		writeWriteError(w, "access denied: "+err.Error())
		return
	}

	log.Printf("HIT: %s | Path: %s", r.URL.Path, validPath)

	dir := filepath.Dir(validPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		writeWriteError(w, err.Error())
		return
	}

	if err := os.WriteFile(validPath, []byte(req.Content), 0644); err != nil {
		writeWriteError(w, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(WriteResponse{Success: true})
}

func writeError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(ListResponse{Error: msg})
}

func writeReadError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(ReadResponse{Error: msg})
}

func writeWriteError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(WriteResponse{Error: msg})
}

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeDeleteError(w, "Invalid JSON body")
		return
	}
	defer r.Body.Close()

	if req.Path == "" {
		writeDeleteError(w, "path is required")
		return
	}

	validPath, err := pathutil.ValidatePath(req.Path)
	if err != nil {
		writeDeleteError(w, "access denied: "+err.Error())
		return
	}

	log.Printf("HIT: %s | Path: %s", r.URL.Path, validPath)

	if _, err := os.Stat(validPath); err != nil {
		writeDeleteError(w, err.Error())
		return
	}

	if err := os.RemoveAll(validPath); err != nil {
		writeDeleteError(w, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(DeleteResponse{Success: true})
}

func writeDeleteError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(DeleteResponse{Error: msg})
}
