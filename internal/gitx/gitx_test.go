package gitx

import (
  "os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGitAwareIfAvailable(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git não encontrado")
	}
	dir := t.TempDir()
	// init simples
	run := func(name string, args ...string) {
		cmd := exec.Command(name, args...)
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			t.Fatalf("falha ao rodar %s %v: %v", name, args, err)
		}
	}
	run("git", "init", "-q")
	run("git", "config", "user.email", "t@t.t")
	run("git", "config", "user.name", "t")
  fp := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(fp, []byte("hi\n"), 0o644); err != nil {
		t.Fatalf("falha ao criar arquivo de teste: %v", err)
	}
	run("git", "add", "a.txt")
	run("git", "commit", "-qm", "init")
	files, err := List(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("git ls-files não retornou arquivos")
	}
}
