package filters_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/harrison-m-freitas/codectx/internal/cli"
	"github.com/harrison-m-freitas/codectx/internal/filters"
)

func mkfile(t *testing.T, dir, name string, data []byte) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestFilters_ExtIncludeExclude(t *testing.T) {
	dir := t.TempDir()
	a := mkfile(t, dir, "a.go", []byte("package x\n"))
	b := mkfile(t, dir, "node_modules/x.js", []byte("alert(1)\n"))
	c := mkfile(t, dir, "z.md", []byte("# x\n"))

	cfg := cli.Config{
		ExtCSV:        "go,md",
		Excludes:      []string{"node_modules"},
		Includes:      []string{},
		SecretsStrict: true,
		BinarySkip:    true,
	}
	if !filters.Decide(a, cfg).Include {
		t.Fatal("a.go deveria ser incluído")
	}
	if filters.Decide(b, cfg).Include {
		t.Fatal("node_modules deve ser excluído")
	}
	if !filters.Decide(c, cfg).Include {
		t.Fatal("z.md deveria ser incluído")
	}
}

func TestFilters_SecretsAndBinary(t *testing.T) {
	dir := t.TempDir()
	_ = mkfile(t, dir, ".env", []byte("SECRET=1\n"))
	bin := mkfile(t, dir, "bin.dat", []byte{0, 1, 2, 3})

	cfg := cli.Config{
		ExtCSV:        "",
		Excludes:      []string{},
		Includes:      []string{},
		SecretsStrict: true,
		BinarySkip:    true,
	}
	if filters.Decide(filepath.Join(dir, ".env"), cfg).Include {
		t.Fatal(".env deveria ser excluído (secreto)")
	}
	if filters.Decide(bin, cfg).Include {
		t.Fatal("binário deveria ser excluído")
	}
}

func TestFilters_MaxBytesAndIncludes(t *testing.T) {
	dir := t.TempDir()
	s := mkfile(t, dir, "src/main.py", []byte("x\n"))
	l := mkfile(t, dir, "big.txt", make([]byte, 1024))

	cfg := cli.Config{
		ExtCSV:        "",
		Excludes:      []string{},
		Includes:      []string{"src/"},
		MaxBytes:      512,
		SecretsStrict: false,
		BinarySkip:    false,
	}
	if !filters.Decide(s, cfg).Include {
		t.Fatal("src/main.py deveria entrar (include substring)")
	}
	if filters.Decide(l, cfg).Include {
		t.Fatal("big.txt deveria ser excluído por tamanho")
	}
}
