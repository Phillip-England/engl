package pathutil

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrPathOutsideRoot = errors.New("path is outside allowed directory")
	ErrInvalidPath     = errors.New("invalid path")
)

var allowedRoot string

func init() {
	cwd, err := os.Getwd()
	if err != nil {
		panic("failed to get working directory: " + err.Error())
	}
	allowedRoot = cwd
}

// GetAllowedRoot returns the allowed root directory
func GetAllowedRoot() string {
	return allowedRoot
}

// SetAllowedRoot sets the allowed root directory (for testing)
func SetAllowedRoot(path string) {
	allowedRoot = path
}

// ValidatePath checks if the given path is within the allowed root directory.
// It resolves the path to absolute, evaluates symlinks, and ensures it doesn't escape.
func ValidatePath(path string) (string, error) {
	if path == "" {
		return "", ErrInvalidPath
	}

	// Make path absolute relative to allowed root
	var absPath string
	if filepath.IsAbs(path) {
		absPath = filepath.Clean(path)
	} else {
		absPath = filepath.Clean(filepath.Join(allowedRoot, path))
	}

	// Evaluate symlinks to prevent symlink escapes
	realPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		// If path doesn't exist yet (for write operations), check parent
		if os.IsNotExist(err) {
			return validateNonExistentPath(absPath)
		}
		return "", err
	}

	// Check if resolved path is within allowed root
	if !isWithinRoot(realPath) {
		return "", ErrPathOutsideRoot
	}

	return realPath, nil
}

// validateNonExistentPath validates a path that doesn't exist yet (for write operations)
func validateNonExistentPath(absPath string) (string, error) {
	// Walk up the path until we find an existing directory
	dir := absPath
	for {
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root without finding existing path
			break
		}
		dir = parent

		info, err := os.Lstat(dir)
		if err == nil {
			// Found existing directory, resolve it
			realDir, err := filepath.EvalSymlinks(dir)
			if err != nil {
				return "", err
			}

			if !isWithinRoot(realDir) {
				return "", ErrPathOutsideRoot
			}

			// Reconstruct the full path with the resolved base
			relPart, _ := filepath.Rel(dir, absPath)
			finalPath := filepath.Join(realDir, relPart)

			if !isWithinRoot(finalPath) {
				return "", ErrPathOutsideRoot
			}

			return finalPath, nil
		} else if !os.IsNotExist(err) {
			return "", err
		}

		_ = info // silence unused warning
	}

	return "", ErrInvalidPath
}

func isWithinRoot(path string) bool {
	// Ensure path starts with allowed root
	relPath, err := filepath.Rel(allowedRoot, path)
	if err != nil {
		return false
	}

	// Check for parent directory traversal
	if strings.HasPrefix(relPath, "..") {
		return false
	}

	return true
}

// IsPathArg checks if a string looks like a path argument (vs a flag)
func IsPathArg(arg string) bool {
	// Flags start with -
	if strings.HasPrefix(arg, "-") {
		return false
	}

	// Looks like a path if it contains path separators or starts with . or /
	if strings.ContainsAny(arg, "/\\") ||
		strings.HasPrefix(arg, ".") ||
		filepath.IsAbs(arg) {
		return true
	}

	// Could be a relative filename - check if it exists or parent exists
	if _, err := os.Stat(arg); err == nil {
		return true
	}
	if _, err := os.Stat(filepath.Dir(arg)); err == nil {
		return true
	}

	return false
}
