package cli

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Paths         []string
	Depth         int
	ExtCSV        string
	Excludes      []string
	Includes      []string
	MaxBytes      int64
	MaxLines      int
	MaxCols       int
	Output        string
	Format        string // plain|markdown|fenced|json|ndjson
	Order         string // path|ext|size|mtime
	Split         bool
	DryRun        bool
	Quiet         bool
	Verbose       bool
	Clipboard     bool
	SecretsStrict bool
	BinarySkip    bool
	IndexOnly     bool
  Jobs          int
  CaseInsensitive bool
  MaxFiles      int
}

func defaults() Config {
	return Config{
		Depth:         0, // 0 = ilimitado
		ExtCSV:        "",
		Excludes:      splitCSV(".git,node_modules,.venv,dist,build,target,.next,.cache"),
		Includes:      []string{},
		MaxBytes:      0,
		MaxLines:      0,
		MaxCols:       0,
		Output:        "context.out",
		Format:        "plain",
		Order:         "path",
		Split:         false,
		DryRun:        false,
		Quiet:         false,
		Verbose:       false,
		Clipboard:     false,
		SecretsStrict: true,
		BinarySkip:    true,
    IndexOnly:     false,
    Jobs:          0, // 0 = auto (max(GOMAXPROCS, 4))
    CaseInsensitive: false,
    MaxFiles:      0,
	}
}

func addCSV(dst *string, csv string) {
  v := strings.TrimSpace(csv)
  if v == "" {
    return
  }
  if *dst == "" {
    *dst = v
    return
  }
  *dst = *dst + "," + v
}

func splitCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
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

func takeValue(haskV bool, val string, peek func() string) (string, int) {
  if haskV {
    return val, 1
  }
  return peek(), 2
}

func atoiOrZero(s string) int {
  n, _ := strconv.Atoi(s)
  return n
}

func atoi64OrZero(s string) int64 {
  n, _ := strconv.ParseInt(s, 10, 64)
  return n
}

type opt struct {
	needsValue bool
	setV       func(string)
	setB       func()
}

func appendCSV(dst *[]string, csv string) {
  *dst = append(*dst, splitCSV(csv)...)
}

func Parse(args []string) (Config, bool) {
	cfg := defaults()
	showHelp := false

  kv := func(set func(string)) opt { return opt{needsValue: true, setV: set} }
  bf := func(set func()) opt { return opt{needsValue: false, setB: set} }
  opts := map[string]opt{
    "-p": kv(func(v string) { cfg.Paths = append(cfg.Paths, v) }),
		"--path": kv(func(v string) { cfg.Paths = append(cfg.Paths, v) }),
		"-d": kv(func(v string) { cfg.Depth = atoiOrZero(v) }),
		"--depth": kv(func(v string) { cfg.Depth = atoiOrZero(v) }),
		"-e": kv(func(v string) { addCSV(&cfg.ExtCSV, v) }),
		"--ext": kv(func(v string) { addCSV(&cfg.ExtCSV, v) }),
		"-x": kv(func(v string) { appendCSV(&cfg.Excludes, v) }),
		"--exclude": kv(func(v string) { appendCSV(&cfg.Excludes, v) }),
		"-i": kv(func(v string) { appendCSV(&cfg.Includes, v) }),
		"--include": kv(func(v string) { appendCSV(&cfg.Includes, v) }),
		"-m": kv(func(v string) { cfg.MaxBytes = atoi64OrZero(v) }),
		"--max-bytes": kv(func(v string) { cfg.MaxBytes = atoi64OrZero(v) }),
		"--max-cols": kv(func(v string) { cfg.MaxCols = atoiOrZero(v) }),
		"-l": kv(func(v string) { cfg.MaxLines = atoiOrZero(v) }),
		"--max-lines": kv(func(v string) { cfg.MaxLines = atoiOrZero(v) }),
		"-o": kv(func(v string) { cfg.Output = v }),
		"--output": kv(func(v string) { cfg.Output = v }),
		"-F": kv(func(v string) { cfg.Format = v }),
		"--format": kv(func(v string) { cfg.Format = v }),
		"-O": kv(func(v string) { cfg.Order = v }),
		"--order": kv(func(v string) { cfg.Order = v }),
		"-j": kv(func(v string) { cfg.Jobs = atoiOrZero(v) }),
		"--jobs": kv(func(v string) { cfg.Jobs = atoiOrZero(v) }),
    "-N": kv(func(v string) { cfg.MaxFiles = atoiOrZero(v) }),
   "--max-files": kv(func(v string) { cfg.MaxFiles = atoiOrZero(v) }),
    "-I": bf(func() { cfg.CaseInsensitive = true }),
    "--ignore-case": bf(func() { cfg.CaseInsensitive = true }),
    "--index-only": bf(func() { cfg.IndexOnly = true }),
		"-c": bf(func() { cfg.Clipboard = true }),
		"-C": bf(func() { cfg.Clipboard = true }),
		"--clipboard": bf(func() { cfg.Clipboard = true }),
		"-S": bf(func() { cfg.Split = true }),
		"--split": bf(func() { cfg.Split = true }),
		"-R": bf(func() { cfg.DryRun = true }),
		"--dry-run": bf(func() { cfg.DryRun = true }),
		"-q": bf(func() { cfg.Quiet = true }),
		"--quiet": bf(func() { cfg.Quiet = true }),
		"-v": bf(func() { cfg.Verbose = true }),
		"--verbose": bf(func() { cfg.Verbose = true }),
		"--danger-include-secrets": bf(func() { cfg.SecretsStrict = false }),
		"--include-binaries": bf(func() { cfg.BinarySkip = false }),
		"-h": bf(func() { showHelp = true }),
		"--help": bf(func() { showHelp = true }),
  }

	i := 0
	for i < len(args) {
		a := args[i]
		peek := func() string {
			if i+1 < len(args) {
				return args[i+1]
			}
			return ""
		}
		shift := func(n int) { i += n }

		// --key=value atalho
		key, val, hasKV := strings.Cut(a, "=")

		if spec, ok := opts[key]; ok {
			if spec.needsValue {
				v, n := takeValue(hasKV, val, peek)
				spec.setV(v)
				shift(n)
				continue
			}
			spec.setB()
			shift(1)
			continue
		}
      if strings.HasPrefix(a, "-") {
        fmt.Fprintf(os.Stderr, "Opção desconhecida: %s. Use -h para ajuda.\n", a)
        showHelp = true
        shift(1)
        break
      }
      fmt.Fprintf(os.Stderr, "Argumento inesperado: %s. Use -h para ajuda.\n", a)
      showHelp = true
      shift(1)
	}
	return cfg, showHelp
}

