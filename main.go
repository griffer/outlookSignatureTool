package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var outlookBackupDestinationPath string
var outlookDataPath string

func main() {
	// Get current user
	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	// Commandline menu
	// Subcommands
	backupCommand := flag.NewFlagSet("backup", flag.ExitOnError)
	restoreCommand := flag.NewFlagSet("restore", flag.ExitOnError)

	// backup subcommand flag pointers
	signatureBackupSrc := backupCommand.String("src", "", "Target outlook profile to backup (Optional).")
	signatureBackupDst := backupCommand.String("dst", "", "Destination of the signatures backup (Required).")

	// restore subcommand flag pointers
	restoreTextPtr := restoreCommand.String("src", "", "Target signature backup to restore. (Required)")

	// Verify that a subcommand has been provided
	if len(os.Args) < 2 {
		flag.Usage = flagUsage
		flag.Usage()
		fmt.Println("\"backup\" or \"restore\" subcommand is required")
		os.Exit(0)
	}

	switch os.Args[1] {
	case "backup":
		backupCommand.Parse(os.Args[2:])
	case "restore":
		restoreCommand.Parse(os.Args[2:])
	default:
		fmt.Println("\"backup\" or \"restore\" subcommand is required")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Evaluate which flags where passed to the subcommand
	if backupCommand.Parsed() {
		if *signatureBackupSrc == "" {
			// If no src value is supplied revert to "Main Profile" standard location
			outlookDataPath = user.HomeDir + "/Library/Group Containers/UBF8T346G9.Office/Outlook/Outlook 15 Profiles/Main Profile/Data"
		} else {
			outlookDataPath = *signatureBackupSrc
		}
		if *signatureBackupDst == "" {
			backupCommand.PrintDefaults()
			os.Exit(0)
		} else {
			outlookBackupDestinationPath = *signatureBackupDst
		}
	}

	if restoreCommand.Parsed() {
		if *restoreTextPtr == "" {
			restoreCommand.PrintDefaults()
			os.Exit(1)
		}
	}

	databaseCheckIfExists(outlookDataPath)
	backupSignatures(databaseReadSignatures(outlookDataPath), outlookDataPath, outlookBackupDestinationPath)

}

func flagUsage() {
	fmt.Printf("Usage: %s [OPTIONS] argument ...\n", os.Args[0])
	flag.PrintDefaults()
}

// backupSignatures queries the outlook database for active signatures, it then
// copies found signatures to the target destination
func backupSignatures(data []string, outlookDataPath string, outlookBackupDestinationPath string) {
	// Creates directory to store the signature backup
	createDirectory(outlookBackupDestinationPath)
	for _, v := range data {
		var split = strings.Split(v, "/")
		var folderName = split[2]
		var signatureName = split[3]
		var signatureSourcePath = outlookDataPath + "/Signatures/" + folderName
		var signatureDestinationPath = outlookBackupDestinationPath + "/" + folderName
		fmt.Println("Backing up signature: " + signatureName)
		// Creates directories for individual signatures
		createDirectory(signatureDestinationPath)
		// Copy signatures to backup destination
		copyFile(signatureSourcePath+"/"+signatureName, signatureDestinationPath+"/"+signatureName)
	}
	// Save signature information gathered from database to plaintext file
	printToFile(outlookBackupDestinationPath+"/"+"sql.txt", data)
}

// Check if the Outlook database is present
func databaseCheckIfExists(outlookDataPath string) {
	outlookDatabasePath := outlookDataPath + "/Outlook.sqlite"
	if _, err := os.Stat(outlookDatabasePath); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Outlook database not found")
			panic(err)
		}
	}
}

// databaseReadSignatures returns a slice of all signatures configured in outlook
func databaseReadSignatures(outlookDataPath string) []string {
	outlookDatabasePath := outlookDataPath + "/Outlook.sqlite"
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

// printToFile writes provided values to filePath
func printToFile(filePath string, values []string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, value := range values {
		fmt.Fprintln(f, value)
	}
	return nil
}
