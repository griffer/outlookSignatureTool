package main

import (
	"database/sql"
	"fmt"
	"os"
	"os/user"

	_ "github.com/mattn/go-sqlite3"
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
	databaseReadSignatures(outlookDatabasePath)
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

// databaseReadSignatures returns a slice of all signatures configured in outlook
func databaseReadSignatures(outlookDatabasePath string) []string {
	var signatureSlice []string
	database, _ := sql.Open("sqlite3", outlookDatabasePath)
	rows, _ := database.Query("SELECT Record_RecordID, PathToDataFile FROM Signatures")
	var PathToDataFile string
	var RecordID string
	for rows.Next() {
		rows.Scan(&RecordID, &PathToDataFile)
		// For now we combine the RecordID field from the database with the PathToDataFile field
		signatureSlice = append(signatureSlice, RecordID+"/"+PathToDataFile)
	}
	fmt.Printf("%v", signatureSlice)
	return signatureSlice
}