func Validate(cfg Config) error {
	if len(cfg.Paths) == 0 {
		return errors.New("nenhum path especificado. Use -p <dir>")
	}
	switch cfg.Format {
	case "plain", "markdown", "fenced", "json", "ndjson":
	default:
		return fmt.Errorf("formato inválido: %s", cfg.Format)
	}
	switch cfg.Order {
	case "path", "ext", "size", "mtime":
	default:
		return fmt.Errorf("ordenação inválida: %s", cfg.Order)
	}
  if cfg.Jobs < 0 {
    return fmt.Errorf("--jobs deve ser >= 0")
  }
  if cfg.MaxFiles < 0 {
    return fmt.Errorf("--max-files deve ser >= 0")
  }
	return nil
}

func Help() string {
	return `USO: codectx [OPÇÕES]

DESCRIÇÃO:
Coleta contexto de código de um ou mais diretórios, gerando arquivo(s) com
conteúdo organizado, filtrado e formatado conforme especificações.

OPÇÕES:
-p, --path DIR         Diretório alvo (pode ser usado múltiplas vezes)
-d, --depth N          Profundidade máxima de recursão (0 = ilimitado)
-e, --ext CSV          Extensões incluídas (ex: "js,ts,py")
-x, --exclude PATTERN  Padrão de exclusão (pode repetir, CSV permitido)
-i, --include PATTERN  Padrão de inclusão (substring; CSV permitido)
-m, --max-bytes N      Ignorar arquivos maiores que N bytes
    --max-cols N       Truncar cada linha para no máximo N colunas (sanitização)
-l, --max-lines N      Limitar linhas por arquivo no contexto
-o, --output FILE      Arquivo de saída, "-" para stdout (padrão: context.out)
-c, --clipboard        Também copiar a saída final para a área de transferência
-S, --split            Gerar múltiplos arquivos por subdiretório (NÃO IMPLEMENTADO)
-F, --format TYPE      Formato: plain|markdown|fenced (padrão: plain)
-O, --order TYPE       Ordenação: path|ext|size|mtime (padrão: path)
-j, --jobs N           Número de jobs paralelos (0 = auto, padrão: 0)
-N, --max-files N      Número máximo de arquivos a processar (0 = ilimitado)
-I, --ignore-case      Tornar filtros de inclusão/exclusão case-insensitive
    --index-only       Apenas gerar índice de arquivos, sem conteúdo
-R, --dry-run          Apenas listar o que seria incluído
-q, --quiet            Menos logs
-v, --verbose          Mais logs
    --danger-include-secrets Incluir arquivos sensíveis (não recomendado)
    --include-binaries       Incluir arquivos binários (não recomendado)
-h, --help             Mostrar esta ajuda

LOGS (variáveis de ambiente):
LOG_LEVEL=0..4 (0=ERROR..4=TRACE), LOG_TS=0|1, LOG_COLOR=auto|always|never,
LOG_JSON=0|1, LOG_FILE=/caminho/arquivo.log, NO_COLOR=1 desativa cor (auto).

EXEMPLOS:
# Múltiplos diretórios
./codectx -p src -p tests -F markdown

# Com filtros por extensão e ordenação por tamanho
./codectx -p . -e "ts,tsx,go" -O size -o context.out

# Dry-run
./codectx -p lib -p bin -p docs --dry-run

# Para stdout com clipboard
./codectx -p src -p tests -o - -C -F fenced

# JSON/NDJSON para pipelines
./codectx -p . -F json --index-only > index.json
./codectx -p . -F ndjson --index-only | jq -c '.path'

# Flags repetíveis/case-insensitive
./codectx -p . -e go -e "md,py" -x node_modules -x ".cache,.venv" -I
`
}
