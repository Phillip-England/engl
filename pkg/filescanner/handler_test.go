package filescanner

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/phillip-england/engl/pkg/pathutil"
)

func withAllowedRoot(t *testing.T, root string) func() {
	old := pathutil.GetAllowedRoot()
	pathutil.SetAllowedRoot(root)
	return func() {
		pathutil.SetAllowedRoot(old)
	}
}

func TestListHandler(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	defer withAllowedRoot(t, tmpDir)()

	subDir := filepath.Join(tmpDir, "subdir")
	os.Mkdir(subDir, 0755)
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("hello"), 0644)
	os.WriteFile(filepath.Join(subDir, "file2.txt"), []byte("world"), 0644)

	tests := []struct {
		name       string
		method     string
		body       any
		wantStatus int
		checkResp  func(*testing.T, ListResponse)
	}{
		{
			name:       "valid request",
			method:     http.MethodPost,
			body:       ListRequest{Path: tmpDir},
			wantStatus: http.StatusOK,
			checkResp: func(t *testing.T, resp ListResponse) {
				if resp.Error != "" {
					t.Errorf("unexpected error: %s", resp.Error)
				}
				if resp.Tree.Name != filepath.Base(tmpDir) {
					t.Errorf("got name %s, want %s", resp.Tree.Name, filepath.Base(tmpDir))
				}
				if !resp.Tree.IsDir {
					t.Error("expected root to be a directory")
				}
				if len(resp.Tree.Files) != 2 {
					t.Errorf("got %d files, want 2", len(resp.Tree.Files))
				}
			},
		},
		{
			name:       "missing path",
			method:     http.MethodPost,
			body:       ListRequest{Path: ""},
			wantStatus: http.StatusBadRequest,
			checkResp: func(t *testing.T, resp ListResponse) {
				if resp.Error != "path is required" {
					t.Errorf("got error %q, want %q", resp.Error, "path is required")
				}
			},
		},
		{
			name:       "invalid path",
			method:     http.MethodPost,
			body:       ListRequest{Path: "/nonexistent/path"},
			wantStatus: http.StatusBadRequest,
			checkResp: func(t *testing.T, resp ListResponse) {
				if resp.Error == "" {
					t.Error("expected an error for nonexistent path")
				}
			},
		},
		{
			name:       "wrong method",
			method:     http.MethodGet,
			body:       nil,
			wantStatus: http.StatusMethodNotAllowed,
			checkResp:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body bytes.Buffer
			if tt.body != nil {
				json.NewEncoder(&body).Encode(tt.body)
			}

			req := httptest.NewRequest(tt.method, "/mcp/tool/file_scanner/list", &body)
			rec := httptest.NewRecorder()

			ListHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d", rec.Code, tt.wantStatus)
			}

			if tt.checkResp != nil {
				var resp ListResponse
				json.NewDecoder(rec.Body).Decode(&resp)
				tt.checkResp(t, resp)
			}
		})
	}
}

