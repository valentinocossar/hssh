package controllers

import (
	"fmt"
	"hssh/config"
	"hssh/templates"
	"io"
	"os"
	"path"

	"github.com/spf13/viper"
)

const configInitializationStatusCode = 2

// InitSSHConfig check if file ~/.ssh/config
// exist and create it if not
func initRequiredHomeSpaceFile(configPath string, template string) (int, error) {
	// Create needed folders if not exist
	err := os.MkdirAll(path.Dir(configPath), os.ModePerm)
	if err != nil {
		fmt.Printf("Error creating folders: %v.\n", err)
		return 1, err
	}

	// Create config file starting from template if not exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		file, err := os.Create(configPath)
		if err != nil {
			fmt.Printf("Error creating file: %v.\n", err)
			return 1, err
		}

		defer file.Close()
		file.WriteString(template)

		fmt.Printf("Created missing %v file!\n", configPath)
		return configInitializationStatusCode, nil
	}

	return 0, nil
}

func initHSSHHostFolder() error {
	err := os.MkdirAll(config.HSSHHostFolderPath, os.ModePerm)
	if err != nil {
		fmt.Printf("Error creating folders: %v.\n", err)
		return err
	}

	return nil
}

func isFolderEmpty(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}

// Init ...
func Init(showAllMessages bool) {

	// Check or create configuration file (config.yml)
	statusCode, err := initRequiredHomeSpaceFile(config.HSSHConfigFilePath, templates.Config)
	if err != nil {
		fmt.Println("An error occured during config.yml initialization.")
		os.Exit(1)
	}

	if statusCode == configInitializationStatusCode {
		fmt.Println("Before starting to use hssh edit the newly created configuration file.")
		os.Exit(0)
	}

	if showAllMessages == true {
		fmt.Println("The config.yml is initialized successfully.")
	}

	// Check or create configuration ssh file (.ssh/config)
	// If not exist the file will created empty
	_, err = initRequiredHomeSpaceFile(config.SSHConfigFilePath, "")
	if err != nil {
		fmt.Println("An error occured during ssh config initialization.")
		os.Exit(1)
	}

	if showAllMessages == true {
		fmt.Println("The .ssh/config is initialized successfully.")
	}

	// Search "config.yml" file in "$HOME/.config/hssh" directory.
	viper.SetConfigFile(config.HSSHConfigFilePath)

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config file: %v.\n", err)
		os.Exit(1)
	}

	// Check or create hssh host folder
	err = initHSSHHostFolder()
	if err != nil {
		fmt.Println("An error occured during the creation of the host folder.")
		os.Exit(1)
	}

	isEmpty, err := isFolderEmpty(config.HSSHHostFolderPath)
	if err != nil || isEmpty == true {
		fmt.Println("The host folder is empty.\nRunning the first sync...")
		Sync()
		os.Exit(0)
	}

	if showAllMessages == true {
		fmt.Println("The host folder exist and is not empty.")
	}
}