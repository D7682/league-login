package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/go-vgo/robotgo"
	"github.com/spf13/cobra"
)

// Constants for file names
const (
	credentialsFileName = "credentials.json"
	defaultUserFileName = "default_user.txt"
)

// Constants for directory paths
var (
	programDirectory string // Directory where the program is located
	dataDirectory    string // Directory where data files are stored (credentials.json and default_user.txt)
)

// Credentials represents the user's login credentials
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Database represents the user credentials database
type Database struct {
	Users []Credentials `json:"users"`
}

// Initialize directory paths
func init() {
	ex, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	programDirectory = filepath.Dir(ex)
	dataDirectory = programDirectory
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

// saveDefaultUser saves the default user to a file in the data directory
func saveDefaultUser(username string) error {
	defaultUserFilePath := filepath.Join(dataDirectory, defaultUserFileName)
	err := os.WriteFile(defaultUserFilePath, []byte(username), 0644)
	if err != nil {
		return fmt.Errorf("failed to save default user: %v", err)
	}
	return nil
}

// getDefaultUser reads the default user from the default user file in the data directory
func getDefaultUser() (string, error) {
	defaultUserFilePath := filepath.Join(dataDirectory, defaultUserFileName)
	data, err := os.ReadFile(defaultUserFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read default user: %v", err)
	}
	return string(data), nil
}

func waitForWindow(title string, timeout time.Duration) bool {
	found := make(chan bool)

	go func() {
		for {
			hwnd := robotgo.FindWindow(title)
			if hwnd != 0 { // Check if the window handle is not 0
				found <- true
				return
			}
			time.Sleep(500 * time.Millisecond)
		}
	}()

	select {
	case <-found:
		return true
	case <-time.After(timeout):
		return false
	}
}

func main() {
	var rootCmd = &cobra.Command{Use: "league-login"}

	credFilePath, err := getCredentialsFilePath()
	if err != nil {
		fmt.Printf("Failed to get credentials file path: %v\n", err)
		return
	}
	// ...

	rootCmd.Run = func(cmd *cobra.Command, args []string) {
		defaultUser, err := getDefaultUser()
		if err != nil {
			fmt.Printf("Failed to get default user: %v\n", err)
			return
		}

		if defaultUser == "" {
			fmt.Println("No default user set. Please use flags or 'setdefault' to set a default user.")
			return
		}

		// Get the credentials of the default user
		credentials, err := ReadCredentials(defaultUser, credFilePath)
		if err != nil {
			fmt.Printf("Failed to get credentials of the default user: %v\n", err)
			return
		}

		// Now you have the credentials of the default user in the 'credentials' variable
		// You can use these credentials as needed
		fmt.Printf("Default user credentials: %+v\n", credentials)

		// Add your logic here to use the credentials
		// Replace with the actual path to your League of Legends executable
		// For example, on Windows, it might be something like:
		// "C:\\Riot Games\\League of Legends\\LeagueClient.exe"
		// On macOS, it could be:
		// "/Applications/League of Legends.app/Contents/LoL/LeagueClient.app/Contents/MacOS/LeagueClient"
		leagueClientPath := "C:\\Riot Games\\Riot Client\\RiotClientServices.exe"

		c := exec.Command(leagueClientPath, "--launch-product=league_of_legends", "--launch-patchline=live")
		err = c.Start()
		if err != nil {
			fmt.Printf("Failed to start League of Legends client: %v\n", err)
			return
		}

		time.Sleep(time.Millisecond * 500)
		robotgo.KeyTap("enter")

		if waitForWindow("Riot Client Main", 60*time.Second) {
			fmt.Println("Riot Client Main window found!")
			// Add your code to interact with the window here
		} else {
			fmt.Println("Timeout: Riot Client Main window not found")
		}

		robotgo.TypeStr(credentials.Username)
		robotgo.KeyTap("tab")
		robotgo.TypeStr(credentials.Password)
		robotgo.KeyTap("enter")
	}

	var newCmd = &cobra.Command{
		Use:   "new",
		Short: "Create a new user",
		Run: func(cmd *cobra.Command, args []string) {
			username, _ := cmd.Flags().GetString("username")
			password, _ := cmd.Flags().GetString("password")

			credentials := Credentials{
				Username: username,
				Password: password,
			}

			err = SaveCredentials(credentials, credFilePath)
			if err != nil {
				fmt.Printf("Failed to save credentials: %v\n", err)
				return
			}

			fmt.Println("User created successfully!")
		},
	}

	var setDefaultCmd = &cobra.Command{
		Use:   "setdefault",
		Short: "Set a user as default",
		Run: func(cmd *cobra.Command, args []string) {
			username := args[0] // Assuming the username is provided as an argument
			err := saveDefaultUser(username)
			if err != nil {
				fmt.Printf("Failed to set default user: %v\n", err)
				return
			}
			fmt.Printf("Default user set to: %s\n", username)
		},
	}

	// Add flags to the "new" command
	newCmd.Flags().StringP("username", "u", "", "Username")
	newCmd.Flags().StringP("password", "p", "", "Password")

	// Add commands to the root command
	rootCmd.AddCommand(newCmd, setDefaultCmd)

	rootCmd.Execute()
}

// getCredentialsFilePath returns the path to the credentials file
func getCredentialsFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user's home directory: %v", err)
	}

	filePath := filepath.Join(homeDir, credentialsFileName)
	return filePath, nil
}
