package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type AIFileAction struct {
	File        string `json:"file"`
	Destination string `json:"destination"`
}

type AIResponse struct {
	Moves         []AIFileAction `json:"moves"`
	NeedWebSearch []string       `json:"need_websearch"`
}

type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
}

type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

type GeminiPart struct {
	Text string `json:"text"`
}

type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func runAIRoute(path string, apiKey string) tea.Cmd {
	return func() tea.Msg {
		files, err := os.ReadDir(path)
		if err != nil {
			return progressMsg(fmt.Sprintf("Error reading dir: %v", err))
		}

		var items []string
		for _, f := range files {
			name := f.Name()
			if !strings.HasPrefix(name, ".") && name != "main.exe" && name != "main" {
				items = append(items, name)
			}
		}

		if len(items) > 30 {
			return doneMsg(fmt.Sprintf("Too many items (%d). Limit is 30. Try again in normal mode.", len(items)))
		}

		if len(items) == 0 {
			return doneMsg("No items found to organize.")
		}

		prompt := fmt.Sprintf(`Organize these messy files and folders into clean category folders.
		Items: %s.

		Rules:
		1. Group files and folders by context (e.g., 'Trip_Photos' folder -> 'Images' category).
		2. Do NOT move a folder into itself.
		3. If an item name is cryptic, mark it for web search.

		Return ONLY raw JSON in this format:
		{"moves": [{"file": "item_name", "destination": "CategoryFolder"}], "need_websearch": ["cryptic_name"]}`,
			strings.Join(items, ", "))

		aiRespText, err := callGemini(apiKey, prompt)
		if err != nil {
			return progressMsg(fmt.Sprintf("AI API Error: %v", err))
		}

		var plan AIResponse
		cleanJSON := extractJSON(aiRespText)
		err = json.Unmarshal([]byte(cleanJSON), &plan)
		if err != nil {
			return progressMsg("Failed to parse AI response")
		}

		for _, move := range plan.Moves {
			err := performMove(path, move.File, move.Destination)
			if err != nil {
				fmt.Sprintf("Failed to move %s: %v", move.File, err)
			}
		}

		if len(plan.NeedWebSearch) > 0 {
			for _, unknownItem := range plan.NeedWebSearch {
				title := performGoogleSearch(unknownItem)

				followUpPrompt := fmt.Sprintf(`I have an item named "%s".
				Google search result says: "%s".
				Based on this, which existing folder from previous list should it go to?
				Return ONLY raw JSON: {"moves": [{"file": "%s", "destination": "folder"}], "need_websearch": []}`,
					unknownItem, title, unknownItem)

				followUpText, err := callGemini(apiKey, followUpPrompt)
				if err == nil {
					var followUpPlan AIResponse
					cleanFollowUp := extractJSON(followUpText)
					json.Unmarshal([]byte(cleanFollowUp), &followUpPlan)

					for _, move := range followUpPlan.Moves {
						performMove(path, move.File, move.Destination)
					}
				}
			}
		}

		return doneMsg("AI Organization Complete")
	}
}

func callGemini(key, prompt string) (string, error) {
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key=" + key

	reqBody := GeminiRequest{
		Contents: []GeminiContent{{Parts: []GeminiPart{{Text: prompt}}}},
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var gResp GeminiResponse
	json.Unmarshal(body, &gResp)

	if len(gResp.Candidates) > 0 && len(gResp.Candidates[0].Content.Parts) > 0 {
		return gResp.Candidates[0].Content.Parts[0].Text, nil
	}
	return "", fmt.Errorf("no response")
}

func performMove(basePath, itemName, destFolder string) error {
	destPath := filepath.Join(basePath, destFolder)
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		os.Mkdir(destPath, 0755)
	}

	oldPath := filepath.Join(basePath, itemName)
	newPath := filepath.Join(destPath, itemName)

	if oldPath == destPath {
		return nil
	}

	if _, err := os.Stat(newPath); err == nil {
		ext := filepath.Ext(itemName)
		nameOnly := strings.TrimSuffix(itemName, ext)
		newPath = filepath.Join(destPath, fmt.Sprintf("%s_duplicate%s", nameOnly, ext))
	}

	return os.Rename(oldPath, newPath)
}

func performGoogleSearch(query string) string {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://www.google.com/search?q="+query, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return "Unknown"
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	src := string(body)

	start := strings.Index(src, "<title>")
	end := strings.Index(src, "</title>")

	if start != -1 && end != -1 {
		clean := src[start+7 : end]
		clean = strings.ReplaceAll(clean, " - Google Search", "")
		return clean
	}
	return "Unknown Context"
}

func extractJSON(str string) string {
	start := strings.Index(str, "{")
	end := strings.LastIndex(str, "}")
	if start == -1 || end == -1 {
		return "{}"
	}
	return str[start : end+1]
}
