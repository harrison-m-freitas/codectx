package scan_test

import (
	"bytes"
  "context"
	"os"
	"path/filepath"
	"testing"

	"github.com/harrison-m-freitas/codectx/internal/cli"
	"github.com/harrison-m-freitas/codectx/internal/format"
	"github.com/harrison-m-freitas/codectx/internal/logx"
	"github.com/harrison-m-freitas/codectx/internal/scan"
)

func TestDryRunPrintsList(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "a.go")
	_ = os.WriteFile(fp, []byte("package x\n"), 0o644)

	cfg := cli.Config{
		Paths:         []string{dir},
		ExtCSV:        "go",
		Excludes:      []string{},
		Includes:      []string{},
		SecretsStrict: true,
		BinarySkip:    true,
		Order:         "path",
		DryRun:        true,
		Output:        "-",
	}
	log := logx.New()
	files, cn, err := scan.List(context.TODO(), cfg, log)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	_ = format.WriteDocHeader(&buf, cfg) // apenas para ter algo no buffer (não exigido em dry-run real)
	// simulamos dry-run imprimindo manualmente (main faz isso)
	for _, fm := range files {
		buf.WriteString(fm.Path)
	}
	if !bytes.Contains(buf.Bytes(), []byte("a.go")) {
		t.Fatal("dry-run deveria listar a.go")
	}
	if cn.TotalBytes == 0 {
		t.Fatal("counters.TotalBytes esperado > 0")
	}
}
