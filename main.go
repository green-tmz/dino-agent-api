package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

type WriteSlotRequest struct {
	SteamID  string          `json:"steamid"`
	FileName string          `json:"file_name"`
	Data     json.RawMessage `json:"data"`
}

type WriteSlotResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	FilePath string `json:"file_path,omitempty"`
	Error    string `json:"error,omitempty"`
}

type FilePathRequest struct {
	FilePath string `json:"file_path"`
}

type FileContentByPathResponse struct {
	Success bool   `json:"success"`
	Content string `json:"content,omitempty"`
	Error   string `json:"error,omitempty"`
	Size    int64  `json:"size,omitempty"`
}

type WriteFileRequest struct {
	FilePath string          `json:"file_path"`
	Data     json.RawMessage `json:"data"`
}

type WriteFileResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	FilePath string `json:"file_path,omitempty"`
	Size     int64  `json:"size,omitempty"`
	Error    string `json:"error,omitempty"`
}

type FileInfoResponse struct {
	Success              bool      `json:"success"`
	FilePath             string    `json:"file_path,omitempty"`
	Exists               bool      `json:"exists"`
	IsDirectory          bool      `json:"is_directory,omitempty"`
	Size                 int64     `json:"size,omitempty"`
	ModTime              time.Time `json:"mod_time,omitempty"`
	ModTimeUnix          int64     `json:"mod_time_unix,omitempty"`
	ModTimeFormatted     string    `json:"mod_time_formatted,omitempty"`
	CreatedTime          time.Time `json:"created_time,omitempty"`
	CreatedTimeUnix      int64     `json:"created_time_unix,omitempty"`
	CreatedTimeFormatted string    `json:"created_time_formatted,omitempty"`
	Error                string    `json:"error,omitempty"`
}

type DeleteFileResponse struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	FilePath   string `json:"file_path,omitempty"`
	Deleted    bool   `json:"deleted"`
	BackupPath string `json:"backup_path,omitempty"`
	Error      string `json:"error,omitempty"`
}

const (
	playersDir = `C:\EVRIMA\surv_server\TheIsle\Saved\Databases\Survival\Players`
	slotsDir   = `C:\EVRIMA\surv_server\TheIsle\Saved\Slots`
	backupDir  = `C:\EVRIMA\surv_server\backups`
)

func writeFileByPath(filePath string, data json.RawMessage) WriteFileResponse {
	log.Printf("Writing file by path: %s", filePath)

	// Проверяем, что путь не пустой
	if filePath == "" {
		result := WriteFileResponse{
			Success: false,
			Error:   "File path is required",
		}
		log.Printf("File path is empty")
		return result
	}

	// Проверяем, что данные не пустые
	if len(data) == 0 {
		result := WriteFileResponse{
			Success: false,
			Error:   "Data is required",
		}
		log.Printf("Data is empty")
		return result
	}

	// Валидируем JSON данные
	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		result := WriteFileResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid JSON data: %v", err),
		}
		log.Printf("Invalid JSON data: %v", err)
		return result
	}

	// Форматируем JSON с отступами для читаемости
	formattedData, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		result := WriteFileResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to format JSON: %v", err),
		}
		log.Printf("Failed to format JSON: %v", err)
		return result
	}

	// Создаем директорию если не существует
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		result := WriteFileResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to create directory: %v", err),
		}
		log.Printf("Failed to create directory %s: %v", dir, err)
		return result
	}

	// Записываем файл
	if err := os.WriteFile(filePath, formattedData, 0644); err != nil {
		result := WriteFileResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to write file: %v", err),
		}
		log.Printf("Failed to write file %s: %v", filePath, err)
		return result
	}

	// Получаем информацию о файле для логирования
	fileInfo, err := os.Stat(filePath)
	var fileSize int64
	if err == nil {
		fileSize = fileInfo.Size()
	}

	log.Printf("Successfully wrote file: %s, size: %d bytes", filePath, fileSize)

	result := WriteFileResponse{
		Success:  true,
		Message:  fmt.Sprintf("File %s successfully written", filepath.Base(filePath)),
		FilePath: filePath,
		Size:     fileSize,
	}
	log.Printf("File write completed successfully")
	return result
}

func getFileContentByPath(filePath string) FileContentByPathResponse {
	log.Printf("Getting file content by path: %s", filePath)

	// Проверяем, что путь не пустой
	if filePath == "" {
		result := FileContentByPathResponse{
			Success: false,
			Error:   "File path is required",
		}
		log.Printf("File path is empty")
		return result
	}

	// Проверяем существование файла
	fileInfo, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		result := FileContentByPathResponse{
			Success: false,
			Error:   "File not found",
		}
		log.Printf("File not found: %s", filePath)
		return result
	} else if err != nil {
		result := FileContentByPathResponse{
			Success: false,
			Error:   fmt.Sprintf("Error checking file: %v", err),
		}
		log.Printf("Error checking file: %v", err)
		return result
	}

	// Проверяем, что это файл, а не директория
	if fileInfo.IsDir() {
		result := FileContentByPathResponse{
			Success: false,
			Error:   "Path points to a directory, not a file",
		}
		log.Printf("Path is a directory: %s", filePath)
		return result
	}

	// Проверяем размер файла (ограничим очень большие файлы)
	if fileInfo.Size() > 10*1024*1024 { // 10MB limit
		result := FileContentByPathResponse{
			Success: false,
			Error:   "File too large (max 10MB)",
		}
		log.Printf("File too large: %d bytes", fileInfo.Size())
		return result
	}

	// Читаем файл
	file, err := os.Open(filePath)
	if err != nil {
		result := FileContentByPathResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to open file: %v", err),
		}
		log.Printf("Failed to open file: %v", err)
		return result
	}
	defer file.Close()

	// Читаем содержимое
	content, err := io.ReadAll(file)
	if err != nil {
		result := FileContentByPathResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to read file: %v", err),
		}
		log.Printf("Failed to read file: %v", err)
		return result
	}

	log.Printf("Successfully read file content, size: %d bytes", len(content))

	result := FileContentByPathResponse{
		Success: true,
		Content: string(content),
		Size:    int64(len(content)),
	}
	log.Printf("File content retrieved successfully")
	return result
}

