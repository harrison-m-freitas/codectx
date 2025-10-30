package format_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/harrison-m-freitas/codectx/internal/cli"
	"github.com/harrison-m-freitas/codectx/internal/format"
	"github.com/harrison-m-freitas/codectx/internal/scan"
)

type errWriter struct{ n int }
func (e *errWriter) Write(p []byte) (int, error) {
	if len(p) > e.n {
		return 0, errors.New("forced write error")
	}
	return e.n, nil
}

func TestProcessFilesPropagatesWriteError(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "a.txt")
	_ = os.WriteFile(fp, []byte("body\n"), 0o644)
	files := []scan.FileMeta{{Path: fp, Size: 5, Index: 0}}
	cfg := cli.Config{Format: "plain"}

	// header ok
	var head bytes.Buffer
	_ = format.WriteDocHeader(&head, cfg)

	// writer que falha depois
	w := &errWriter{n: 0}
	_, err := format.ProcessFiles(context.TODO(), w, files, cfg, nil)
	if err == nil {
		t.Fatal("esperava erro de escrita")
	}
}
