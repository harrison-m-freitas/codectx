package filters

import (
	"path/filepath"
	"strings"

	"github.com/harrison-m-freitas/codectx/internal/cli"
	"github.com/harrison-m-freitas/codectx/internal/util"
)

var DefaultSecretExcludes = []string{
	".env", ".env.", // .env e .env.*
	".pem", ".key", ".p12", ".pfx", ".asc", ".gpg",
	".git-credentials", ".npmrc", ".pypirc", ".s3cfg", ".boto",
	".azure", ".aws", ".ssh",
	".kdbx", ".keystore", ".jks",
}

type Decision struct {
	Include bool
	Reason  string
}

func Decide(path string, cfg cli.Config) Decision {
	pathSlashed := util.ToSlash(path)

  // 1) Extensão (CSV permitido)
  if cfg.ExtCSV != "" && !hasAllowedExt(pathSlashed, cfg.ExtCSV) {
		return Decision{false, "ext"}
	}

  // 2) Excludes (substring, CSV permitido)
  if isExcludedPath(pathSlashed, cfg.Excludes, cfg.CaseInsensitive) {
		return Decision{false, "exclude"}
	}

	// 3) Segredos
	if cfg.SecretsStrict && isSensitiveFile(pathSlashed) {
		return Decision{false, "secret"}
	}

	// 4) Binários (NUL)
	if cfg.BinarySkip {
		if bin, _ := util.IsBinary(path); bin {
			return Decision{false, "binary"}
		}
	}

  // 5) Includes (caminho completo)
  if len(cfg.Includes) > 0 && !isIncludedPath(pathSlashed, cfg.Includes, cfg.CaseInsensitive) {
		return Decision{false, "include"}
	}

	// 6) Tamanho
	if cfg.MaxBytes > 0 && util.FileSize(path) > cfg.MaxBytes {
		return Decision{false, "size"}
	}

	return Decision{true, "ok"}
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func hasAllowedExt(pathSlashed string, csv string) bool {
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(pathSlashed)), ".")
	if ext == "" {
		return false
	}
	for _, e := range splitCSV(csv) {
		if strings.ToLower(e) == ext {
			return true
		}
	}
	return false
}

func isExcludedPath(pathSlashed string, excludes []string, insensitive bool) bool {
	if len(excludes) == 0 {
		return false
	}
	ps := pathSlashed // alias
  if insensitive {
    ps = strings.ToLower(pathSlashed)
  }
	for _, p := range excludes {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
    needle := p
    if insensitive {
      needle = strings.ToLower(needle)
    }
		if strings.ContainsAny(needle, "*/?") || strings.Contains(needle, "/") {
			if strings.Contains(ps, needle) {
				return true
			}
			continue
		}
		needle = "/" + needle + "/"
		if strings.Contains(ps+"/", needle) || strings.HasSuffix(ps, "/"+needle) {
			return true
		}
	}
	return false
}

func isIncludedPath(pathSlashed string, includes []string, insensitive bool) bool {
  ps := pathSlashed
  if insensitive {
    ps = strings.ToLower(ps)
  }
  for _, inc := range includes {
    if inc == "" {
       continue
     }
    needle := inc
    if insensitive {
      needle = strings.ToLower(needle)
    }
    if strings.Contains(ps, needle) {
       return true
     }
   }
   return false
}

func isSensitiveFile(path string) bool {
	if strings.Contains(path, "/.ssh/") ||
		strings.Contains(path, "/.aws/") ||
		strings.Contains(path, "/.azure/") ||
		strings.Contains(path, "/.gnupg/") ||
		strings.Contains(path, "/.secrets/") {
		return true
	}
	base := util.Base(path)
	if base == ".env" || strings.HasPrefix(base, ".env.") {
		return true
	}
	lowBase := strings.ToLower(base)
	for _, suf := range []string{".pem", ".key", ".p12", ".pfx", ".asc", ".gpg", ".kdbx", ".keystore", ".jks"} {
		if strings.HasSuffix(lowBase, suf) {
			return true
		}
	}
	if base == ".git-credentials" || base == ".npmrc" || base == ".pypirc" || base == "composer.auth.json" {
		return true
	}
	// substrings típicas
	l := strings.ToLower(path)
	if strings.Contains(l, "secret") || strings.Contains(l, "secrets") ||
		strings.Contains(l, "token") || strings.Contains(l, "apikey") ||
		strings.Contains(l, "api_key") || strings.Contains(l, "password") ||
		strings.Contains(l, "passwd") {
		return true
	}
	return false
}
