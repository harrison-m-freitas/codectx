package scan_test

import (
	"context"
  "fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/harrison-m-freitas/codectx/internal/cli"
	"github.com/harrison-m-freitas/codectx/internal/logx"
	"github.com/harrison-m-freitas/codectx/internal/scan"
)

func TestMaxFilesTruncatesAndSignals(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < 5; i++ {
		name := fmt.Sprintf("f%c.txt", 'a'+i) // fa.txt, fb.txt, ...
		_ = os.WriteFile(filepath.Join(dir, name), []byte("x\n"), 0o644)
	}
	cfg := cli.Config{
		Paths:   []string{dir},
		MaxFiles: 2,
		Order:   "path",
		SecretsStrict: true,
		BinarySkip:    true,
	}
	log := logx.New()
	files, _, err := scan.List(context.TODO(), cfg, log)
	if err == nil {
		t.Fatal("esperava ErrMaxFilesExceeded")
	}
	if err != scan.ErrMaxFilesExceeded {
		t.Fatalf("erro sentinela esperado, got=%v", err)
	}
	if len(files) != 2 {
		t.Fatalf("deveria truncar para 2, got=%d", len(files))
	}
}
