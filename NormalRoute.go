package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

var extensionMap = map[string]string{
	// Images
	".jpg": "Images", ".jpeg": "Images", ".png": "Images", ".gif": "Images", ".webp": "Images", ".svg": "Images",
	// Documents
	".pdf": "Documents", ".doc": "Documents", ".docx": "Documents", ".txt": "Documents", ".xlsx": "Documents", ".ppt": "Documents",
	// Audio
	".mp3": "Audio", ".wav": "Audio", ".flac": "Audio", ".ogg": "Audio", ".m4a": "Audio",
	// Video
	".mp4": "Video", ".mkv": "Video", ".avi": "Video", ".mov": "Video", ".webm": "Video",
	// Archives
	".zip": "Archives", ".rar": "Archives", ".7z": "Archives", ".tar": "Archives", ".gz": "Archives",
	// Executables
	".exe": "Installers", ".msi": "Installers", ".bat": "Scripts", ".sh": "Scripts",
	// Code
	".go": "Code", ".py": "Code", ".js": "Code", ".html": "Code", ".css": "Code", ".json": "Code",
}

func runNormalRoute(path string) tea.Cmd {
	return func() tea.Msg {
		files, err := os.ReadDir(path)
		if err != nil {
			return progressMsg(fmt.Sprintf("Error reading dir: %v", err))
		}

		var movedCount int
		var searchCount int

		for _, f := range files {
			name := f.Name()

			if strings.HasPrefix(name, ".") || name == "main.exe" || name == "main" || f.IsDir() {
				continue
			}

			ext := strings.ToLower(filepath.Ext(name))
			destFolder := "Misc" 
			found := false

			if folder, exists := extensionMap[ext]; exists {
				destFolder = folder
				found = true
			}

			if !found {
				searchCount++
				time.Sleep(1500 * time.Millisecond)

				title := strings.ToLower(performGoogleSearch(name))

				if strings.Contains(title, "software") || strings.Contains(title, "download") || strings.Contains(title, "install") {
					destFolder = "Installers"
				} else if strings.Contains(title, "image") || strings.Contains(title, "photo") || strings.Contains(title, "picture") {
					destFolder = "Images"
				} else if strings.Contains(title, "music") || strings.Contains(title, "song") || strings.Contains(title, "audio") {
					destFolder = "Audio"
				} else if strings.Contains(title, "video") || strings.Contains(title, "movie") || strings.Contains(title, "film") {
					destFolder = "Video"
				} else if strings.Contains(title, "driver") {
					destFolder = "Drivers"
				} else if strings.Contains(title, "tutorial") || strings.Contains(title, "guide") || strings.Contains(title, "manual") {
					destFolder = "Documents"
				}
			}

			err := performMove(path, name, destFolder)
			if err == nil {
				movedCount++
			}
		}

		return doneMsg(fmt.Sprintf("Organization Complete. Moved %d files. Performed %d web searches.", movedCount, searchCount))
	}
}
