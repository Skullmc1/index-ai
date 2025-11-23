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
	".jpg": "Images", ".jpeg": "Images", ".png": "Images", ".gif": "Images", ".webp": "Images",
	".pdf": "Documents", ".doc": "Documents", ".docx": "Documents", ".txt": "Documents",
	".mp3": "Audio", ".wav": "Audio", ".flac": "Audio",
	".mp4": "Video", ".mkv": "Video", ".avi": "Video",
	".zip": "Archives", ".rar": "Archives", ".7z": "Archives",
	".exe": "Installers", ".msi": "Installers",
	".go": "Code", ".py": "Code", ".js": "Code", ".html": "Code", ".css": "Code",
}

func runNormalRoute(path string) tea.Cmd {
	return func() tea.Msg {
		entries, err := os.ReadDir(path)
		if err != nil {
			return progressMsg(fmt.Sprintf("Error reading dir: %v", err))
		}

		var movedCount int
		var searchCount int

		for _, entry := range entries {
			name := entry.Name()

			if strings.HasPrefix(name, ".") || name == "main.exe" || name == "main" {
				continue
			}

			category := "Misc"
			resolved := false

			if entry.IsDir() {
				category, resolved = scanFolderContents(filepath.Join(path, name))
			} else {
				ext := strings.ToLower(filepath.Ext(name))
				if val, ok := extensionMap[ext]; ok {
					category = val
					resolved = true
				}
			}

			if !resolved {
				searchCount++
				time.Sleep(1200 * time.Millisecond)
				category = scanWebContext(name)
			}

			if category != "Misc" {
				err := performMove(path, name, category)
				if err == nil {
					movedCount++
				}
			}
		}

		return doneMsg(fmt.Sprintf("Normal Scan Complete. Organized %d items. Performed %d web searches.", movedCount, searchCount))
	}
}

func scanFolderContents(dirPath string) (string, bool) {
	items, err := os.ReadDir(dirPath)
	if err != nil {
		return "Misc", false
	}

	gameScore := 0
	softScore := 0

	for _, item := range items {
		lower := strings.ToLower(item.Name())

		if strings.HasSuffix(lower, "steam_api.dll") || strings.HasSuffix(lower, "unityplayer.dll") || strings.Contains(lower, "shader") || strings.Contains(lower, "save") {
			gameScore += 5
		}
		if lower == "levels" || lower == "maps" || lower == "mods" || lower == "textures" {
			gameScore += 3
		}

		if strings.HasPrefix(lower, "readme") || strings.HasPrefix(lower, "license") || strings.HasPrefix(lower, "changes") {
			softScore += 1
		}
		if lower == "src" || lower == "bin" || lower == "lib" || lower == "include" {
			softScore += 3
		}
		if strings.Contains(lower, "setup") || strings.Contains(lower, "installer") {
			softScore += 2
		}
	}

	if gameScore > 0 && gameScore >= softScore {
		return "Games", true
	}
	if softScore > 4 {
		return "Software", true
	}

	return "Misc", false
}

func scanWebContext(query string) string {
	title := strings.ToLower(performGoogleSearch(query))

	if strings.Contains(title, "game") || strings.Contains(title, "steam") || strings.Contains(title, "rpg") || strings.Contains(title, "fps") || strings.Contains(title, "multiplayer") || strings.Contains(title, "walkthrough") || strings.Contains(title, "metacritic") || strings.Contains(title, "ign") {
		return "Games"
	}

	if strings.Contains(title, "software") || strings.Contains(title, "download") || strings.Contains(title, "tool") || strings.Contains(title, "utility") || strings.Contains(title, "github") || strings.Contains(title, "version") || strings.Contains(title, "open source") {
		return "Software"
	}

	if strings.Contains(title, "movie") || strings.Contains(title, "film") || strings.Contains(title, "imdb") || strings.Contains(title, "rotten tomatoes") || strings.Contains(title, "cast") {
		return "Movies"
	}

	if strings.Contains(title, "album") || strings.Contains(title, "song") || strings.Contains(title, "lyrics") || strings.Contains(title, "spotify") || strings.Contains(title, "band") {
		return "Music"
	}

	return "Misc"
}
