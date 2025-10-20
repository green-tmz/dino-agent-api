package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	_ "strings"
	"time"
)

type CheckRequest struct {
	SteamID   string `json:"steamid"`
	OldSlotID string `json:"old_slot_id,omitempty"`
	SlotID    string `json:"slot_id,omitempty"`
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

type TransferResponse struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	PlayerFile string `json:"player_file"`
	SlotFile   string `json:"slot_file"`
	Error      string `json:"error,omitempty"`
}

type EmptySlotResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	SlotFile string `json:"slot_file"`
	Error    string `json:"error,omitempty"`
}

type RestoreSlotResponse struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	PlayerFile string `json:"player_file"`
	SlotFile   string `json:"slot_file"`
	Error      string `json:"error,omitempty"`
}

const (
	playersDir = `C:\EVRIMA\surv_server\TheIsle\Saved\Databases\Survival\Players`
	slotsDir   = `C:\EVRIMA\surv_server\TheIsle\Saved\Slots`
)

func checkPlayerFile(steamid string) CheckResponse {
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

func getSlotFileContent(steamid, slotID string) FileContentResponse {
	slotFile := filepath.Join(slotsDir, steamid, slotID+".json")

	// Проверяем существование файла
	if _, err := os.Stat(slotFile); os.IsNotExist(err) {
		return FileContentResponse{
			Success: false,
			Error:   "Slot file not found",
		}
	} else if err != nil {
		return FileContentResponse{
			Success: false,
			Error:   err.Error(),
		}
	}

	// Читаем файл
	file, err := os.Open(slotFile)
	if err != nil {
		return FileContentResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to open slot file: %v", err),
		}
	}
	defer file.Close()

	// Читаем содержимое
	content, err := io.ReadAll(file)
	if err != nil {
		return FileContentResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to read slot file: %v", err),
		}
	}

	// Валидируем JSON (опционально, но рекомендуется)
	var jsonData json.RawMessage
	if err := json.Unmarshal(content, &jsonData); err != nil {
		return FileContentResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid JSON in slot file: %v", err),
		}
	}

	return FileContentResponse{
		Success: true,
		Content: jsonData,
	}
}

func transferPlayerSlot(steamid, oldSlotID string) TransferResponse {
	playerFile := filepath.Join(playersDir, steamid+".json")
	remoteDir := filepath.Join(`C:\EVRIMA\surv_server\TheIsle\Saved\Slots`, steamid)
	oldSlotFile := filepath.Join(remoteDir, oldSlotID+".json")

	// Проверяем существование исходного файла
	if _, err := os.Stat(playerFile); os.IsNotExist(err) {
		return TransferResponse{
			Success: false,
			Error:   "Player file not found",
		}
	} else if err != nil {
		return TransferResponse{
			Success: false,
			Error:   fmt.Sprintf("Error checking player file: %v", err),
		}
	}

	// Читаем содержимое файла игрока
	content, err := os.ReadFile(playerFile)
	if err != nil {
		return TransferResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to read player file: %v", err),
		}
	}

	// Проверяем, не является ли файл пустым или невалидным
	var decoded map[string]interface{}
	if err := json.Unmarshal(content, &decoded); err != nil || len(decoded) == 0 {
		// Создаем новый валидный JSON
		newData := map[string]interface{}{
			"slot_id":  oldSlotID,
			"datafile": nil,
		}
		content, err = json.MarshalIndent(newData, "", "  ")
		if err != nil {
			return TransferResponse{
				Success: false,
				Error:   fmt.Sprintf("Failed to create new JSON: %v", err),
			}
		}
		log.Printf("Created new JSON structure for invalid file")
	} else if _, exists := decoded["slot_id"]; !exists {
		// Добавляем slot_id если его нет
		decoded["slot_id"] = oldSlotID
		content, err = json.MarshalIndent(decoded, "", "  ")
		if err != nil {
			return TransferResponse{
				Success: false,
				Error:   fmt.Sprintf("Failed to add slot_id to JSON: %v", err),
			}
		}
		log.Printf("Added slot_id to existing JSON")
	}

	// Создаем директорию для слотов если не существует
	if err := os.MkdirAll(remoteDir, 0755); err != nil {
		return TransferResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to create directory: %v", err),
		}
	}

	// Сохраняем в слот
	if err := os.WriteFile(oldSlotFile, content, 0644); err != nil {
		return TransferResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to write slot file: %v", err),
		}
	}

	// Очищаем старый файл игрока после сохранения
	if err := os.Remove(playerFile); err != nil {
		// Логируем ошибку, но не прерываем выполнение
		log.Printf("Warning: Failed to delete player file: %v", err)
	}

	log.Printf("Old slot %s transferred from %s to %s", oldSlotID, playerFile, oldSlotFile)

	return TransferResponse{
		Success:    true,
		Message:    fmt.Sprintf("Slot %s successfully transferred", oldSlotID),
		PlayerFile: playerFile,
		SlotFile:   oldSlotFile,
	}
}