func checkPlayerFile(steamid string) CheckResponse {
	log.Printf("Checking player file for SteamID: %s", steamid)
	playerFile := filepath.Join(playersDir, steamid+".json")

	if _, err := os.Stat(playerFile); os.IsNotExist(err) {
		result := CheckResponse{
			Exists:   false,
			FilePath: playerFile,
		}
		log.Printf("Player file not found: %s", playerFile)
		return result
	} else if err != nil {
		result := CheckResponse{
			Exists:   false,
			FilePath: playerFile,
			Error:    err.Error(),
		}
		log.Printf("Error checking player file: %v", err)
		return result
	}

	result := CheckResponse{
		Exists:   true,
		FilePath: playerFile,
	}
	log.Printf("Player file exists: %s", playerFile)
	return result
}

func getPlayerFileContent(steamid string) FileContentResponse {
	log.Printf("Getting player file content for SteamID: %s", steamid)
	playerFile := filepath.Join(playersDir, steamid+".json")

	// Проверяем существование файла
	if _, err := os.Stat(playerFile); os.IsNotExist(err) {
		result := FileContentResponse{
			Success: false,
			Error:   "File not found",
		}
		log.Printf("Player file not found: %s", playerFile)
		return result
	} else if err != nil {
		result := FileContentResponse{
			Success: false,
			Error:   err.Error(),
		}
		log.Printf("Error checking player file: %v", err)
		return result
	}

	// Читаем файл
	file, err := os.Open(playerFile)
	if err != nil {
		result := FileContentResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to open file: %v", err),
		}
		log.Printf("Failed to open player file: %v", err)
		return result
	}
	defer file.Close()

	// Читаем содержимое
	content, err := io.ReadAll(file)
	if err != nil {
		result := FileContentResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to read file: %v", err),
		}
		log.Printf("Failed to read player file: %v", err)
		return result
	}

	// Валидируем JSON (опционально, но рекомендуется)
	var jsonData json.RawMessage
	if err := json.Unmarshal(content, &jsonData); err != nil {
		result := FileContentResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid JSON in file: %v", err),
		}
		log.Printf("Invalid JSON in player file: %v", err)
		return result
	}

	result := FileContentResponse{
		Success: true,
		Content: jsonData,
	}
	log.Printf("Successfully read player file content, length: %d bytes", len(content))
	return result
}

func getSlotFileContent(steamid, slotID string) FileContentResponse {
	log.Printf("Getting slot file content for SteamID: %s, SlotID: %s", steamid, slotID)
	slotFile := filepath.Join(slotsDir, steamid, slotID+".json")

	// Проверяем существование файла
	if _, err := os.Stat(slotFile); os.IsNotExist(err) {
		result := FileContentResponse{
			Success: false,
			Error:   "Slot file not found",
		}
		log.Printf("Slot file not found: %s", slotFile)
		return result
	} else if err != nil {
		result := FileContentResponse{
			Success: false,
			Error:   err.Error(),
		}
		log.Printf("Error checking slot file: %v", err)
		return result
	}

	// Читаем файл
	file, err := os.Open(slotFile)
	if err != nil {
		result := FileContentResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to open slot file: %v", err),
		}
		log.Printf("Failed to open slot file: %v", err)
		return result
	}
	defer file.Close()

	// Читаем содержимое
	content, err := io.ReadAll(file)
	if err != nil {
		result := FileContentResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to read slot file: %v", err),
		}
		log.Printf("Failed to read slot file: %v", err)
		return result
	}

	// Валидируем JSON (опционально, но рекомендуется)
	var jsonData json.RawMessage
	if err := json.Unmarshal(content, &jsonData); err != nil {
		result := FileContentResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid JSON in slot file: %v", err),
		}
		log.Printf("Invalid JSON in slot file: %v", err)
		return result
	}

	result := FileContentResponse{
		Success: true,
		Content: jsonData,
	}
	log.Printf("Successfully read slot file content, length: %d bytes", len(content))
	return result
}

