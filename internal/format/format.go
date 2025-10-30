package format

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/harrison-m-freitas/codectx/internal/cli"
	"github.com/harrison-m-freitas/codectx/internal/logx"
	"github.com/harrison-m-freitas/codectx/internal/scan"
	"github.com/harrison-m-freitas/codectx/internal/util"
)

type Metrics struct {
	Files int
	Bytes int64
}

func WriteDocHeader(w io.Writer, cfg cli.Config) error {
  if cfg.Format == "json" {
    _, err := io.WriteString(w, "[\n")
    return err
  }
  if cfg.Format == "ndjson" {
    return nil
  }
	paths := strings.Join(cfg.Paths, ":")
	_, err := fmt.Fprintf(w, "# Code Context\n# Generated: %s\n# Paths: %s\n", util.Now().Format("2006-01-02 15:04:05 -0700"), paths)
	if cfg.ExtCSV != "" {
		_, _ = fmt.Fprintf(w, "# Extensions: %s\n", cfg.ExtCSV)
	}
	if cfg.Depth > 0 {
		_, _ = fmt.Fprintf(w, "# Max Depth: %d\n", cfg.Depth)
	}
	_, _ = fmt.Fprintln(w)
	return err
}

func WriteSummaryFooter(w io.Writer, cfg cli.Config, skippedBin, skippedSec int) error {
  if cfg.Format == "json" {
    _, err := io.WriteString(w, "\n]\n")
    return err
  }
  if cfg.Format == "ndjson" {
    return nil
  }
	switch cfg.Format {
	case "markdown":
		_, err := fmt.Fprintf(w, "\n---\n**Resumo adicional:** binários ignorados: %d · arquivos sensíveis ignorados: %d\n", skippedBin, skippedSec)
		return err
	default:
		_, err := fmt.Fprintf(w, "\n---\nResumo adicional: binários ignorados=%d ; arquivos sensíveis ignorados=%d\n", skippedBin, skippedSec)
		return err
	}
}

func ProcessFiles(ctx context.Context, w io.Writer, files []scan.FileMeta, cfg cli.Config, log *logx.Logger) (*Metrics, error) {
	// Índices determinísticos já atribuídos
	type result struct {
		idx int
		buf []byte
		err error
		sz  int64
	}

	workers := cfg.Jobs
	if workers <= 0 {
		workers = runtime.GOMAXPROCS(0)
		if workers < 4 {
			workers = 4
		}
	}
	jobs := make(chan scan.FileMeta)
	out := make(chan result)

	var wg sync.WaitGroup
	worker := func() {
		defer wg.Done()
		for fm := range jobs {
      var (
        b      []byte
        bytesW int64
        err    error
      )
      switch cfg.Format {
      case "json":
        b, bytesW, err = renderOneJSON(fm, cfg)
      case "ndjson":
        line, _, e := renderOneJSON(fm, cfg)
        err = e
        if err == nil {
          // cada linha termina com \n
          b = append(line, '\n')
        }
      default:
        b, bytesW, err = renderOneText(fm, cfg)
      }
			out <- result{idx: fm.Index, buf: b, err: err, sz: bytesW}
		}
	}
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go worker()
	}

	go func() {
		for _, fm := range files {
			select {
			case <-ctx.Done():
				return
			case jobs <- fm:
			}
		}
		close(jobs)
		wg.Wait()
		close(out)
	}()

	// Reunião dos resultados por índice para impressão ordenada
	results := make([][]byte, len(files))
	var totalBytes int64
	for r := range out {
		if r.err != nil {
			return nil, r.err
		}
		results[r.idx] = r.buf
		totalBytes += r.sz
	}

	if cfg.Format == "json" {
    for i := range results {
      if results[i] == nil {
        continue
      }
      if i > 0 {
        if _, err := io.WriteString(w, ",\n"); err != nil {
          return nil, err
        }
      }
      if _, err := w.Write(results[i]); err != nil {
        return nil, err
      }
    }
  } else {
    for i := range results {
      if results[i] != nil {
        if _, err := w.Write(results[i]); err != nil {
          return nil, err
        }
      }
    }
  }
	return &Metrics{Files: len(files), Bytes: totalBytes}, nil
}

