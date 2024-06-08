package utils

import (
	"os"
	"path/filepath"
)

// GetDataDirectory creates and retrieves the whoip data directory
func GetDataDirectory() string {
	createAndCheckDirectory := func(directory string) bool {
		// Try to create the directory if it doesn't exist
		if err := os.MkdirAll(directory, 0755); err != nil {
			return false
		}
		// Check if the directory is readable and writable
		info, err := os.Stat(directory)
		if err != nil {
			return false
		}
		mode := info.Mode()
		if mode&(1<<(uint(7))) != 0 && mode&(1<<(uint(7)-2)) != 0 {
			return true
		}
		return false
	}

	var directory string

	xdgDataHome := os.Getenv("XDG_DATA_HOME")
	if xdgDataHome != "" {
		directory = filepath.Join(xdgDataHome, "whoip")
		if createAndCheckDirectory(directory) {
			return directory
		}
	}

	home := os.Getenv("HOME")
	if home != "" {
		directory = filepath.Join(home, ".local", "share", "whoip")
		if createAndCheckDirectory(directory) {
			return directory
		}
	}

	directory = filepath.Join("/tmp", "whoip")
	if createAndCheckDirectory(directory) {
		return directory
	}

	panic("Failed to create data directory on '$XDG_DATA_HOME/whoip', '$HOME/.local/share/whoip' or '/tmp/whoip'.")
}