func transferPlayerSlot(steamid, oldSlotID string) TransferResponse {
	log.Printf("Transferring player slot for SteamID: %s, OldSlotID: %s", steamid, oldSlotID)
	playerFile := filepath.Join(playersDir, steamid+".json")
	remoteDir := filepath.Join(`C:\EVRIMA\surv_server\TheIsle\Saved\Slots`, steamid)
	oldSlotFile := filepath.Join(remoteDir, oldSlotID+".json")

	// Проверяем существование исходного файла
	if _, err := os.Stat(playerFile); os.IsNotExist(err) {
		result := TransferResponse{
			Success: false,
			Error:   "Player file not found",
		}
		log.Printf("Player file not found: %s", playerFile)
		return result
	} else if err != nil {
		result := TransferResponse{
			Success: false,
			Error:   fmt.Sprintf("Error checking player file: %v", err),
		}
		log.Printf("Error checking player file: %v", err)
		return result
	}

	// Читаем содержимое файла игрока
	content, err := os.ReadFile(playerFile)
	if err != nil {
		result := TransferResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to read player file: %v", err),
		}
		log.Printf("Failed to read player file: %v", err)
		return result
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
			result := TransferResponse{
				Success: false,
				Error:   fmt.Sprintf("Failed to create new JSON: %v", err),
			}
			log.Printf("Failed to create new JSON: %v", err)
			return result
		}
		log.Printf("Created new JSON structure for invalid file")
	} else if _, exists := decoded["slot_id"]; !exists {
		// Добавляем slot_id если его нет
		decoded["slot_id"] = oldSlotID
		content, err = json.MarshalIndent(decoded, "", "  ")
		if err != nil {
			result := TransferResponse{
				Success: false,
				Error:   fmt.Sprintf("Failed to add slot_id to JSON: %v", err),
			}
			log.Printf("Failed to add slot_id to JSON: %v", err)
			return result
		}
		log.Printf("Added slot_id to existing JSON")
	}

	// Создаем директорию для слотов если не существует
	if err := os.MkdirAll(remoteDir, 0755); err != nil {
		result := TransferResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to create directory: %v", err),
		}
		log.Printf("Failed to create directory: %v", err)
		return result
	}

	// Сохраняем в слот
	if err := os.WriteFile(oldSlotFile, content, 0644); err != nil {
		result := TransferResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to write slot file: %v", err),
		}
		log.Printf("Failed to write slot file: %v", err)
		return result
	}

	// Очищаем старый файл игрока после сохранения
	if err := os.Remove(playerFile); err != nil {
		// Логируем ошибку, но не прерываем выполнение
		log.Printf("Warning: Failed to delete player file: %v", err)
	}

	log.Printf("Old slot %s transferred from %s to %s", oldSlotID, playerFile, oldSlotFile)

	result := TransferResponse{
		Success:    true,
		Message:    fmt.Sprintf("Slot %s successfully transferred", oldSlotID),
		PlayerFile: playerFile,
		SlotFile:   oldSlotFile,
	}
	log.Printf("Transfer completed successfully")
	return result
}

func createEmptySlot(steamid, oldSlotID string) EmptySlotResponse {
	log.Printf("Creating empty slot for SteamID: %s, SlotID: %s", steamid, oldSlotID)
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
		result := EmptySlotResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to marshal JSON: %v", err),
		}
		log.Printf("Failed to marshal JSON for empty slot: %v", err)
		return result
	}

	// Создаем директорию для слотов если не существует
	if err := os.MkdirAll(remoteDir, 0755); err != nil {
		result := EmptySlotResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to create directory: %v", err),
		}
		log.Printf("Failed to create directory: %v", err)
		return result
	}

	// Сохраняем пустой слот
	if err := os.WriteFile(oldSlotFile, emptySlotJSON, 0644); err != nil {
		result := EmptySlotResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to write empty slot file: %v", err),
		}
		log.Printf("Failed to write empty slot file: %v", err)
		return result
	}

	log.Printf("Empty slot created for steamid %s, slot %s at %s", steamid, oldSlotID, oldSlotFile)

	result := EmptySlotResponse{
		Success:  true,
		Message:  fmt.Sprintf("Empty slot %s created successfully", oldSlotID),
		SlotFile: oldSlotFile,
	}
	log.Printf("Empty slot creation completed successfully")
	return result
}

func restoreSlotFromFile(steamid, slotID string) RestoreSlotResponse {
	log.Printf("Restoring slot from file for SteamID: %s, SlotID: %s", steamid, slotID)
	remoteDir := filepath.Join(slotsDir, steamid)
	playersDirPath := playersDir
	slotFile := filepath.Join(remoteDir, slotID+".json")
	playerFile := filepath.Join(playersDirPath, steamid+".json")

	// Создаем директории если не существуют
	if err := os.MkdirAll(remoteDir, 0755); err != nil {
		result := RestoreSlotResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to create remote directory: %v", err),
		}
		log.Printf("Failed to create remote directory: %v", err)
		return result
	}

	if err := os.MkdirAll(playersDirPath, 0755); err != nil {
		result := RestoreSlotResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to create players directory: %v", err),
		}
		log.Printf("Failed to create players directory: %v", err)
		return result
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
			result := RestoreSlotResponse{
				Success: false,
				Error:   fmt.Sprintf("Failed to marshal empty slot JSON: %v", err),
			}
			log.Printf("Failed to marshal empty slot JSON: %v", err)
			return result
		}

		// Сохраняем пустой слот
		if err := os.WriteFile(slotFile, jsonData, 0644); err != nil {
			result := RestoreSlotResponse{
				Success: false,
				Error:   fmt.Sprintf("Failed to create empty slot file: %v", err),
			}
			log.Printf("Failed to create empty slot file: %v", err)
			return result
		}

		log.Printf("Created empty slot: %s", slotFile)

		// Используем созданный пустой слот для восстановления
		jsonData = jsonData
	} else if err != nil {
		result := RestoreSlotResponse{
			Success: false,
			Error:   fmt.Sprintf("Error checking slot file: %v", err),
		}
		log.Printf("Error checking slot file: %v", err)
		return result
	}

	// Читаем данные из файла слота (всегда берем данные только из файла слота)
	slotContent, err := os.ReadFile(slotFile)
	if err != nil {
		result := RestoreSlotResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to read slot file: %v", err),
		}
		log.Printf("Failed to read slot file: %v", err)
		return result
	}

	// Валидируем JSON из слота
	var slotData map[string]interface{}
	if err := json.Unmarshal(slotContent, &slotData); err != nil {
		result := RestoreSlotResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid JSON in slot file: %v", err),
		}
		log.Printf("Invalid JSON in slot file: %v", err)
		return result
	}

	// Обновляем slot_id на актуальный (на случай если в файле старый)
	slotData["slot_id"] = slotID

	// Форматируем JSON для записи
	jsonData, err := json.MarshalIndent(slotData, "", "  ")
	if err != nil {
		result := RestoreSlotResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to marshal JSON for player file: %v", err),
		}
		log.Printf("Failed to marshal JSON for player file: %v", err)
		return result
	}

	// Записываем данные в файл игрока
	if err := os.WriteFile(playerFile, jsonData, 0644); err != nil {
		result := RestoreSlotResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to write player file: %v", err),
		}
		log.Printf("Failed to write player file: %v", err)
		return result
	}

	log.Printf("Slot restored from %s to %s", slotFile, playerFile)

	result := RestoreSlotResponse{
		Success:    true,
		Message:    fmt.Sprintf("Slot %s successfully restored to player file", slotID),
		PlayerFile: playerFile,
		SlotFile:   slotFile,
	}
	log.Printf("Slot restoration completed successfully")
	return result
}

