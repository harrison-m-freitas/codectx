package filters_test

import (
	"path/filepath"
	"testing"

	"github.com/harrison-m-freitas/codectx/internal/cli"
	"github.com/harrison-m-freitas/codectx/internal/filters"
)

func TestIncludeExcludeIgnoreCase(t *testing.T) {
	p := filepath.ToSlash("/x/Node_Modules/pkg/a.go")

	cfg := cli.Config{
		ExtCSV:          "GO",
		Excludes:        []string{"node_modules"},
		Includes:        []string{},
		SecretsStrict:   true,
		BinarySkip:      true,
		CaseInsensitive: true,
	}
	if filters.Decide(p, cfg).Include {
		t.Fatal("deveria excluir por exclude (case-insensitive)")
	}

	// agora remove exclude e usa include case-insensitive
	cfg.Excludes = nil
	cfg.Includes = []string{"PKG"}
	if !filters.Decide(p, cfg).Include {
		t.Fatal("deveria incluir por include (case-insensitive)")
	}
}