// createEmptySlot создает пустой слот при смерти динозавра
func createEmptySlot(steamid, oldSlotID string) EmptySlotResponse {
	remoteDir := filepath.Join(`C:\EVRIMA\surv_server\TheIsle\Saved\Slots`, steamid)
	oldSlotFile := filepath.Join(remoteDir, oldSlotID+".json")

	// Создаем структуру для пустого слота
	emptySlot := map[string]interface{}{
		"slot_id":  oldSlotID,
		"datafile": nil,
	}

	// Форматируем JSON с отступами
	emptySlotJSON, err := json.MarshalIndent(emptySlot, "", "  ")
	if err != nil {
		return EmptySlotResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to marshal JSON: %v", err),
		}
	}

	// Создаем директорию для слотов если не существует
	if err := os.MkdirAll(remoteDir, 0755); err != nil {
		return EmptySlotResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to create directory: %v", err),
		}
	}

	// Сохраняем пустой слот
	if err := os.WriteFile(oldSlotFile, emptySlotJSON, 0644); err != nil {
		return EmptySlotResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to write empty slot file: %v", err),
		}
	}

	log.Printf("Empty slot created for steamid %s, slot %s at %s", steamid, oldSlotID, oldSlotFile)

	return EmptySlotResponse{
		Success:  true,
		Message:  fmt.Sprintf("Empty slot %s created successfully", oldSlotID),
		SlotFile: oldSlotFile,
	}
}

