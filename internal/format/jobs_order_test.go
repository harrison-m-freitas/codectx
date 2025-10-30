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

func TestJobsKeepsDeterministicOrder(t *testing.T) {
	dir := t.TempDir()
	paths := []string{"a.txt", "b.txt", "c.txt"}
	var metas []scan.FileMeta
	for i, p := range paths {
		fp := filepath.Join(dir, p)
		_ = os.WriteFile(fp, []byte(p+"\n"), 0o644)
		metas = append(metas, scan.FileMeta{Path: fp, Size: 3, Index: i})
	}
	cfg := cli.Config{Format: "plain", Jobs: 8}

	var w bytes.Buffer
	_ = format.WriteDocHeader(&w, cfg)
	_, _ = format.ProcessFiles(context.TODO(), &w, metas, cfg, nil)
	_ = format.WriteSummaryFooter(&w, cfg, 0, 0)

	out := w.String()
	// a ordem de cabeçalhos precisa seguir a ordem dos índices (a, b, c)
	idxA := strings.Index(out, "FILE: "+filepath.ToSlash(filepath.Join(dir, "a.txt")))
	idxB := strings.Index(out, "FILE: "+filepath.ToSlash(filepath.Join(dir, "b.txt")))
	idxC := strings.Index(out, "FILE: "+filepath.ToSlash(filepath.Join(dir, "c.txt")))
	if !(idxA >= 0 && idxB > idxA && idxC > idxB) {
		t.Fatalf("ordem não determinística com jobs>1")
	}
}
