package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"strings"

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
	outlookSignaturesPath := outlookDataPath + "/Signatures"
	outlookBackupDestinationPath := "/tmp/outlookBackup"

	databaseCheckIfExists(outlookDatabasePath)
	backupSignatures(databaseReadSignatures(outlookDatabasePath), outlookSignaturesPath, outlookBackupDestinationPath)
}

// backupSignatures queries the outlook database for active signatures, it then
// copies found signatures to the target destination
func backupSignatures(data []string, outlookSignaturesPath string, outlookBackupDestinationPath string) {
	for _, v := range data {
		var split = strings.Split(v, "/")
		var folderName = split[2]
		var signatureName = split[3]
		var signatureSourcePath = outlookSignaturesPath + "/" + folderName
		var signatureDestinationPath = outlookBackupDestinationPath + "/" + folderName
		fmt.Println("Backing up signature: " + signatureName)
		// Creates directory to store the signature
		createDirectory(signatureDestinationPath)
		// Copy signatures to backup destination
		copyFile(signatureSourcePath+"/"+signatureName, signatureDestinationPath+"/"+signatureName)
	}
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

// createDirectory creates directory at given path
func createDirectory(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0775)
	}
}

// TODO: Fail on file already exist
// copyFile copies file at src to dst
func copyFile(src string, dst string) {
	// Read file from src
	data, err := ioutil.ReadFile(src)
	if err != nil {
		panic(err)
	}
	// Write file to dst
	err = ioutil.WriteFile(dst, data, 0644)
	if err != nil {
		panic(err)
	}
}
