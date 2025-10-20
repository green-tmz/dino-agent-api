package main

import (
	"encoding/json"
	"fmt"
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

func checkPlayerFile(steamid string) CheckResponse {
	playersDir := "/surv_server/TheIsle/Saved/Databases/Survival/Players"
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

func checkHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Поддержка CORS если нужно
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

func main() {
	http.HandleFunc("/check", checkHandler)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status": "ok"}`))
	})

	port := ":8080"
	fmt.Printf("Server starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
