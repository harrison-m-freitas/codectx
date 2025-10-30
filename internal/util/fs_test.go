package util_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/harrison-m-freitas/codectx/internal/util"
)

func TestFileSizeAndMTime(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "a.txt")
	_ = os.WriteFile(fp, []byte("abc"), 0o644)

	if sz := util.FileSize(fp); sz != 3 {
		t.Fatalf("FileSize=3, got %d", sz)
	}
	if mt := util.FileMTime(fp); mt <= 0 {
		t.Fatalf("FileMTime esperado >0, got %d", mt)
	}
}

func TestIsBinary(t *testing.T) {
	dir := t.TempDir()
	txt := filepath.Join(dir, "t.txt")
	bin := filepath.Join(dir, "b.bin")
	_ = os.WriteFile(txt, []byte("hello"), 0o644)
	_ = os.WriteFile(bin, []byte{0x01, 0x00, 0x02}, 0o644)

	isBin, err := util.IsBinary(txt)
	if err != nil || isBin {
		t.Fatalf("texto não deve ser binário; err=%v isBin=%v", err, isBin)
	}
	isBin, err = util.IsBinary(bin)
	if err != nil || !isBin {
		t.Fatalf("binário deve ser true; err=%v isBin=%v", err, isBin)
	}
}

func TestSha256Short8(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "h.txt")
	_ = os.WriteFile(fp, []byte("abc\n"), 0o644)

	h, err := util.Sha256Short8(fp)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(h) != 8 {
		t.Fatalf("hash curto com 8 chars; got %q (%d)", h, len(h))
	}
}

func TestPrefixAndSlash(t *testing.T) {
	if !util.HasPathPrefix("/a/b", "/a/") {
		t.Fatal("prefixo deveria ser true")
	}
	if util.HasPathPrefix("/x/b", "/a/") {
		t.Fatal("prefixo deveria ser false")
	}
	if s := util.ToSlash(`a\b\c`); !strings.Contains(s, "/") {
		t.Fatalf("ToSlash deveria normalizar: %q", s)
	}
}

func TestNowMonotonicity(t *testing.T) {
	t1 := util.Now()
	time.Sleep(5 * time.Millisecond)
	t2 := util.Now()
	if !t2.After(t1) && !t2.Equal(t1) {
		t.Fatal("Now deveria ser não decrescente")
	}
}
