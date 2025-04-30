# go-glftpd-dirlog

Library for reading and writing new format glFTPD `dirlog` files (288-byte records).

## Reference implementation

A sample implementation is available in `cmd/` dir.

```
go run github.com/StarHack/go-glftpd-dirlog/cmd@latest

Usage of cmd:
  -add string
    	comma-separated paths to add
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
