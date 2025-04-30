package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/StarHack/go-glftpd-dirlog/dirlog"
)

func main() {
	filePath := flag.String("file", "/glftpd/ftp-data/logs/dirlog", "path to dirlog file")
	list := flag.Bool("list", false, "list dirlog entries")
	find := flag.String("find", "", "case-insensitive substring to find in paths")
	add := flag.String("add", "", "comma-separated paths to add")
	removeIndex := flag.String("remove-index", "", "comma-separated record numbers to remove (1-based)")
	removePath := flag.String("remove-path", "", "comma-separated paths to remove (case-insensitive match)")
	flag.Parse()

	dl, err := dirlog.LoadFile(*filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading dirlog: %v\n", err)
		os.Exit(1)
	}

	if *list {
		for i, rec := range dl.Entries {
			fmt.Printf("%5d: %s %s\n", i+1, rec.Path(), rec.Info())
		}
		return
	}

	if *find != "" {
		idxs, matches := dl.FindAllByPath(*find)
		if len(idxs) == 0 {
			fmt.Printf("no matches for %q\n", *find)
			return
		}
		for j, rec := range matches {
			fmt.Printf("%5d: %s\n", idxs[j]+1, rec.Path())
		}
		return
	}

	changed := false

	if *add != "" {
		for _, p := range strings.Split(*add, ",") {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			if err := dl.AddPath(p); err != nil {
				fmt.Fprintf(os.Stderr, "error adding path %q: %v\n", p, err)
				os.Exit(1)
			}
		}
		fmt.Printf("added paths: %s\n", *add)
		changed = true
	}

	if *removeIndex != "" {
		toRem := parseList(*removeIndex)
		for idx := range toRem {
			if !dl.DeleteAt(idx - 1) {
				fmt.Fprintf(os.Stderr, "no entry at index %d\n", idx)
			}
		}
		fmt.Printf("removed entries by index: %s\n", *removeIndex)
		changed = true
	}

	if *removePath != "" {
		for _, p := range strings.Split(*removePath, ",") {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			if dl.DeleteByPath(p) {
				fmt.Printf("removed path: %s\n", p)
			} else {
				fmt.Fprintf(os.Stderr, "path not found: %s\n", p)
			}
		}
		changed = true
	}

	if changed {
		if err := dl.SaveFile(*filePath); err != nil {
			fmt.Fprintf(os.Stderr, "error saving dirlog: %v\n", err)
			os.Exit(1)
		}
	}

	if !*list && *find == "" && !changed {
		flag.Usage()
	}
}

func parseList(s string) map[int]bool {
	m := make(map[int]bool)
	for _, part := range strings.Split(s, ",") {
		if _, err := fmt.Sscanf(strings.TrimSpace(part), "%d", new(int)); err == nil {
			var idx int
			fmt.Sscanf(strings.TrimSpace(part), "%d", &idx)
			if idx > 0 {
				m[idx] = true
			}
		}
	}
	return m
}
