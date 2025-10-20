package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type CheckRequest struct {
	SteamID string `json:"steamid"`
}

type CheckResponse struct {
	Exists   bool   `json:"exists"`
	FilePath string `json:"file_path"`
	Error    string `json:"error,omitempty"`
}

type FileContentResponse struct {
	Success bool            `json:"success"`
	Content json.RawMessage `json:"content,omitempty"`
	Error   string          `json:"error,omitempty"`
}

func checkPlayerFile(steamid string) CheckResponse {
	playersDir := `C:\EVRIMA\surv_server\TheIsle\Saved\Databases\Survival\Players`
	playerFile := filepath.Join(playersDir, steamid+".json")

	if _, err := os.Stat(playerFile); os.IsNotExist(err) {
		return CheckResponse{
			Exists:   false,
			FilePath: playerFile,
		}
	} else if err != nil {
		return CheckResponse{
			Exists:   false,
			FilePath: playerFile,
			Error:    err.Error(),
		}
	}

	return CheckResponse{
		Exists:   true,
		FilePath: playerFile,
	}
}

func getPlayerFileContent(steamid string) FileContentResponse {
	playersDir := `C:\EVRIMA\surv_server\TheIsle\Saved\Databases\Survival\Players`
	playerFile := filepath.Join(playersDir, steamid+".json")

	// Проверяем существование файла
	if _, err := os.Stat(playerFile); os.IsNotExist(err) {
		return FileContentResponse{
			Success: false,
			Error:   "File not found",
		}
	} else if err != nil {
		return FileContentResponse{
			Success: false,
			Error:   err.Error(),
		}
	}

	// Читаем файл
	file, err := os.Open(playerFile)
	if err != nil {
		return FileContentResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to open file: %v", err),
		}
	}
	defer file.Close()

	// Читаем содержимое
	content, err := io.ReadAll(file)
	if err != nil {
		return FileContentResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to read file: %v", err),
		}
	}

	// Валидируем JSON (опционально, но рекомендуется)
	var jsonData json.RawMessage
	if err := json.Unmarshal(content, &jsonData); err != nil {
		return FileContentResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid JSON in file: %v", err),
		}
	}

	return FileContentResponse{
		Success: true,
		Content: jsonData,
	}
}

func checkHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

	if r.Method == "OPTIONS" {
		return
	}

	var req CheckRequest

	switch r.Method {
	case "GET":
		steamid := r.URL.Query().Get("steamid")
		if steamid == "" {
			http.Error(w, `{"error": "steamid parameter is required"}`, http.StatusBadRequest)
			return
		}
		req.SteamID = steamid

	case "POST":
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
			return
		}

	default:
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if req.SteamID == "" {
		http.Error(w, `{"error": "steamid is required"}`, http.StatusBadRequest)
		return
	}

	response := checkPlayerFile(req.SteamID)
	json.NewEncoder(w).Encode(response)
}

func fileContentHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

	if r.Method == "OPTIONS" {
		return
	}

	var req CheckRequest

	switch r.Method {
	case "GET":
		steamid := r.URL.Query().Get("steamid")
		if steamid == "" {
			http.Error(w, `{"error": "steamid parameter is required"}`, http.StatusBadRequest)
			return
		}
		req.SteamID = steamid

	case "POST":
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
			return
		}

	default:
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if req.SteamID == "" {
		http.Error(w, `{"error": "steamid is required"}`, http.StatusBadRequest)
		return
	}

	response := getPlayerFileContent(req.SteamID)
	json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/check", checkHandler)
	http.HandleFunc("/file", fileContentHandler)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status": "ok"}`))
	})

	port := ":8080"
	fmt.Printf("Server starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