// restoreSlotFromFile восстанавливает слот из файла слотов в файл игрока
func restoreSlotFromFile(steamid, slotID string) RestoreSlotResponse {
	remoteDir := filepath.Join(slotsDir, steamid)
	playersDirPath := playersDir
	slotFile := filepath.Join(remoteDir, slotID+".json")
	playerFile := filepath.Join(playersDirPath, steamid+".json")

	// Создаем директории если не существуют
	if err := os.MkdirAll(remoteDir, 0755); err != nil {
		return RestoreSlotResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to create remote directory: %v", err),
		}
	}

	if err := os.MkdirAll(playersDirPath, 0755); err != nil {
		return RestoreSlotResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to create players directory: %v", err),
		}
	}

	// Проверяем существование файла слота
	if _, err := os.Stat(slotFile); os.IsNotExist(err) {
		// Если файла нет → создаём пустой слот
		emptySlot := map[string]interface{}{
			"slot_id":  slotID,
			"datafile": nil,
			"created":  time.Now().Format("2006-01-02 15:04:05"),
		}

		jsonData, err := json.MarshalIndent(emptySlot, "", "  ")
		if err != nil {
			return RestoreSlotResponse{
				Success: false,
				Error:   fmt.Sprintf("Failed to marshal empty slot JSON: %v", err),
			}
		}

		// Сохраняем пустой слот
		if err := os.WriteFile(slotFile, jsonData, 0644); err != nil {
			return RestoreSlotResponse{
				Success: false,
				Error:   fmt.Sprintf("Failed to create empty slot file: %v", err),
			}
		}

		log.Printf("Created empty slot: %s", slotFile)

		// Используем созданный пустой слот для восстановления
		jsonData = jsonData
	} else if err != nil {
		return RestoreSlotResponse{
			Success: false,
			Error:   fmt.Sprintf("Error checking slot file: %v", err),
		}
	}

	// Читаем данные из файла слота (всегда берем данные только из файла слота)
	slotContent, err := os.ReadFile(slotFile)
	if err != nil {
		return RestoreSlotResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to read slot file: %v", err),
		}
	}

	// Валидируем JSON из слота
	var slotData map[string]interface{}
	if err := json.Unmarshal(slotContent, &slotData); err != nil {
		return RestoreSlotResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid JSON in slot file: %v", err),
		}
	}

	// Обновляем slot_id на актуальный (на случай если в файле старый)
	slotData["slot_id"] = slotID

	// Форматируем JSON для записи
	jsonData, err := json.MarshalIndent(slotData, "", "  ")
	if err != nil {
		return RestoreSlotResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to marshal JSON for player file: %v", err),
		}
	}

	// Записываем данные в файл игрока
	if err := os.WriteFile(playerFile, jsonData, 0644); err != nil {
		return RestoreSlotResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to write player file: %v", err),
		}
	}

	log.Printf("Slot restored from %s to %s", slotFile, playerFile)

	return RestoreSlotResponse{
		Success:    true,
		Message:    fmt.Sprintf("Slot %s successfully restored to player file", slotID),
		PlayerFile: playerFile,
		SlotFile:   slotFile,
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

func playerFileContentHandler(w http.ResponseWriter, r *http.Request) {
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

func slotFileContentHandler(w http.ResponseWriter, r *http.Request) {
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
		slotID := r.URL.Query().Get("slot_id")
		if steamid == "" || slotID == "" {
			http.Error(w, `{"error": "steamid and slot_id parameters are required"}`, http.StatusBadRequest)
			return
		}
		req.SteamID = steamid
		req.SlotID = slotID

	case "POST":
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
			return
		}

	default:
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if req.SteamID == "" || req.SlotID == "" {
		http.Error(w, `{"error": "steamid and slot_id are required"}`, http.StatusBadRequest)
		return
	}

	response := getSlotFileContent(req.SteamID, req.SlotID)
	json.NewEncoder(w).Encode(response)
}

func transferHandler(w http.ResponseWriter, r *http.Request) {
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
		oldSlotID := r.URL.Query().Get("old_slot_id")
		if steamid == "" || oldSlotID == "" {
			http.Error(w, `{"error": "steamid and old_slot_id parameters are required"}`, http.StatusBadRequest)
			return
		}
		req.SteamID = steamid
		req.OldSlotID = oldSlotID

	case "POST":
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
			return
		}

	default:
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if req.SteamID == "" || req.OldSlotID == "" {
		http.Error(w, `{"error": "steamid and old_slot_id are required"}`, http.StatusBadRequest)
		return
	}

	response := transferPlayerSlot(req.SteamID, req.OldSlotID)
	json.NewEncoder(w).Encode(response)
}

func emptySlotHandler(w http.ResponseWriter, r *http.Request) {
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
		oldSlotID := r.URL.Query().Get("old_slot_id")
		if steamid == "" || oldSlotID == "" {
			http.Error(w, `{"error": "steamid and old_slot_id parameters are required"}`, http.StatusBadRequest)
			return
		}
		req.SteamID = steamid
		req.OldSlotID = oldSlotID

	case "POST":
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
			return
		}

	default:
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if req.SteamID == "" || req.OldSlotID == "" {
		http.Error(w, `{"error": "steamid and old_slot_id are required"}`, http.StatusBadRequest)
		return
	}

	response := createEmptySlot(req.SteamID, req.OldSlotID)
	json.NewEncoder(w).Encode(response)
}

func restoreSlotHandler(w http.ResponseWriter, r *http.Request) {
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
		slotID := r.URL.Query().Get("slot_id")
		if steamid == "" || slotID == "" {
			http.Error(w, `{"error": "steamid and slot_id parameters are required"}`, http.StatusBadRequest)
			return
		}
		req.SteamID = steamid
		req.SlotID = slotID

	case "POST":
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
			return
		}

	default:
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if req.SteamID == "" || req.SlotID == "" {
		http.Error(w, `{"error": "steamid and slot_id are required"}`, http.StatusBadRequest)
		return
	}

	response := restoreSlotFromFile(req.SteamID, req.SlotID)
	json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/check", checkHandler)
	http.HandleFunc("/player-file", playerFileContentHandler)
	http.HandleFunc("/slot-file", slotFileContentHandler)
	http.HandleFunc("/transfer", transferHandler)
	http.HandleFunc("/empty-slot", emptySlotHandler)
	http.HandleFunc("/restore-slot", restoreSlotHandler)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status": "ok"}`))
	})

	port := ":8080"
	fmt.Printf("Server starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
