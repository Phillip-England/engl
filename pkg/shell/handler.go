package shell

import (
	"encoding/json"
	"log"
	"net/http"
	"os/exec"

	"github.com/phillip-england/engl/pkg/pathutil"
)

type ExecRequest struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

type ExecResponse struct {
	Output string `json:"output,omitempty"`
	Error  string `json:"error,omitempty"`
}

type ListResponse struct {
	Commands []Command `json:"commands"`
}

// ListHandler returns all available commands
func ListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("HIT: %s", r.URL.Path)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ListResponse{Commands: AllowedCommands})
}

// ExecHandler executes an allowed shell command
func ExecHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ExecRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeExecError(w, "Invalid JSON body")
		return
	}
	defer r.Body.Close()

	if req.Command == "" {
		writeExecError(w, "command is required")
		return
	}

	if !commandAllowed(req.Command) {
		writeExecError(w, "command not allowed: "+req.Command)
		return
	}

	// Validate path arguments
	validatedArgs := make([]string, len(req.Args))
	for i, arg := range req.Args {
		if pathutil.IsPathArg(arg) {
			validPath, err := pathutil.ValidatePath(arg)
			if err != nil {
				writeExecError(w, "access denied for argument '"+arg+"': "+err.Error())
				return
			}
			validatedArgs[i] = validPath
		} else {
			validatedArgs[i] = arg
		}
	}

	log.Printf("HIT: %s | Command: %s %v", r.URL.Path, req.Command, validatedArgs)

	cmd := exec.Command(req.Command, validatedArgs...)
	cmd.Dir = pathutil.GetAllowedRoot()
	output, err := cmd.CombinedOutput()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ExecResponse{
			Output: string(output),
			Error:  err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ExecResponse{Output: string(output)})
}

func writeExecError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(ExecResponse{Error: msg})
}
