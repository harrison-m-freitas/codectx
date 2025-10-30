package clipboard

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/harrison-m-freitas/codectx/internal/logx"
)

func TestClipboardSkipOnStdout(t *testing.T) {
	log := logx.New()
	cb := New(log)
	err := cb.CopyFile("-", true)
	if err == nil {
		t.Fatal("esperava erro ao copiar com stdout")
	}
}

func TestClipboardFileIfToolAvailable(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "out.txt")
	_ = os.WriteFile(fp, []byte("x"), 0o644)

	log := logx.New()
	cb := New(log)
	_ = cb.CopyFile(fp, false)
}
