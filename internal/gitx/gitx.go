package gitx

import (
	"bytes"
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
)

func hasGit() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

func isInsideRepo(path string) bool {
  cmd := gitCmd(path, "rev-parse", "--is-inside-work-tree")
	return cmd.Run() == nil
}

func RepoRoot(path string) (string, error) {
	if !hasGit() {
		return "", errors.New("git não encontrado")
	}
  cmd := gitCmd(path, "rev-parse", "--show-toplevel")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", errors.New("não é repositório git")
	}
	return strings.TrimSpace(out.String()), nil
}

func List(path string) ([]string, error) {
	if !hasGit() || !isInsideRepo(path) {
		return nil, errors.New("git indisponível ou caminho fora de repo")
	}
	root, err := RepoRoot(path)
	if err != nil {
		return nil, err
	}
	rel, rerr := filepath.Rel(root, path)
	if rerr != nil || rel == "" {
		rel = "."
	}
	cmd := gitCmd(root, "ls-files", "-co", "--exclude-standard", "--", rel)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return mapFiles(root, splitLines(out.String())), nil
}

func gitCmd(dir string, args ...string) *exec.Cmd {
	all := append([]string{"-C", dir}, args...)
	return exec.Command("git", all...)
}

func splitLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.Split(s, "\n")
}

func mapFiles(root string, lines []string) []string {
	var files []string
	for _, f := range lines {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		files = append(files, filepath.Join(root, f))
	}
	return files
}