func TestReadHandler(t *testing.T) {
	tmpDir := t.TempDir()
	defer withAllowedRoot(t, tmpDir)()

	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("hello world"), 0644)

	tests := []struct {
		name       string
		method     string
		body       any
		wantStatus int
		checkResp  func(*testing.T, ReadResponse)
	}{
		{
			name:       "valid request",
			method:     http.MethodPost,
			body:       ReadRequest{Path: testFile},
			wantStatus: http.StatusOK,
			checkResp: func(t *testing.T, resp ReadResponse) {
				if resp.Error != "" {
					t.Errorf("unexpected error: %s", resp.Error)
				}
				if resp.Content != "hello world" {
					t.Errorf("got content %q, want %q", resp.Content, "hello world")
				}
			},
		},
		{
			name:       "missing path",
			method:     http.MethodPost,
			body:       ReadRequest{Path: ""},
			wantStatus: http.StatusBadRequest,
			checkResp: func(t *testing.T, resp ReadResponse) {
				if resp.Error != "path is required" {
					t.Errorf("got error %q, want %q", resp.Error, "path is required")
				}
			},
		},
		{
			name:       "path is directory",
			method:     http.MethodPost,
			body:       ReadRequest{Path: tmpDir},
			wantStatus: http.StatusBadRequest,
			checkResp: func(t *testing.T, resp ReadResponse) {
				if resp.Error != "path is a directory, not a file" {
					t.Errorf("got error %q, want %q", resp.Error, "path is a directory, not a file")
				}
			},
		},
		{
			name:       "nonexistent file",
			method:     http.MethodPost,
			body:       ReadRequest{Path: "/nonexistent/file.txt"},
			wantStatus: http.StatusBadRequest,
			checkResp: func(t *testing.T, resp ReadResponse) {
				if resp.Error == "" {
					t.Error("expected an error for nonexistent file")
				}
			},
		},
		{
			name:       "wrong method",
			method:     http.MethodGet,
			body:       nil,
			wantStatus: http.StatusMethodNotAllowed,
			checkResp:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body bytes.Buffer
			if tt.body != nil {
				json.NewEncoder(&body).Encode(tt.body)
			}

			req := httptest.NewRequest(tt.method, "/mcp/tool/file_scanner/read", &body)
			rec := httptest.NewRecorder()

			ReadHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d", rec.Code, tt.wantStatus)
			}

			if tt.checkResp != nil {
				var resp ReadResponse
				json.NewDecoder(rec.Body).Decode(&resp)
				tt.checkResp(t, resp)
			}
		})
	}
}

func TestWriteHandler(t *testing.T) {
	tmpDir := t.TempDir()
	defer withAllowedRoot(t, tmpDir)()

	tests := []struct {
		name       string
		method     string
		body       any
		wantStatus int
		checkResp  func(*testing.T, WriteResponse)
		verifyFile func(*testing.T)
	}{
		{
			name:       "valid request",
			method:     http.MethodPost,
			body:       WriteRequest{Path: filepath.Join(tmpDir, "new.txt"), Content: "test content"},
			wantStatus: http.StatusOK,
			checkResp: func(t *testing.T, resp WriteResponse) {
				if resp.Error != "" {
					t.Errorf("unexpected error: %s", resp.Error)
				}
				if !resp.Success {
					t.Error("expected success to be true")
				}
			},
			verifyFile: func(t *testing.T) {
				content, err := os.ReadFile(filepath.Join(tmpDir, "new.txt"))
				if err != nil {
					t.Errorf("failed to read file: %v", err)
				}
				if string(content) != "test content" {
					t.Errorf("got content %q, want %q", string(content), "test content")
				}
			},
		},
		{
			name:       "creates parent directories",
			method:     http.MethodPost,
			body:       WriteRequest{Path: filepath.Join(tmpDir, "nested", "dir", "file.txt"), Content: "nested"},
			wantStatus: http.StatusOK,
			checkResp: func(t *testing.T, resp WriteResponse) {
				if !resp.Success {
					t.Error("expected success to be true")
				}
			},
			verifyFile: func(t *testing.T) {
				content, err := os.ReadFile(filepath.Join(tmpDir, "nested", "dir", "file.txt"))
				if err != nil {
					t.Errorf("failed to read file: %v", err)
				}
				if string(content) != "nested" {
					t.Errorf("got content %q, want %q", string(content), "nested")
				}
			},
		},
		{
			name:       "missing path",
			method:     http.MethodPost,
			body:       WriteRequest{Path: "", Content: "test"},
			wantStatus: http.StatusBadRequest,
			checkResp: func(t *testing.T, resp WriteResponse) {
				if resp.Error != "path is required" {
					t.Errorf("got error %q, want %q", resp.Error, "path is required")
				}
			},
		},
		{
			name:       "wrong method",
			method:     http.MethodGet,
			body:       nil,
			wantStatus: http.StatusMethodNotAllowed,
			checkResp:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body bytes.Buffer
			if tt.body != nil {
				json.NewEncoder(&body).Encode(tt.body)
			}

			req := httptest.NewRequest(tt.method, "/mcp/tool/file_scanner/write", &body)
			rec := httptest.NewRecorder()

			WriteHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d", rec.Code, tt.wantStatus)
			}

			if tt.checkResp != nil {
				var resp WriteResponse
				json.NewDecoder(rec.Body).Decode(&resp)
				tt.checkResp(t, resp)
			}

			if tt.verifyFile != nil {
				tt.verifyFile(t)
			}
		})
	}
}

