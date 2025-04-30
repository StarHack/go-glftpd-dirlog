package dirlog

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestAddFindDeleteSaveLoad(t *testing.T) {
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "dirlog")
	dl, err := NewFile(file)
	if err != nil {
		t.Fatalf("NewFile failed: %v", err)
	}

	dir1 := filepath.Join(tmpDir, "sub1")
	if err := os.MkdirAll(dir1, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	f1 := filepath.Join(dir1, "f1.txt")
	if err := os.WriteFile(f1, []byte("hello"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	if err := dl.AddPath(dir1); err != nil {
		t.Fatalf("AddPath failed: %v", err)
	}
	if idx, rec := dl.FindByPath(dir1); rec == nil || idx != 0 {
		t.Fatalf("FindByPath returned idx=%d, rec=%v", idx, rec)
	} else if rec.Files != 1 {
		t.Errorf("expected 1 file, got %d", rec.Files)
	}

	if !dl.DeleteByPath(dir1) {
		t.Fatal("DeleteByPath failed")
	}
	if len(dl.Entries) != 0 {
		t.Errorf("expected 0 entries after delete, got %d", len(dl.Entries))
	}

	if err := dl.AddPath(dir1); err != nil {
		t.Fatalf("AddPath failed second time: %v", err)
	}
	if err := dl.SaveFile(file); err != nil {
		t.Fatalf("SaveFile failed: %v", err)
	}

	dl2, err := LoadFile(file)
	if err != nil {
		t.Fatalf("LoadFile failed: %v", err)
	}
	if len(dl2.Entries) != 1 {
		t.Errorf("expected 1 entry after reload, got %d", len(dl2.Entries))
	}
}

func TestPathsAndFindAll(t *testing.T) {
	dl := New()
	want := []string{"/foo", "/bar", "/foo/bar"}
	for _, p := range want {
		var rec DirLog
		if err := rec.SetPath(p); err != nil {
			t.Fatalf("SetPath(%q) failed: %v", p, err)
		}
		dl.Add(rec)
	}

	gotPaths := dl.Paths()
	if !reflect.DeepEqual(gotPaths, want) {
		t.Errorf("Paths() = %v; want %v", gotPaths, want)
	}

	idxs, logs := dl.FindAllByPath("foo")
	if len(idxs) != 2 || len(logs) != 2 {
		t.Fatalf("FindAllByPath(\"foo\") = %v, %v; want 2 matches", idxs, logs)
	}

	if idxs[0] != 0 || logs[0].Path() != "/foo" {
		t.Errorf("first match = idx %d, path %q; want idx 0, path \"/foo\"", idxs[0], logs[0].Path())
	}
	if idxs[1] != 2 || logs[1].Path() != "/foo/bar" {
		t.Errorf("second match = idx %d, path %q; want idx 2, path \"/foo/bar\"", idxs[1], logs[1].Path())
	}
}

func TestInfoFormatting(t *testing.T) {
	rec := DirLog{}
	rec.Status = 0
	rec.Uptime = int32(time.Now().Add(-26 * time.Hour).Unix())
	rec.Files = 5
	rec.Bytes = 5 * 1024 * 1024
	rec.SetPath("/some")
	info := rec.Info()
	wantSuffix := "(5F/5.1M/1d 2h)"
	if !strings.HasSuffix(info, wantSuffix) {
		t.Errorf("Info() = %q; want suffix %q", info, wantSuffix)
	}
}

func TestLoadFileError(t *testing.T) {
	tmp := t.TempDir()
	f := filepath.Join(tmp, "bad")
	os.WriteFile(f, []byte{0, 1, 2}, 0644)
	if _, err := LoadFile(f); err == nil {
		t.Errorf("LoadFile on bad file should error")
	}
}

func TestSetPathTooLong(t *testing.T) {
	var rec DirLog
	long := strings.Repeat("a", len(rec.Dirname)+1)
	if err := rec.SetPath(long); err == nil {
		t.Errorf("SetPath on too-long path should error")
	}
}
