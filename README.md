# outlookSignatureTool
_Backup and restore Microsoft Outlook signatures on macOS_

### Introduction
I looked for a tool to backup/restore Outlook 2019 signatures, and a came up short. After some trial and error, I came up with this slightly hack-ish way to do so.

### Usage

##### Backup
To backup signatures from Outlook run:

`outlookSignatureTool backup -backup /path/to/backup/directory`

You can pass in the optional _-outlook_ flag if you want to backup a non-default Outlook profile:

`outlookSignatureTool backup -backup /path/to/backup/directory -outlook /path/to/target/outlook/profile`

##### Restore
To restore signatures to Outlook run:

`outlookSignatureTool restore -backup /path/to/backup/directory`

You can pass in the optional _-outlook_ flag if you want to restore to a non-default Outlook profile:

`outlookSignatureTool restore -backup /path/to/backup/directory -outlook /path/to/target/outlook/profile`

#### How it works
To backup and restore signatures, we have to read/write to Outlooks embedded sqlite database.
The backup contains a _sql.txt_ file with information from the Signatures table, and the folder/file structure from Outlooks signatures folder.

The signature files seem to have some relationship to the database row id. They cannot be renamed or saved under another row id, or they will not work.

This means you cannot freely mix and match signatures from different Outlook backups, as you could risk row id's colliding.

#### Why not bash
Why write this in Go, instead of a few lines of Bash? Simple i wanted to learn a bit of Go :).

#### Notice
This is a work in progress. It works very well in my environment, but your luck may vary. The code has no logic to handle duplicate files in the backup location, so it will overwrite duplicates.