func TestDeleteHandler(t *testing.T) {
	// Create a shared temp dir and set it as allowed root
	tmpDir := t.TempDir()
	defer withAllowedRoot(t, tmpDir)()

	tests := []struct {
		name       string
		method     string
		setup      func(t *testing.T) string
		body       func(path string) any
		wantStatus int
		checkResp  func(*testing.T, DeleteResponse)
		verifyDel  func(*testing.T, string)
	}{
		{
			name:   "delete file",
			method: http.MethodPost,
			setup: func(t *testing.T) string {
				path := filepath.Join(tmpDir, "to_delete.txt")
				os.WriteFile(path, []byte("delete me"), 0644)
				return path
			},
			body:       func(path string) any { return DeleteRequest{Path: path} },
			wantStatus: http.StatusOK,
			checkResp: func(t *testing.T, resp DeleteResponse) {
				if !resp.Success {
					t.Error("expected success to be true")
				}
			},
			verifyDel: func(t *testing.T, path string) {
				if _, err := os.Stat(path); !os.IsNotExist(err) {
					t.Error("file should have been deleted")
				}
			},
		},
		{
			name:   "delete directory",
			method: http.MethodPost,
			setup: func(t *testing.T) string {
				dir := filepath.Join(tmpDir, "subdir")
				os.Mkdir(dir, 0755)
				os.WriteFile(filepath.Join(dir, "file.txt"), []byte("content"), 0644)
				return dir
			},
			body:       func(path string) any { return DeleteRequest{Path: path} },
			wantStatus: http.StatusOK,
			checkResp: func(t *testing.T, resp DeleteResponse) {
				if !resp.Success {
					t.Error("expected success to be true")
				}
			},
			verifyDel: func(t *testing.T, path string) {
				if _, err := os.Stat(path); !os.IsNotExist(err) {
					t.Error("directory should have been deleted")
				}
			},
		},
		{
			name:       "missing path",
			method:     http.MethodPost,
			setup:      func(t *testing.T) string { return "" },
			body:       func(path string) any { return DeleteRequest{Path: ""} },
			wantStatus: http.StatusBadRequest,
			checkResp: func(t *testing.T, resp DeleteResponse) {
				if resp.Error != "path is required" {
					t.Errorf("got error %q, want %q", resp.Error, "path is required")
				}
			},
		},
		{
			name:   "nonexistent path",
			method: http.MethodPost,
			setup: func(t *testing.T) string {
				return filepath.Join(tmpDir, "nonexistent", "file.txt")
			},
			body:       func(path string) any { return DeleteRequest{Path: path} },
			wantStatus: http.StatusBadRequest,
			checkResp: func(t *testing.T, resp DeleteResponse) {
				if resp.Error == "" {
					t.Error("expected an error for nonexistent path")
				}
			},
		},
		{
			name:       "wrong method",
			method:     http.MethodGet,
			setup:      func(t *testing.T) string { return "" },
			body:       func(path string) any { return nil },
			wantStatus: http.StatusMethodNotAllowed,
			checkResp:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup(t)

			var body bytes.Buffer
			if b := tt.body(path); b != nil {
				json.NewEncoder(&body).Encode(b)
			}

			req := httptest.NewRequest(tt.method, "/mcp/tool/file_scanner/delete", &body)
			rec := httptest.NewRecorder()

			DeleteHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d", rec.Code, tt.wantStatus)
			}

			if tt.checkResp != nil {
				var resp DeleteResponse
				json.NewDecoder(rec.Body).Decode(&resp)
				tt.checkResp(t, resp)
			}

			if tt.verifyDel != nil {
				tt.verifyDel(t, path)
			}
		})
	}
}
