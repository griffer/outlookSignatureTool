package main

import (
	"bufio"
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
	signatureBackupSrc := backupCommand.String("outlook", "", "Target outlook profile to backup from. (Optional).")
	signatureBackupDst := backupCommand.String("backup", "", "Target destination to store backup in. (Required).")

	// restore subcommand flag pointers
	signatureRestoreSrc := restoreCommand.String("backup", "", "Target backup folder to restore from. (Required)")
	signatureRestoreDst := restoreCommand.String("outlook", "", "Target outlook profile to restore to. (Optional)")

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

	// Evaluate which flags where passed to the backup subcommand
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
		databaseCheckIfExists(outlookDataPath)
		backupSignatures(databaseReadSignatures(outlookDataPath), outlookDataPath, outlookBackupDestinationPath)
	}

	// Evaluate which flags where passed to the restore subcommand
	if restoreCommand.Parsed() {
		if *signatureRestoreSrc == "" {
			restoreCommand.PrintDefaults()
			os.Exit(1)
		} else {
			outlookBackupDestinationPath = *signatureRestoreSrc
		}
		if *signatureRestoreDst == "" {
			// If no dst value is supplied revert to "Main Profile" standard location
			outlookDataPath = user.HomeDir + "/Library/Group Containers/UBF8T346G9.Office/Outlook/Outlook 15 Profiles/Main Profile/Data"
		} else {
			outlookDataPath = *signatureBackupSrc
		}
		databaseCheckIfExists(outlookDataPath)
		restoreSignatures(outlookDataPath, outlookBackupDestinationPath)
	}

}

func flagUsage() {
	fmt.Printf("Usage: %s [OPTIONS] argument ...\n", os.Args[0])
	flag.PrintDefaults()
}

// Rudimentary check of the signature backup file
func backupSignaturesVerify(outlookBackupDestinationPath string) {
	if _, err := os.Stat(outlookBackupDestinationPath + "/sql.txt"); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("sql.txt not found in provided backup path")
			panic(err)
		}
	}
}

func restoreSignatures(outlookDataPath string, outlookBackupDestinationPath string) {
	file, err := os.Open(outlookBackupDestinationPath + "/sql.txt")
	if err != nil {
		fmt.Println("sql.txt not found in provided backup path")
		panic(err)
	}
	defer file.Close()

	// Verify the signature backup files
	backupSignaturesVerify(outlookBackupDestinationPath)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() { // internally, it advances token based on sperator
		var split = strings.Split(scanner.Text(), "/")
		var recordID = split[0]
		var folderName = split[2]
		var signatureName = split[3]
		fmt.Println("Restoring signature: " + signatureName)
		// Create signature directory in the outlook Signatures directory
		createDirectory(outlookDataPath + "/Signatures/" + folderName)
		// Copy the signature to the outlook Signatures directory
		copyFile(outlookBackupDestinationPath+"/"+folderName+"/"+signatureName, outlookDataPath+"/Signatures/"+folderName+"/"+signatureName)
		// Write signature information to the Signatures table
		databaseWriteSignatures(outlookDataPath, folderName+"/"+signatureName, recordID)
	}
	// Updates the sqlite_sequence table with the new max recordID
	databaseUpdateSignaturesMaxRowID(outlookDataPath)
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
	return signatureSlice
}

// Write restored signature paths to the Signatures table
func databaseWriteSignatures(outlookDataPath string, pathToDataFile string, recordID string) {
	outlookDatabasePath := outlookDataPath + "/Outlook.sqlite"
	database, _ := sql.Open("sqlite3", outlookDatabasePath)
	statement, _ := database.Prepare("INSERT INTO Signatures (Record_RecordID, PathToDataFile) VALUES (?,?)")
	statement.Exec(recordID, "Signatures/"+pathToDataFile)
}

// databaseUpdateSignaturesMaxRowID updates seq in the sqlite_sequence table
// to reflect the highest RecordID found in the Signatures table
func databaseUpdateSignaturesMaxRowID(outlookDataPath string) {
	outlookDatabasePath := outlookDataPath + "/Outlook.sqlite"
	database, _ := sql.Open("sqlite3", outlookDatabasePath)
	// Get current signatures setup in outlook
	rows, _ := database.Query("SELECT Record_RecordID FROM Signatures ORDER BY Record_RecordID DESC LIMIT 1")
	var recordRecordID int
	for rows.Next() {
		rows.Scan(&recordRecordID)
	}
	// Updates seq field in the sqlite_sequence table with the Record_RecordID found above
	statement, _ := database.Prepare("UPDATE sqlite_sequence SET seq = ? WHERE name = 'Signatures'")
	statement.Exec(recordRecordID)
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