func writeSlotFile(steamid, fileName string, data json.RawMessage) WriteSlotResponse {
	log.Printf("Writing slot file for SteamID: %s, FileName: %s", steamid, fileName)

	// Проверяем, что fileName имеет расширение .json
	if filepath.Ext(fileName) != ".json" {
		fileName = fileName + ".json"
	}

	remoteDir := filepath.Join(slotsDir, steamid)
	filePath := filepath.Join(remoteDir, fileName)

	// Создаем директорию если не существует
	if err := os.MkdirAll(remoteDir, 0755); err != nil {
		result := WriteSlotResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to create directory: %v", err),
		}
		log.Printf("Failed to create directory: %v", err)
		return result
	}

	// Если данные не предоставлены, создаем структуру по умолчанию
	if len(data) == 0 || string(data) == "null" {
		defaultData := map[string]interface{}{
			"slot_id":  strings.TrimSuffix(fileName, ".json"),
			"datafile": nil,
			"created":  time.Now().Format("2006-01-02 15:04:05"),
		}

		jsonData, err := json.MarshalIndent(defaultData, "", "  ")
		if err != nil {
			result := WriteSlotResponse{
				Success: false,
				Error:   fmt.Sprintf("Failed to marshal default JSON: %v", err),
			}
			log.Printf("Failed to marshal default JSON: %v", err)
			return result
		}
		data = jsonData
		log.Printf("Using default data structure for slot file")
	} else {
		// Валидируем предоставленные JSON данные
		var jsonData interface{}
		if err := json.Unmarshal(data, &jsonData); err != nil {
			result := WriteSlotResponse{
				Success: false,
				Error:   fmt.Sprintf("Invalid JSON data: %v", err),
			}
			log.Printf("Invalid JSON data: %v", err)
			return result
		}

		// Переформатируем JSON с отступами
		formattedData, err := json.MarshalIndent(jsonData, "", "  ")
		if err != nil {
			result := WriteSlotResponse{
				Success: false,
				Error:   fmt.Sprintf("Failed to format JSON: %v", err),
			}
			log.Printf("Failed to format JSON: %v", err)
			return result
		}
		data = formattedData
		log.Printf("Using provided data for slot file, length: %d bytes", len(data))
	}

	// Записываем файл
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		result := WriteSlotResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to write file: %v", err),
		}
		log.Printf("Failed to write file: %v", err)
		return result
	}

	log.Printf("Data written to slot file for steamid %s, file %s at %s", steamid, fileName, filePath)

	result := WriteSlotResponse{
		Success:  true,
		Message:  fmt.Sprintf("Data successfully written to %s", fileName),
		FilePath: filePath,
	}
	log.Printf("Write slot file completed successfully")
	return result
}

func writeFileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

	if r.Method == "OPTIONS" {
		return
	}

	var req WriteFileRequest

	switch r.Method {
	case "GET":
		filePath := r.URL.Query().Get("file_path")
		dataStr := r.URL.Query().Get("data")

		if filePath == "" {
			log.Printf("Write file handler: missing file_path parameter in GET request")
			http.Error(w, `{"error": "file_path parameter is required"}`, http.StatusBadRequest)
			return
		}

		req.FilePath = filePath
		if dataStr != "" {
			req.Data = json.RawMessage(dataStr)
		}

	case "POST":
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("Write file handler: invalid JSON in POST request: %v", err)
			http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
			return
		}

	default:
		log.Printf("Write file handler: method not allowed: %s", r.Method)
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if req.FilePath == "" {
		log.Printf("Write file handler: file_path is required")
		http.Error(w, `{"error": "file_path is required"}`, http.StatusBadRequest)
		return
	}

	if len(req.Data) == 0 {
		log.Printf("Write file handler: data is required")
		http.Error(w, `{"error": "data is required"}`, http.StatusBadRequest)
		return
	}

	log.Printf("Write file handler processing request for path: %s", req.FilePath)
	response := writeFileByPath(req.FilePath, req.Data)
	log.Printf("Write file handler response: Success=%t, Error=%s, Size=%d", response.Success, response.Error, response.Size)
	json.NewEncoder(w).Encode(response)
}

func fileContentByPathHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

	if r.Method == "OPTIONS" {
		return
	}

	var req FilePathRequest

	switch r.Method {
	case "GET":
		filePath := r.URL.Query().Get("file_path")
		if filePath == "" {
			log.Printf("File content by path handler: missing file_path parameter in GET request")
			http.Error(w, `{"error": "file_path parameter is required"}`, http.StatusBadRequest)
			return
		}
		req.FilePath = filePath

	case "POST":
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("File content by path handler: invalid JSON in POST request: %v", err)
			http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
			return
		}

	default:
		log.Printf("File content by path handler: method not allowed: %s", r.Method)
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if req.FilePath == "" {
		log.Printf("File content by path handler: file_path is required")
		http.Error(w, `{"error": "file_path is required"}`, http.StatusBadRequest)
		return
	}

	log.Printf("File content by path handler processing request for path: %s", req.FilePath)
	response := getFileContentByPath(req.FilePath)
	log.Printf("File content by path handler response: Success=%t, Error=%s, Size=%d", response.Success, response.Error, response.Size)
	json.NewEncoder(w).Encode(response)
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
			log.Printf("Check handler: missing steamid parameter in GET request")
			http.Error(w, `{"error": "steamid parameter is required"}`, http.StatusBadRequest)
			return
		}
		req.SteamID = steamid

	case "POST":
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("Check handler: invalid JSON in POST request: %v", err)
			http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
			return
		}

	default:
		log.Printf("Check handler: method not allowed: %s", r.Method)
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if req.SteamID == "" {
		log.Printf("Check handler: steamid is required")
		http.Error(w, `{"error": "steamid is required"}`, http.StatusBadRequest)
		return
	}

	log.Printf("Check handler processing request for SteamID: %s", req.SteamID)
	response := checkPlayerFile(req.SteamID)
	log.Printf("Check handler response: Exists=%t, Error=%s", response.Exists, response.Error)
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
			log.Printf("Player file content handler: missing steamid parameter in GET request")
			http.Error(w, `{"error": "steamid parameter is required"}`, http.StatusBadRequest)
			return
		}
		req.SteamID = steamid

	case "POST":
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("Player file content handler: invalid JSON in POST request: %v", err)
			http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
			return
		}

	default:
		log.Printf("Player file content handler: method not allowed: %s", r.Method)
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if req.SteamID == "" {
		log.Printf("Player file content handler: steamid is required")
		http.Error(w, `{"error": "steamid is required"}`, http.StatusBadRequest)
		return
	}

	log.Printf("Player file content handler processing request for SteamID: %s", req.SteamID)
	response := getPlayerFileContent(req.SteamID)
	log.Printf("Player file content handler response: Success=%t, Error=%s", response.Success, response.Error)
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
			log.Printf("Slot file content handler: missing parameters in GET request - steamid: %s, slot_id: %s", steamid, slotID)
			http.Error(w, `{"error": "steamid and slot_id parameters are required"}`, http.StatusBadRequest)
			return
		}
		req.SteamID = steamid
		req.SlotID = slotID

	case "POST":
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("Slot file content handler: invalid JSON in POST request: %v", err)
			http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
			return
		}

	default:
		log.Printf("Slot file content handler: method not allowed: %s", r.Method)
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if req.SteamID == "" || req.SlotID == "" {
		log.Printf("Slot file content handler: steamid and slot_id are required")
		http.Error(w, `{"error": "steamid and slot_id are required"}`, http.StatusBadRequest)
		return
	}

	log.Printf("Slot file content handler processing request for SteamID: %s, SlotID: %s", req.SteamID, req.SlotID)
	response := getSlotFileContent(req.SteamID, req.SlotID)
	log.Printf("Slot file content handler response: Success=%t, Error=%s", response.Success, response.Error)
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
			log.Printf("Transfer handler: missing parameters in GET request - steamid: %s, old_slot_id: %s", steamid, oldSlotID)
			http.Error(w, `{"error": "steamid and old_slot_id parameters are required"}`, http.StatusBadRequest)
			return
		}
		req.SteamID = steamid
		req.OldSlotID = oldSlotID

	case "POST":
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("Transfer handler: invalid JSON in POST request: %v", err)
			http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
			return
		}

	default:
		log.Printf("Transfer handler: method not allowed: %s", r.Method)
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if req.SteamID == "" || req.OldSlotID == "" {
		log.Printf("Transfer handler: steamid and old_slot_id are required")
		http.Error(w, `{"error": "steamid and old_slot_id are required"}`, http.StatusBadRequest)
		return
	}

	log.Printf("Transfer handler processing request for SteamID: %s, OldSlotID: %s", req.SteamID, req.OldSlotID)
	response := transferPlayerSlot(req.SteamID, req.OldSlotID)
	log.Printf("Transfer handler response: Success=%t, Error=%s", response.Success, response.Error)
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
			log.Printf("Empty slot handler: missing parameters in GET request - steamid: %s, old_slot_id: %s", steamid, oldSlotID)
			http.Error(w, `{"error": "steamid and old_slot_id parameters are required"}`, http.StatusBadRequest)
			return
		}
		req.SteamID = steamid
		req.OldSlotID = oldSlotID

	case "POST":
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("Empty slot handler: invalid JSON in POST request: %v", err)
			http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
			return
		}

	default:
		log.Printf("Empty slot handler: method not allowed: %s", r.Method)
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if req.SteamID == "" || req.OldSlotID == "" {
		log.Printf("Empty slot handler: steamid and old_slot_id are required")
		http.Error(w, `{"error": "steamid and old_slot_id are required"}`, http.StatusBadRequest)
		return
	}

	log.Printf("Empty slot handler processing request for SteamID: %s, OldSlotID: %s", req.SteamID, req.OldSlotID)
	response := createEmptySlot(req.SteamID, req.OldSlotID)
	log.Printf("Empty slot handler response: Success=%t, Error=%s", response.Success, response.Error)
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
			log.Printf("Restore slot handler: missing parameters in GET request - steamid: %s, slot_id: %s", steamid, slotID)
			http.Error(w, `{"error": "steamid and slot_id parameters are required"}`, http.StatusBadRequest)
			return
		}
		req.SteamID = steamid
		req.SlotID = slotID

	case "POST":
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("Restore slot handler: invalid JSON in POST request: %v", err)
			http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
			return
		}

	default:
		log.Printf("Restore slot handler: method not allowed: %s", r.Method)
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if req.SteamID == "" || req.SlotID == "" {
		log.Printf("Restore slot handler: steamid and slot_id are required")
		http.Error(w, `{"error": "steamid and slot_id are required"}`, http.StatusBadRequest)
		return
	}

	log.Printf("Restore slot handler processing request for SteamID: %s, SlotID: %s", req.SteamID, req.SlotID)
	response := restoreSlotFromFile(req.SteamID, req.SlotID)
	log.Printf("Restore slot handler response: Success=%t, Error=%s", response.Success, response.Error)
	json.NewEncoder(w).Encode(response)
}

func writeSlotHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

	if r.Method == "OPTIONS" {
		return
	}

	var req WriteSlotRequest

	switch r.Method {
	case "GET":
		steamid := r.URL.Query().Get("steamid")
		fileName := r.URL.Query().Get("file_name")
		dataStr := r.URL.Query().Get("data")

		if steamid == "" || fileName == "" {
			log.Printf("Write slot handler: missing parameters in GET request - steamid: %s, file_name: %s", steamid, fileName)
			http.Error(w, `{"error": "steamid and file_name parameters are required"}`, http.StatusBadRequest)
			return
		}

		req.SteamID = steamid
		req.FileName = fileName
		if dataStr != "" {
			req.Data = json.RawMessage(dataStr)
		}

	case "POST":
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("Write slot handler: invalid JSON in POST request: %v", err)
			http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
			return
		}

	default:
		log.Printf("Write slot handler: method not allowed: %s", r.Method)
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if req.SteamID == "" || req.FileName == "" {
		log.Printf("Write slot handler: steamid and file_name are required")
		http.Error(w, `{"error": "steamid and file_name are required"}`, http.StatusBadRequest)
		return
	}

	log.Printf("Write slot handler processing request for SteamID: %s, FileName: %s", req.SteamID, req.FileName)
	response := writeSlotFile(req.SteamID, req.FileName, req.Data)
	log.Printf("Write slot handler response: Success=%t, Error=%s", response.Success, response.Error)
	json.NewEncoder(w).Encode(response)
}

func getFileInfo(filePath string) FileInfoResponse {
	log.Printf("Getting file info for: %s", filePath)

	// Проверяем, что путь не пустой
	if filePath == "" {
		result := FileInfoResponse{
			Success: false,
			Error:   "File path is required",
		}
		log.Printf("File path is empty")
		return result
	}

	// Получаем информацию о файле
	fileInfo, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		result := FileInfoResponse{
			Success:  true,
			FilePath: filePath,
			Exists:   false,
		}
		log.Printf("File not found: %s", filePath)
		return result
	} else if err != nil {
		result := FileInfoResponse{
			Success:  false,
			FilePath: filePath,
			Error:    fmt.Sprintf("Error getting file info: %v", err),
		}
		log.Printf("Error getting file info for %s: %v", filePath, err)
		return result
	}

	// Получаем время создания файла (для Windows)
	createdTime := getFileCreationTime(filePath)

	// Форматируем время для удобства чтения
	modTimeFormatted := fileInfo.ModTime().Format("2006-01-02 15:04:05")
	createdTimeFormatted := ""
	if !createdTime.IsZero() {
		createdTimeFormatted = createdTime.Format("2006-01-02 15:04:05")
	}

	result := FileInfoResponse{
		Success:              true,
		FilePath:             filePath,
		Exists:               true,
		IsDirectory:          fileInfo.IsDir(),
		Size:                 fileInfo.Size(),
		ModTime:              fileInfo.ModTime(),
		ModTimeUnix:          fileInfo.ModTime().Unix(),
		ModTimeFormatted:     modTimeFormatted,
		CreatedTime:          createdTime,
		CreatedTimeUnix:      createdTime.Unix(),
		CreatedTimeFormatted: createdTimeFormatted,
	}

	log.Printf("File info retrieved successfully: exists=%t, size=%d, mod_time=%s",
		result.Exists, result.Size, result.ModTimeFormatted)

	return result
}

func getFileCreationTime(filePath string) time.Time {
	file, err := os.Open(filePath)
	if err != nil {
		return time.Time{} // возвращаем нулевое время в случае ошибки
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return time.Time{} // возвращаем нулевое время в случае ошибки
	}

	return fileInfo.ModTime()
}

func fileInfoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

	if r.Method == "OPTIONS" {
		return
	}

	var req FilePathRequest

	switch r.Method {
	case "GET":
		filePath := r.URL.Query().Get("file_path")
		if filePath == "" {
			log.Printf("File info handler: missing file_path parameter in GET request")
			http.Error(w, `{"error": "file_path parameter is required"}`, http.StatusBadRequest)
			return
		}
		req.FilePath = filePath

	case "POST":
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("File info handler: invalid JSON in POST request: %v", err)
			http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
			return
		}

	default:
		log.Printf("File info handler: method not allowed: %s", r.Method)
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if req.FilePath == "" {
		log.Printf("File info handler: file_path is required")
		http.Error(w, `{"error": "file_path is required"}`, http.StatusBadRequest)
		return
	}

	log.Printf("File info handler processing request for path: %s", req.FilePath)
	response := getFileInfo(req.FilePath)
	log.Printf("File info handler response: Success=%t, Exists=%t, Error=%s",
		response.Success, response.Exists, response.Error)
	json.NewEncoder(w).Encode(response)
}

