package main

import (
	"fmt"
	"os"
	"os/user"
)

func main() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	// Placeholder paths
	outlookDataPath := user.HomeDir + "/Library/Group Containers/UBF8T346G9.Office/Outlook/Outlook 15 Profiles/Main Profile/Data"
	outlookDatabasePath := outlookDataPath + "/Outlook.sqlite"
	databaseCheckIfExists(outlookDatabasePath)
}

// Check if the Outlook database is present
func databaseCheckIfExists(outlookDatabasePath string) {
	if _, err := os.Stat(outlookDatabasePath); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Outlook database not found")
			panic(err)
		}
	}
}
