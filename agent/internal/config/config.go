package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
)

var (
	// ErrEnvFileNotFound is returned when the .env file is not found
	ErrEnvFileNotFound = errors.New(".env file not found")

	// loadOnce ensures .env is loaded only once
	loadOnce sync.Once

	// loaded indicates if the .env file has been loaded
	loaded bool
)

// LoadEnv loads environment variables from the .env file
func LoadEnv() error {
	var err error
	loadOnce.Do(func() {
		err = loadEnvFile(".env")
		if err != nil {
			return
		}
		loaded = true
	})
	return err
}

// loadEnvFile reads the .env file and sets environment variables
func loadEnvFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrEnvFileNotFound
		}
		return fmt.Errorf("error opening .env file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Skip comments and empty lines
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue // Skip malformed lines
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if len(value) > 1 && (value[0] == '"' && value[len(value)-1] == '"' ||
			value[0] == '\'' && value[len(value)-1] == '\'') {
			value = value[1 : len(value)-1]
		}

		// Set environment variable if it's not already set
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading .env file: %w", err)
	}

	return nil
}

// Get retrieves an environment variable with a fallback value
func Get(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return fallback
}

// MustGet retrieves an environment variable or panics if it's not set
func MustGet(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		panic(fmt.Sprintf("Required environment variable %s is not set", key))
	}
	return value
}

// GetInt retrieves an integer environment variable with a fallback value
func GetInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		var result int
		if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
			return result
		}
	}
	return fallback
}

// GetBool retrieves a boolean environment variable with a fallback value
func GetBool(key string, fallback bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		value = strings.ToLower(value)
		if value == "true" || value == "1" || value == "yes" || value == "y" {
			return true
		}
		if value == "false" || value == "0" || value == "no" || value == "n" {
			return false
		}
	}
	return fallback
}
