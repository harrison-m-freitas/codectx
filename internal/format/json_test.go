package format_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/harrison-m-freitas/codectx/internal/cli"
	"github.com/harrison-m-freitas/codectx/internal/format"
	"github.com/harrison-m-freitas/codectx/internal/scan"
)

func TestJSONIndexOnlyArray(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "a.go")
	_ = os.WriteFile(fp, []byte("package x\nfunc main(){}\n"), 0o644)

	fm := scan.FileMeta{Path: fp, Size: 24, Index: 0}
	cfg := cli.Config{Format: "json", IndexOnly: true}

	var w bytes.Buffer
	if err := format.WriteDocHeader(&w, cfg); err != nil {
		t.Fatal(err)
	}
	if _, err := format.ProcessFiles(context.TODO(), &w, []scan.FileMeta{fm}, cfg, nil); err != nil {
		t.Fatal(err)
	}
	if err := format.WriteSummaryFooter(&w, cfg, 0, 0); err != nil {
		t.Fatal(err)
	}
	out := w.Bytes()
	var arr []map[string]any
	if err := json.Unmarshal(out, &arr); err != nil {
		t.Fatalf("JSON inválido: %v\n%s", err, string(out))
	}
	if len(arr) != 1 {
		t.Fatalf("esperava 1 item, got %d", len(arr))
	}
	if _, ok := arr[0]["content"]; ok {
		t.Fatalf("index-only: campo content não deveria existir")
	}
}

func TestNDJSONLines(t *testing.T) {
	dir := t.TempDir()
	fp1 := filepath.Join(dir, "a.go")
	fp2 := filepath.Join(dir, "b.go")
	_ = os.WriteFile(fp1, []byte("package a\n"), 0o644)
	_ = os.WriteFile(fp2, []byte("package b\n"), 0o644)

	files := []scan.FileMeta{
		{Path: fp1, Size: 10, Index: 0},
		{Path: fp2, Size: 10, Index: 1},
	}
	cfg := cli.Config{Format: "ndjson", IndexOnly: true}

	var w bytes.Buffer
	if err := format.WriteDocHeader(&w, cfg); err != nil {
		t.Fatal(err)
	}
	if _, err := format.ProcessFiles(context.TODO(), &w, files, cfg, nil); err != nil {
		t.Fatal(err)
	}
	if err := format.WriteSummaryFooter(&w, cfg, 0, 0); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(w.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("esperava 2 linhas NDJSON, got %d\n%s", len(lines), w.String())
	}
	for _, ln := range lines {
		var obj map[string]any
		if err := json.Unmarshal([]byte(ln), &obj); err != nil {
			t.Fatalf("linha NDJSON inválida: %v", err)
		}
	}
}
