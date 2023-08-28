package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/manifoldco/promptui"
	"golang.org/x/crypto/argon2"
)

// Credentials represents the user's login credentials
type Credentials struct {
	Username string `json:"username"`
	Password []byte `json:"password"`
}

// Database represents the user credentials database
type Database struct {
	Users []Credentials `json:"users"`
}

// DeleteFile deletes the file at the specified path
func DeleteFile(path string) error {
	err := os.Remove(path)
	if err != nil {
		return fmt.Errorf("failed to delete file: %v", err)
	}
	return nil
}

// SaveCredentials saves the user's credentials to the database file
func SaveCredentials(creds Credentials, filePath string) error {
	// Read the existing database
	db, err := readDatabase(filePath)
	if err != nil {
		return err
	}

	// Check if the user already exists
	for _, user := range db.Users {
		if user.Username == creds.Username {
			return fmt.Errorf("user already exists")
		}
	}

	// Append the new user to the database
	db.Users = append(db.Users, creds)

	// Write the updated database to the file
	err = writeDatabase(db, filePath)
	if err != nil {
		return fmt.Errorf("failed to save credentials: %v", err)
	}

	return nil
}

// ReadCredentials reads the user's credentials from the database file
func ReadCredentials(username, filePath string) (Credentials, error) {
	// Read the existing database
	db, err := readDatabase(filePath)
	if err != nil {
		return Credentials{}, err
	}

	// Find the user in the database
	for _, user := range db.Users {
		if user.Username == username {
			return user, nil
		}
	}

	return Credentials{}, fmt.Errorf("user not found")
}

// readDatabase reads the database from the file
func readDatabase(filePath string) (Database, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		// If the file doesn't exist, return an empty database
		if os.IsNotExist(err) {
			return Database{}, nil
		}
		return Database{}, fmt.Errorf("failed to read database file: %v", err)
	}

	var db Database
	err = json.Unmarshal(file, &db)
	if err != nil {
		return Database{}, fmt.Errorf("failed to unmarshal database: %v", err)
	}

	return db, nil
}

// writeDatabase writes the database to the file
func writeDatabase(db Database, filePath string) error {
	data, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal database: %v", err)
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write database file: %v", err)
	}

	return nil
}

func getCredentialsFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user's home directory: %v", err)
	}

	filePath := filepath.Join(homeDir, "credentials.json")
	return filePath, nil
}

func main() {
	prompt := promptui.Select{
		Label: "Select action",
		Items: []string{"Create New User", "Login as Existing User"},
	}

	_, result, _ := prompt.Run()

	filePath, err := getCredentialsFilePath()
	if err != nil {
		fmt.Printf("Failed to get credentials file path: %v\n", err)
		return
	}

	// Check if the credentials file exists
	_, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		// Create an empty database if the file doesn't exist
		db := Database{Users: []Credentials{}}
		err = writeDatabase(db, filePath)
		if err != nil {
			fmt.Printf("Failed to create credentials file: %v\n", err)
			return
		}
	}

	switch result {
	case "Create New User":
		usernamePrompt := promptui.Prompt{
			Label: "Enter username",
		}
		username, err := usernamePrompt.Run()
		if err != nil {
			fmt.Printf("Failed to read username: %v\n", err)
			return
		}

		passwordPrompt := promptui.Prompt{
			Label: "Enter password",
			Mask:  '*',
		}
		password, err := passwordPrompt.Run()
		if err != nil {
			fmt.Printf("Failed to read password: %v\n", err)
			return
		}

		credentials := Credentials{
			Username: username,
			Password: []byte(password),
		}

		err = SaveCredentials(credentials, filePath)
		if err != nil {
			fmt.Printf("Failed to save credentials: %v\n", err)
			return
		}

		fmt.Println("User created successfully!")

	case "Login as Existing User":
		// Check if the credentials file exists
		_, err := os.Stat(filePath)
		if os.IsNotExist(err) {
			fmt.Println("No existing user found. Please create a new user.")
			return
		}

		// File exists, check if it's older than 30 days
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			fmt.Printf("Failed to get file info: %v\n", err)
			return
		}

		expirationDate := fileInfo.ModTime().Add(30 * 24 * time.Hour)
		if time.Now().After(expirationDate) {
			err := DeleteFile(filePath)
			if err != nil {
				fmt.Printf("Failed to delete credentials file: %v\n", err)
				return
			}

			fmt.Println("Credentials expired! Please log in again.")
		} else {
			usernamePrompt := promptui.Prompt{
				Label: "Enter username",
			}
			username, err := usernamePrompt.Run()
			if err != nil {
				fmt.Printf("Failed to read username: %v\n", err)
				return
			}

			credentials, err := ReadCredentials(username, filePath)
			if err != nil {
				fmt.Printf("Failed to read credentials: %v\n", err)
				return
			}

			passwordPrompt := promptui.Prompt{
				Label: "Enter password",
				Mask:  '*',
			}
			password, err := passwordPrompt.Run()
			if err != nil {
				fmt.Printf("Failed to read password: %v\n", err)
				return
			}

			// Hash the entered password using Argon2
			hashedPassword := argon2.IDKey([]byte(password), []byte("somesalt"), 1, 64*1024, 4, 32)

			// Compare the hashed passwords
			if len(credentials.Password) != len(hashedPassword) {
				fmt.Println("Invalid password!")
				return
			}
			for i := range credentials.Password {
				if credentials.Password[i] != hashedPassword[i] {
					fmt.Println("Invalid password!")
					return
				}
			}

			// Passwords match, proceed with login
			fmt.Println("Login successful!")
			// TODO: Add your League of Legends logic here
		}
	}
}
