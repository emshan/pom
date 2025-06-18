package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

func getDataDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	
	dataDir := filepath.Join(homeDir, ".local", "share", "pomodoro")
	err = os.MkdirAll(dataDir, 0755)
	if err != nil {
		return "", err
	}
	
	return dataDir, nil
}

func getSessionFilename() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	
	// Create a hash of the current working directory
	hash := md5.Sum([]byte(cwd))
	sessionID := hex.EncodeToString(hash[:])
	
	dataDir, err := getDataDir()
	if err != nil {
		return "", err
	}
	
	return filepath.Join(dataDir, sessionID+".json"), nil
}

func saveTodosToFile(todos []TodoItem) error {
	filename, err := getSessionFilename()
	if err != nil {
		return err
	}
	
	data, err := json.MarshalIndent(todos, "", "  ")
	if err != nil {
		return err
	}
	
	return ioutil.WriteFile(filename, data, 0644)
}

func loadTodosFromFile() ([]TodoItem, error) {
	filename, err := getSessionFilename()
	if err != nil {
		return nil, err
	}
	
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		// File doesn't exist, return empty slice
		return []TodoItem{}, nil
	}
	
	var todos []TodoItem
	err = json.Unmarshal(data, &todos)
	if err != nil {
		return nil, err
	}
	
	return todos, nil
}