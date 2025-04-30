# go-glftpd-dirlog

Library for reading and writing new format glFTPD `dirlog` files (288-byte records).

- Add new dirs to dirlog
- Remove dirs from dirlog
- Search dirlog
- List entries in dirlog

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
