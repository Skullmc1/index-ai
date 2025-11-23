# Index-AI

Index-AI is a command-line tool written in Go that automatically organizes files and folders in a directory into categorized subdirectories. It offers two modes of operation: a "Normal" mode that uses file extensions and heuristic-based scanning, and an "AI" mode that leverages the Gemini API for more intelligent, context-aware organization.

## Features

- **Two Organization Modes:**
    - **Normal Mode:** Categorizes files based on their extensions (e.g., `.jpg` -> Images, `.pdf` -> Documents). It can also identify games and software folders by inspecting their contents and performing Google searches for unknown items.
    - **AI Mode:** Uses the Gemini 2.0 Flash model to analyze a list of file and folder names and determine the best categories for them. It can handle more ambiguous names and can perform follow-up web searches for clarification.
- **Interactive CLI:** A user-friendly interface built with `bubbletea` guides you through the process.
- **Safe Moves:** Creates category folders if they don't exist and handles potential file name conflicts by renaming duplicates.

## How It Works

### Normal Mode

1.  **Extension Mapping:** Files are first sorted based on a predefined map of common extensions (e.g., `.mp4` is moved to a "Video" folder).
2.  **Folder Scanning:** If an item is a directory, the tool scans its contents for clues to determine if it's a game, software, or something else.
3.  **Web Context:** For items that cannot be categorized, the tool performs a Google search on the name and analyzes the search result title to guess the category (e.g., "game", "movie", "software").

### AI Mode

1.  **API Call:** The list of files and folders (up to 30 items) is sent to the Gemini API with a prompt asking it to categorize them.
2.  **JSON Response:** The AI returns a structured JSON response detailing which folder each item should be moved to. It can also flag items that require a web search.
3.  **Web Search Follow-up:** If the AI requests a web search for a cryptic item, the tool performs the search and sends a second request to the AI with the search result, asking for a final categorization.
4.  **File Movement:** The tool executes the moves described in the AI's response.

## Prerequisites

- **Go:** You need to have Go installed to run the application.
- **Gemini API Key:** To use the AI mode, you must have an API key for the Gemini API.

## How to Run

1.  **Clone the repository or download the source code.**

2.  **Navigate to the project directory:**
    ```sh
    cd index-ai
    ```

3.  **Run the application:**
    ```sh
    go run .
    ```
    *(On Windows, you can also use `go run main.go AIRoute.go NormalRoute.go`)*

4.  **Follow the on-screen prompts:**
    - Choose whether to use AI mode.
    - If using AI mode, enter your Gemini API key.
    - Specify the folder path you want to organize (or press Enter to use the current directory).

The tool will then process the files and display a completion message.
