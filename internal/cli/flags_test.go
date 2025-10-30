package cli_test

import (
	"testing"
	"github.com/harrison-m-freitas/codectx/internal/cli"
)

func TestExtRepeatableAndCSV(t *testing.T) {
	args := []string{"-e", "go", "-e", "MD,Py"}
	cfg, _ := cli.Parse(args)
	// a ordem não importa; splitCSV será feito nos filtros
	if cfg.ExtCSV == "" || !(len(cfg.ExtCSV) >= 4) {
		t.Fatalf("ExtCSV deveria acumular, got=%q", cfg.ExtCSV)
	}
}

func TestIgnoreCaseFlag(t *testing.T) {
	args := []string{"-I", "-x", "Node_Modules"}
	cfg, _ := cli.Parse(args)
	if !cfg.CaseInsensitive {
		t.Fatal("CaseInsensitive esperado true")
	}
	found := false
	for _, ex := range cfg.Excludes {
		if ex == "Node_Modules" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("esperava 'Node_Modules' nos excludes; got=%v", cfg.Excludes)
	}
}
