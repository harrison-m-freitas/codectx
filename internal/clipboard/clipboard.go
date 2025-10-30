package clipboard

import (
	"fmt"
	"os"
	"os/exec"

	cb "github.com/atotto/clipboard"
)

type Clip struct {
	source string // "github.com/atotto/clipboard" | "pbcopy" | "wl-copy" | ...
	log    interface {
		Info(string, ...any)
		Debug(string, ...any)
		Warn(string, ...any)
	}
}

func New(log interface {
	Info(string, ...any)
	Debug(string, ...any)
	Warn(string, ...any)
}) *Clip {
	return &Clip{source: "", log: log}
}

func (c *Clip) CopyFile(outputPath string, isStdout bool) error {
	if isStdout {
		return fmt.Errorf("clipboard com -o - não suportado; use -o <arquivo> para copiar")
	}
	data, err := os.ReadFile(outputPath)
	if err != nil {
		return err
	}

	if err := cb.WriteAll(string(data)); err == nil {
		c.source = "github.com/atotto/clipboard"
		return nil
	}

	for _, t := range tools {
    if c.copyViaTool(t, data) {
      return nil
    }
  }
	return fmt.Errorf("nenhuma ferramenta de clipboard disponível")
}

func (c *Clip) Source() string {
	if c.source == "" {
		return "desconhecido"
	}
	return c.source
}

type toolSpec struct {
  name string
  args []string
}

var tools = []toolSpec{
	{name: "pbcopy"},
	{name: "wl-copy"},
	{name: "xclip", args: []string{"-selection", "clipboard"}},
	{name: "xsel", args: []string{"-ib"}},
	{name: "clip.exe"},
}

func (c *Clip) copyViaTool(t toolSpec, data []byte) bool {
	cmd := exec.Command(t.name, t.args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return false
	}
	if err := cmd.Start(); err != nil {
		_ = stdin.Close()
		return false
	}
	if _, err := stdin.Write(data); err != nil {
		_ = stdin.Close()
		_ = cmd.Wait()
		return false
	}
	_ = stdin.Close()
	if err := cmd.Wait(); err != nil {
		return false
	}
	c.source = t.name
	return true
}
