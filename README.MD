# go-glftpd-dirlog

Library for reading and writing new format glFTPD `dirlog` files (288-byte records). The dirlog is the on-disk, fixed-size binary log that tracks every directory-related event (new uploads, nukes, unnukes, deletions) so the system and site-stats scripts can keep tabs on what’s happening. This is meant for sysops of glFTPD installations allowing them to edit the respective dirlog file.

## Features

- Add new dirs to dirlog
- Remove dirs from dirlog
- Search dirlog
- List entries in dirlog

Each record contains:

- **status**
  - `0 = NEWDIR`
  - `1 = NUKE`
  - `2 = UNNUKE`
  - `3 = DELETED`
- **uptime** creation or nuke timestamp (seconds since epoch)
- **uploader/group** the UID/GID of the user who created (or nuked) the directory
- **files** number of files in the directory at the moment of the event
- **bytes** total size of those files in bytes
- **dirname** full path to the directory
- **dummy/padding** reserved for compatibility

## Reference implementation

A sample implementation is available in `cmd/` dir.

```
go run github.com/StarHack/go-glftpd-dirlog/cmd@latest

Usage of ./cmd:
  -add string
    	comma-separated paths to add
  -add-all string
    	add all immediate subdirectories of this directory
  -file string
    	path to dirlog file (default "/glftpd/ftp-data/logs/dirlog")
  -find string
    	case-insensitive substring to find in paths
  -list
    	list dirlog entries
  -remove-index string
    	comma-separated record numbers to remove (1-based)
  -remove-path string
    	comma-separated paths to remove (case-insensitive match)
```

## Sample usage

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/StarHack/go-glftpd-dirlog/dirlog"
)

func main() {
	file := "/glftpd/ftp-data/logs/dirlog"
	dl, err := dirlog.LoadFile(file)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No existing dirlog, creating new one.")
			dl = dirlog.New()
		} else {
			log.Fatalf("failed to load dirlog: %v", err)
		}
	}

	fmt.Println("Current entries:")
	for i, rec := range dl.Entries {
		fmt.Printf("  %3d: %s %s\n", i+1, rec.Path(), rec.Info())
	}

	query := "myalbum"
	if idx, rec := dl.FindByPath(query); rec != nil {
		fmt.Printf("\nFound match for %q at #%d: %s %s\n",
			query, idx+1, rec.Path(), rec.Info())
	} else {
		fmt.Printf("\nNo match for %q\n", query)
	}

	newDir := "my-new-dir"
	if err := dl.AddPath(newDir); err != nil {
		log.Fatalf("AddPath failed: %v", err)
	}
	fmt.Printf("\nAdded/updated: %s\n", newDir)

	if err := dl.SaveFile(file); err != nil {
		log.Fatalf("SaveFile failed: %v", err)
	}
	fmt.Printf("Saved %d entries to %s\n", len(dl.Entries), file)
}
```
