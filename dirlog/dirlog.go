package dirlog

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const RecordSize = 288

type DirLog struct {
	Status   uint16    //  0–1
	Pad0     [2]byte   //  2–3
	Uptime   int32     //  4–7
	Uploader uint16    //  8–9
	Group    uint16    // 10–11
	Files    uint16    // 12–13
	Pad1     [2]byte   // 14–15
	Bytes    uint64    // 16–23
	Dirname  [255]byte // 24–278
	Dummy    [9]byte   // 279–287 (8 bytes used + 1 pad to align to 4)
}

func (d *DirLog) Path() string {
	buf := d.Dirname[:]
	if i := bytes.IndexByte(buf, 0); i >= 0 {
		return string(buf[:i])
	}
	return string(buf)
}

func (d *DirLog) SetPath(p string) error {
	if len(p) >= len(d.Dirname) {
		return errors.New("path too long")
	}
	for i := range d.Dirname {
		d.Dirname[i] = 0
	}
	copy(d.Dirname[:], p)
	return nil
}

type DirLogs struct {
	Entries []DirLog
}

func LoadFile(path string) (*DirLogs, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(data)%RecordSize != 0 {
		return nil, errors.New("invalid file size")
	}
	buf := bytes.NewReader(data)
	var entries []DirLog
	for {
		var rec DirLog
		if err := binary.Read(buf, binary.LittleEndian, &rec); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		entries = append(entries, rec)
	}
	return &DirLogs{Entries: entries}, nil
}

func (dl *DirLogs) SaveFile(path string) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "dirlog.tmp")
	if err != nil {
		return err
	}
	for _, rec := range dl.Entries {
		if err := binary.Write(tmp, binary.LittleEndian, &rec); err != nil {
			tmp.Close()
			return err
		}
	}
	name := tmp.Name()
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(name, path)
}

func (dl *DirLogs) Paths() []string {
	out := make([]string, len(dl.Entries))
	for i, rec := range dl.Entries {
		out[i] = rec.Path()
	}
	return out
}

func (dl *DirLogs) FindByPath(query string) (int, *DirLog) {
	terms := strings.Fields(strings.ToLower(query))
	for i := range dl.Entries {
		pathLower := strings.ToLower(dl.Entries[i].Path())
		matched := true
		for _, term := range terms {
			if !strings.Contains(pathLower, term) {
				matched = false
				break
			}
		}
		if matched {
			return i, &dl.Entries[i]
		}
	}
	return -1, nil
}

func (dl *DirLogs) FindAllByPath(query string) ([]int, []*DirLog) {
	terms := strings.Fields(strings.ToLower(query))
	var idxs []int
	var logs []*DirLog
	for i := range dl.Entries {
		pathLower := strings.ToLower(dl.Entries[i].Path())
		ok := true
		for _, term := range terms {
			if !strings.Contains(pathLower, term) {
				ok = false
				break
			}
		}
		if ok {
			idxs = append(idxs, i)
			logs = append(logs, &dl.Entries[i])
		}
	}
	return idxs, logs
}

func (dl *DirLogs) Add(entry DirLog) {
	dl.Entries = append(dl.Entries, entry)
}

func (dl *DirLogs) AddPath(p string) error {
	var rec DirLog
	if err := rec.SetPath(p); err != nil {
		return err
	}
	dl.Entries = append(dl.Entries, rec)
	return nil
}

func (dl *DirLogs) DeleteAt(i int) bool {
	if i < 0 || i >= len(dl.Entries) {
		return false
	}
	dl.Entries = append(dl.Entries[:i], dl.Entries[i+1:]...)
	return true
}

func (dl *DirLogs) DeleteByPath(p string) bool {
	if idx, _ := dl.FindByPath(p); idx >= 0 {
		return dl.DeleteAt(idx)
	}
	return false
}
