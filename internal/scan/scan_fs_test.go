package scan_test

import (
  "context"
	"os"
	"path/filepath"
	"testing"

	"github.com/harrison-m-freitas/codectx/internal/cli"
	"github.com/harrison-m-freitas/codectx/internal/logx"
	"github.com/harrison-m-freitas/codectx/internal/scan"
)

func TestFSDepthAndOrder(t *testing.T) {
	dir := t.TempDir()
	mk := func(p string) {
		fp := filepath.Join(dir, p)
		_ = os.MkdirAll(filepath.Dir(fp), 0o755)
		_ = os.WriteFile(fp, []byte(p+"\n"), 0o644)
	}
	mk("a/a1.txt")
	mk("a/a2.txt")
	mk("b/b1.txt")
	mk("b/c/c1.txt")

	cfg := cli.Config{
		Paths:         []string{dir},
		Depth:         1,
		ExtCSV:        "txt",
		Excludes:      []string{},
		Includes:      []string{},
		SecretsStrict: true,
		BinarySkip:    true,
		Order:         "path",
	}
	log := logx.New()
	list, _, err := scan.List(context.TODO(), cfg, log)
	if err != nil {
		t.Fatal(err)
	}
	// com depth=1 não deve incluir b/c/c1.txt
	foundDeep := false
	for _, fm := range list {
		if filepath.Base(fm.Path) == "c1.txt" {
			foundDeep = true
		}
	}
	if foundDeep {
		t.Fatal("c1.txt não deveria ser listado com depth=1")
	}
}
