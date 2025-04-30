package dirlog

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const RecordSize = 288

type DirLog struct {
	Status   uint16
	_pad0    [2]byte
	Uptime   int32
	Uploader uint16
	Group    uint16
	Files    uint16
	_pad1    [2]byte
	Bytes    uint64
	Dirname  [255]byte
	Dummy    [8]byte
}

func (d *DirLog) Path() string {
	if i := bytes.IndexByte(d.Dirname[:], 0); i >= 0 {
		return string(d.Dirname[:i])
	}
	return string(d.Dirname[:])
}

func (d *DirLog) SetPath(p string) error {
	if len(p) >= len(d.Dirname) {
		return errors.New("path too long")
	}
	copy(d.Dirname[:], p)
	for i := len(p); i < len(d.Dirname); i++ {
		d.Dirname[i] = 0
	}
	return nil
}

type DirLogs struct {
	Entries []DirLog
}

func LoadFile(path string) (*DirLogs, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(data)%RecordSize != 0 {
		return nil, errors.New("invalid file size")
	}
	buf := bytes.NewReader(data)
	entries, err := readAll(buf)
	if err != nil {
		return nil, err
	}
	return &DirLogs{Entries: entries}, nil
}

func (dl *DirLogs) SaveFile(path string) error {
	dir := filepath.Dir(path)
	tmp, err := ioutil.TempFile(dir, "dirlog.tmp")
	if err != nil {
		return err
	}
	if err := writeAll(tmp, dl.Entries); err != nil {
		tmp.Close()
		return err
	}
	name := tmp.Name()
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(name, path)
}

func readAll(r io.Reader) ([]DirLog, error) {
	var out []DirLog
	for {
		var rec DirLog
		err := binary.Read(r, binary.LittleEndian, &rec)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	return out, nil
}

func writeAll(w io.Writer, logs []DirLog) error {
	for _, rec := range logs {
		if err := binary.Write(w, binary.LittleEndian, &rec); err != nil {
			return err
		}
	}
	return nil
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
	var rec DirLog
	if err := rec.SetPath(p); err != nil {
		return err
	}
	dl.Entries = append(dl.Entries, rec)
	return nil
}

func (dl *DirLogs) DeleteByPath(p string) bool {
	idx, _ := dl.FindByPath(p)
	if idx < 0 {
		return false
	}
	dl.Entries = append(dl.Entries[:idx], dl.Entries[idx+1:]...)
	return true
}

func (dl *DirLogs) DeleteAt(index int) bool {
	if index < 0 || index >= len(dl.Entries) {
		return false
	}
	dl.Entries = append(dl.Entries[:index], dl.Entries[index+1:]...)
	return true
}

func (dl *DirLogs) Paths() []string {
	out := make([]string, len(dl.Entries))
	for i, rec := range dl.Entries {
		out[i] = rec.Path()
	}
	return out
}
