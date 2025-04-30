package dirlog

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	RecordSize = 288
	PathOffset = 24
	PathMax    = 255
)

type DirLog struct {
	Data [RecordSize]byte
}

func (d *DirLog) Path() string {
	buf := d.Data[PathOffset:]
	if i := bytes.IndexByte(buf, 0); i >= 0 {
		return string(buf[:i])
	}
	return string(buf)
}

func (d *DirLog) SetPath(p string) error {
	if len(p) >= PathMax {
		return errors.New("path too long")
	}
	for i := 0; i < PathMax; i++ {
		d.Data[PathOffset+i] = 0
	}
	copy(d.Data[PathOffset:], p)
	return nil
}

type DirLogs struct {
	Entries []DirLog
}

func LoadFile(path string) (*DirLogs, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(b)%RecordSize != 0 {
		return nil, errors.New("invalid file size")
	}
	n := len(b) / RecordSize
	dl := &DirLogs{Entries: make([]DirLog, n)}
	for i := 0; i < n; i++ {
		copy(dl.Entries[i].Data[:], b[i*RecordSize:(i+1)*RecordSize])
	}
	return dl, nil
}

func (dl *DirLogs) SaveFile(path string) error {
	dir := filepath.Dir(path)
	tmp, err := ioutil.TempFile(dir, "dirlog.tmp")
	if err != nil {
		return err
	}
	for _, e := range dl.Entries {
		if _, err := tmp.Write(e.Data[:]); err != nil {
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
	for i, e := range dl.Entries {
		out[i] = e.Path()
	}
	return out
}

func (dl *DirLogs) FindByPath(p string) (int, *DirLog) {
	for i := range dl.Entries {
		if strings.EqualFold(dl.Entries[i].Path(), p) {
			return i, &dl.Entries[i]
		}
	}
	return -1, nil
}

func (dl *DirLogs) Add(entry DirLog) {
	dl.Entries = append(dl.Entries, entry)
}

func (dl *DirLogs) AddPath(p string) error {
	var e DirLog
	if err := e.SetPath(p); err != nil {
		return err
	}
	dl.Entries = append(dl.Entries, e)
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
		dl.DeleteAt(idx)
		return true
	}
	return false
}