func deleteFileByPath(filePath string, backup bool) DeleteFileResponse {
	log.Printf("Deleting file by path: %s, backup: %t", filePath, backup)

	// Проверяем, что путь не пустой
	if filePath == "" {
		result := DeleteFileResponse{
			Success: false,
			Error:   "File path is required",
		}
		log.Printf("File path is empty")
		return result
	}

	// Проверяем существование файла
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		result := DeleteFileResponse{
			Success:  true,
			FilePath: filePath,
			Deleted:  false,
			Message:  "File does not exist, nothing to delete",
		}
		log.Printf("File not found, nothing to delete: %s", filePath)
		return result
	} else if err != nil {
		result := DeleteFileResponse{
			Success:  false,
			FilePath: filePath,
			Error:    fmt.Sprintf("Error checking file: %v", err),
		}
		log.Printf("Error checking file %s: %v", filePath, err)
		return result
	}

	var backupPath string

	// Создаем бэкап если требуется
	if backup {
		backupPath = createBackup(filePath)
		if backupPath != "" {
			log.Printf("Backup created: %s", backupPath)
		} else {
			log.Printf("Failed to create backup for: %s", filePath)
			// Продолжаем удаление даже если бэкап не удался
		}
	}

	// Удаляем файл
	if err := os.Remove(filePath); err != nil {
		result := DeleteFileResponse{
			Success:    false,
			FilePath:   filePath,
			BackupPath: backupPath,
			Error:      fmt.Sprintf("Failed to delete file: %v", err),
		}
		log.Printf("Failed to delete file %s: %v", filePath, err)
		return result
	}

	// Проверяем, что файл действительно удален
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		result := DeleteFileResponse{
			Success:    true,
			Message:    fmt.Sprintf("File %s successfully deleted", filepath.Base(filePath)),
			FilePath:   filePath,
			Deleted:    true,
			BackupPath: backupPath,
		}
		log.Printf("File successfully deleted: %s", filePath)
		return result
	} else {
		result := DeleteFileResponse{
			Success:    false,
			FilePath:   filePath,
			BackupPath: backupPath,
			Error:      "File still exists after deletion attempt",
		}
		log.Printf("File still exists after deletion: %s", filePath)
		return result
	}
}

func createBackup(filePath string) string {
	// Создаем имя для бэкап файла
	fileName := filepath.Base(filePath)
	backupFileName := fmt.Sprintf("%s_%s.backup",
		strings.TrimSuffix(fileName, filepath.Ext(fileName)),
		time.Now().Format("20060102_150405"))

	backupPath := filepath.Join(backupDir, backupFileName)

	// Создаем директорию для бэкапов если не существует
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		log.Printf("Failed to create backup directory %s: %v", backupDir, err)
		return ""
	}

	// Копируем файл
	sourceFile, err := os.Open(filePath)
	if err != nil {
		log.Printf("Failed to open source file for backup %s: %v", filePath, err)
		return ""
	}
	defer sourceFile.Close()

	backupFile, err := os.Create(backupPath)
	if err != nil {
		log.Printf("Failed to create backup file %s: %v", backupPath, err)
		return ""
	}
	defer backupFile.Close()

	_, err = io.Copy(backupFile, sourceFile)
	if err != nil {
		log.Printf("Failed to copy file to backup %s: %v", backupPath, err)
		return ""
	}

	log.Printf("Backup created successfully: %s", backupPath)
	return backupPath
}

func deletePlayerFile(steamid string) DeleteFileResponse {
	playerFile := filepath.Join(playersDir, steamid+".json")
	return deleteFileByPath(playerFile, true) // Всегда делаем бэкап для файлов игроков
}

func deleteSlotFile(steamid, slotID string) DeleteFileResponse {
	slotFile := filepath.Join(slotsDir, steamid, slotID+".json")
	return deleteFileByPath(slotFile, true) // Всегда делаем бэкап для файлов слотов
}

func deleteEmptyDirectory(dirPath string) DeleteFileResponse {
	log.Printf("Deleting directory: %s", dirPath)

	// Проверяем, что путь не пустой
	if dirPath == "" {
		result := DeleteFileResponse{
			Success: false,
			Error:   "Directory path is required",
		}
		log.Printf("Directory path is empty")
		return result
	}

	// Проверяем существование директории
	fileInfo, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		result := DeleteFileResponse{
			Success:  true,
			FilePath: dirPath,
			Deleted:  false,
			Message:  "Directory does not exist, nothing to delete",
		}
		log.Printf("Directory not found, nothing to delete: %s", dirPath)
		return result
	} else if err != nil {
		result := DeleteFileResponse{
			Success:  false,
			FilePath: dirPath,
			Error:    fmt.Sprintf("Error checking directory: %v", err),
		}
		log.Printf("Error checking directory %s: %v", dirPath, err)
		return result
	}

	// Проверяем, что это действительно директория
	if !fileInfo.IsDir() {
		result := DeleteFileResponse{
			Success:  false,
			FilePath: dirPath,
			Error:    "Path is not a directory",
		}
		log.Printf("Path is not a directory: %s", dirPath)
		return result
	}

	// Проверяем, что директория пуста
	dir, err := os.Open(dirPath)
	if err != nil {
		result := DeleteFileResponse{
			Success:  false,
			FilePath: dirPath,
			Error:    fmt.Sprintf("Failed to open directory: %v", err),
		}
		log.Printf("Failed to open directory %s: %v", dirPath, err)
		return result
	}
	defer dir.Close()

	_, err = dir.Readdirnames(1) // Пытаемся прочитать хотя бы один файл
	if err == nil {
		result := DeleteFileResponse{
			Success:  false,
			FilePath: dirPath,
			Error:    "Directory is not empty",
		}
		log.Printf("Directory is not empty: %s", dirPath)
		return result
	}

	// Удаляем пустую директорию
	if err := os.Remove(dirPath); err != nil {
		result := DeleteFileResponse{
			Success:  false,
			FilePath: dirPath,
			Error:    fmt.Sprintf("Failed to delete directory: %v", err),
		}
		log.Printf("Failed to delete directory %s: %v", dirPath, err)
		return result
	}

	result := DeleteFileResponse{
		Success:  true,
		Message:  fmt.Sprintf("Directory %s successfully deleted", filepath.Base(dirPath)),
		FilePath: dirPath,
		Deleted:  true,
	}
	log.Printf("Directory successfully deleted: %s", dirPath)
	return result
}

func deleteFileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

	if r.Method == "OPTIONS" {
		return
	}

	var req FilePathRequest
	backup := true // По умолчанию создаем бэкап

	switch r.Method {
	case "GET":
		filePath := r.URL.Query().Get("file_path")
		backupParam := r.URL.Query().Get("backup")

		if filePath == "" {
			log.Printf("Delete file handler: missing file_path parameter in GET request")
			http.Error(w, `{"error": "file_path parameter is required"}`, http.StatusBadRequest)
			return
		}

		req.FilePath = filePath
		if backupParam == "false" {
			backup = false
		}

	case "POST":
		var requestBody struct {
			FilePath string `json:"file_path"`
			Backup   *bool  `json:"backup,omitempty"`
		}

		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			log.Printf("Delete file handler: invalid JSON in POST request: %v", err)
			http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
			return
		}

		req.FilePath = requestBody.FilePath
		if requestBody.Backup != nil {
			backup = *requestBody.Backup
		}

	default:
		log.Printf("Delete file handler: method not allowed: %s", r.Method)
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if req.FilePath == "" {
		log.Printf("Delete file handler: file_path is required")
		http.Error(w, `{"error": "file_path is required"}`, http.StatusBadRequest)
		return
	}

	log.Printf("Delete file handler processing request for path: %s, backup: %t", req.FilePath, backup)
	response := deleteFileByPath(req.FilePath, backup)
	log.Printf("Delete file handler response: Success=%t, Deleted=%t, Error=%s",
		response.Success, response.Deleted, response.Error)
	json.NewEncoder(w).Encode(response)
}

func deletePlayerFileHandler(w http.ResponseWriter, r *http.Request) {
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
			log.Printf("Delete player file handler: missing steamid parameter in GET request")
			http.Error(w, `{"error": "steamid parameter is required"}`, http.StatusBadRequest)
			return
		}
		req.SteamID = steamid

	case "POST":
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("Delete player file handler: invalid JSON in POST request: %v", err)
			http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
			return
		}

	default:
		log.Printf("Delete player file handler: method not allowed: %s", r.Method)
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if req.SteamID == "" {
		log.Printf("Delete player file handler: steamid is required")
		http.Error(w, `{"error": "steamid is required"}`, http.StatusBadRequest)
		return
	}

	log.Printf("Delete player file handler processing request for SteamID: %s", req.SteamID)
	response := deletePlayerFile(req.SteamID)
	log.Printf("Delete player file handler response: Success=%t, Deleted=%t, Error=%s",
		response.Success, response.Deleted, response.Error)
	json.NewEncoder(w).Encode(response)
}

func deleteSlotFileHandler(w http.ResponseWriter, r *http.Request) {
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
			log.Printf("Delete slot file handler: missing parameters in GET request - steamid: %s, slot_id: %s", steamid, slotID)
			http.Error(w, `{"error": "steamid and slot_id parameters are required"}`, http.StatusBadRequest)
			return
		}
		req.SteamID = steamid
		req.SlotID = slotID

	case "POST":
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("Delete slot file handler: invalid JSON in POST request: %v", err)
			http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
			return
		}

	default:
		log.Printf("Delete slot file handler: method not allowed: %s", r.Method)
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if req.SteamID == "" || req.SlotID == "" {
		log.Printf("Delete slot file handler: steamid and slot_id are required")
		http.Error(w, `{"error": "steamid and slot_id are required"}`, http.StatusBadRequest)
		return
	}

	log.Printf("Delete slot file handler processing request for SteamID: %s, SlotID: %s", req.SteamID, req.SlotID)
	response := deleteSlotFile(req.SteamID, req.SlotID)
	log.Printf("Delete slot file handler response: Success=%t, Deleted=%t, Error=%s",
		response.Success, response.Deleted, response.Error)
	json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/check", checkHandler)
	http.HandleFunc("/player-file", playerFileContentHandler)
	http.HandleFunc("/slot-file", slotFileContentHandler)
	http.HandleFunc("/transfer", transferHandler)
	http.HandleFunc("/empty-slot", emptySlotHandler)
	http.HandleFunc("/restore-slot", restoreSlotHandler)
	http.HandleFunc("/write-slot", writeSlotHandler)
	http.HandleFunc("/file-content", fileContentByPathHandler)
	http.HandleFunc("/write-file", writeFileHandler)
	http.HandleFunc("/file-info", fileInfoHandler)
	http.HandleFunc("/delete-file", deleteFileHandler)
	http.HandleFunc("/delete-player-file", deletePlayerFileHandler)
	http.HandleFunc("/delete-slot-file", deleteSlotFileHandler)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Health check requested from %s", r.RemoteAddr)
		w.Write([]byte(`{"status": "ok"}`))
	})

	port := ":8080"
	fmt.Printf("Server starting on port %s\n", port)
	log.Printf("Server started successfully on port %s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
