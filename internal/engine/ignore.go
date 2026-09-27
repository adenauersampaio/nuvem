package engine

import "strings"

// IsIgnoredDir returns true if a directory should be excluded from sync.
func IsIgnoredDir(name string) bool {
	return strings.HasPrefix(name, ".")
}

// IsIgnoredFile returns true if a file is a hidden, editor backup, lock, or transient file.
func IsIgnoredFile(name string) bool {
	// Hidden files (e.g. .git, .goutputstream, .nuvem)
	if strings.HasPrefix(name, ".") {
		return true
	}
	// Backup or temporary editor files
	if strings.HasSuffix(name, "~") || strings.HasSuffix(name, "#") {
		return true
	}
	// Common lock and temporary patterns
	lower := strings.ToLower(name)
	if strings.Contains(lower, ".~lock.") || strings.HasPrefix(lower, "~$") {
		return true
	}
	if strings.HasSuffix(lower, ".tmp") ||
		strings.HasSuffix(lower, ".swp") ||
		strings.HasSuffix(lower, ".swo") ||
		strings.HasSuffix(lower, ".bak") ||
		strings.HasSuffix(lower, ".crdownload") ||
		strings.HasSuffix(lower, ".part") {
		return true
	}
	return false
}