func renderOneText(fm scan.FileMeta, cfg cli.Config) ([]byte, int64, error) {
	pathOut := util.ToSlash(fm.Path)
	size := fm.Size
	hash, _ := util.Sha256Short8(fm.Path)
	lines := countLines(fm.Path)

	var b strings.Builder
  headerFor(&b, cfg, pathOut, size, hash, lines)

  written := 0
  if !cfg.IndexOnly {
    written, _ = writeBody(&b, fm.Path, cfg.MaxLines, cfg.MaxCols)
  }

  footerFor(&b, cfg)
	return []byte(b.String()), int64(written), nil
}

type jsonRec struct {
  Path    string `json:"path"`
  Size    int64  `json:"size"`
  Hash    string `json:"hash"`
  Lines   int    `json:"lines"`
  MTime   int64  `json:"mtime"`
  Ext     string `json:"ext"`
  Index   int    `json:"index"`
  Content string `json:"content,omitempty"`
}

func renderOneJSON(fm scan.FileMeta, cfg cli.Config) ([]byte, int64, error) {
  hash, _ := util.Sha256Short8(fm.Path)
  rec := jsonRec{
    Path:  util.ToSlash(fm.Path),
    Size:  fm.Size,
    Hash:  hash,
    Lines: countLines(fm.Path),
    MTime: fm.MTime,
    Ext:   strings.TrimPrefix(strings.ToLower(filepath.Ext(fm.Path)), "."),
    Index: fm.Index,
  }
  var written int
  if !cfg.IndexOnly {
    var sb strings.Builder
    w, _ := writeBody(&sb, fm.Path, cfg.MaxLines, cfg.MaxCols)
    written = w
    rec.Content = sb.String()
  }
  b, err := json.Marshal(rec)
  return b, int64(written), err
}

func headerFor(b *strings.Builder, cfg cli.Config, pathOut string, size int64, hash string, lines int) {
	switch cfg.Format {
	case "markdown":
		fmt.Fprintf(b, "\n## %s\n\n", pathOut)
		fmt.Fprintf(b, " - **Size:** %d bytes\n", size)
		fmt.Fprintf(b, " - **Hash:** %s\n", hash)
		fmt.Fprintf(b, " - **Lines:** %d\n\n", lines)
		b.WriteString("```\n")
	case "fenced":
		fmt.Fprintf(b, "\n```%s\n", fencedLang(pathOut))
		fmt.Fprintf(b, "# File: %s\n", pathOut)
		fmt.Fprintf(b, "# Size: %d bytes | Hash: %s | Lines: %d\n", size, hash, lines)
	default:
		b.WriteString("\n================================================================================\n")
		fmt.Fprintf(b, "FILE: %s\n", pathOut)
		fmt.Fprintf(b, "SIZE: %d bytes | HASH: %s | LINES: %d\n", size, hash, lines)
		b.WriteString("--------------------------------------------------------------------------------\n")
	}
}

func footerFor(b *strings.Builder, cfg cli.Config) {
	switch cfg.Format {
	case "markdown", "fenced":
		b.WriteString("\n```\n")
	default:
		b.WriteString("\n")
	}
}

func fencedLang(pathOut string) string {
	ext := strings.ToLower(filepath.Ext(pathOut))
	if len(ext) > 0 && ext[0] == '.' {
		return ext[1:]
	}
	return ext
}

func writeBody(b *strings.Builder, path string, maxLines, maxCols int) (int, error) {
	f, err := util.OpenRead(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	// aumenta limite de buffer para linhas longas
	const maxCap = 4 * 1024 * 1024
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, maxCap)

	lines := 0
	written := 0
	for sc.Scan() {
		line := sc.Text()
		if maxCols > 0 && len([]rune(line)) > maxCols {
			line = truncateCols(line, maxCols)
		}
		b.WriteString(line)
		b.WriteByte('\n')
		written += len(line) + 1
		lines++
		if maxLines > 0 && lines >= maxLines {
			fmt.Fprintf(b, "\n[... truncado em %d linhas ...]\n", maxLines)
			break
		}
	}
	return written, nil
}

func truncateCols(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	r = r[:max]
	return string(r) + "... [truncated]"
}

func countLines(path string) int {
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	lines := 0
	for sc.Scan() {
		lines++
	}
	return lines
}
