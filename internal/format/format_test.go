package format_test

import (
	"bytes"
  "context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/harrison-m-freitas/codectx/internal/cli"
	"github.com/harrison-m-freitas/codectx/internal/format"
	"github.com/harrison-m-freitas/codectx/internal/scan"
)

func TestFormatMarkdownAndFenced(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "a.go")
	_ = os.WriteFile(p, []byte("package main\nfunc main(){}\n"), 0o644)

	fm := scan.FileMeta{Path: p, Size: 24, Index: 0}
	cfg := cli.Config{Format: "markdown"}

	var w bytes.Buffer
	_ = format.WriteDocHeader(&w, cfg)
	_, _ = format.ProcessFiles(context.TODO(), &w, []scan.FileMeta{fm}, cfg, nil)
	out := w.String()
	if !strings.Contains(out, "## ") || !strings.Contains(out, "```") {
		t.Fatalf("markdown inválido:\n%s", out)
	}

	cfg.Format = "fenced"
	w.Reset()
	_ = format.WriteDocHeader(&w, cfg)
	_, _ = format.ProcessFiles(context.TODO(), &w, []scan.FileMeta{fm}, cfg, nil)
	out = w.String()
	if !strings.Contains(out, "```go") {
		t.Fatalf("fenced deveria conter ```go; got:\n%s", out)
	}
}
