package main

import (
	"context"
  "io"
	"fmt"
	"os"
	"time"
  "errors"
  "os/signal"
  "syscall"
  "path/filepath"

	"github.com/harrison-m-freitas/codectx/internal/cli"
	"github.com/harrison-m-freitas/codectx/internal/clipboard"
	"github.com/harrison-m-freitas/codectx/internal/format"
	"github.com/harrison-m-freitas/codectx/internal/logx"
	"github.com/harrison-m-freitas/codectx/internal/scan"
	"github.com/harrison-m-freitas/codectx/internal/util"
)

type outSink struct {
  f         *os.File
  isStdout  bool
  finalPath string
  tmpDir    string
  tmpFile   string
}

func openOutputAtomic(path string, log *logx.Logger) (*outSink, error) {
  if path == "-" {
    return &outSink{f: os.Stdout, isStdout: true, finalPath: "-"}, nil
  }
  dir := filepath.Dir(path)
  td, err := os.MkdirTemp(dir, ".codectx-*")
  if err != nil {
    return nil, err
  }
  tf := filepath.Join(td, "out.tmp")
  f, err := util.CreateWrite(tf)
  if err != nil {
    _ = os.RemoveAll(td)
    return nil, err
  }
  log.Info("Escrevendo (atômico) para: %s", path)
  return &outSink{f: f, finalPath: path, tmpDir: td, tmpFile: tf}, nil
}

func (o *outSink) Writer() io.Writer { return o.f }
func (o *outSink) IsStdout() bool { return o.isStdout }
func (o *outSink) Commit() error {
  if o.isStdout {
    return nil
  }
  if err := o.f.Close(); err != nil {
    return err
  }
  return os.Rename(o.tmpFile, o.finalPath)
}
func (o *outSink) Cleanup() {
  if o.isStdout {
    return
  }
  _ = os.RemoveAll(o.tmpDir)
}

func main() {
	cfg, show := cli.Parse(os.Args[1:])
	if show {
		fmt.Fprint(os.Stderr, cli.Help())
		os.Exit(0)
	}

	// Inicializa logger segundo flags e variáveis de ambiente
	log := logx.New().
		WithEnv().
		WithQuiet(cfg.Quiet).
		WithVerbose(cfg.Verbose)
	if err := log.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: falha ao inicializar logs: %v\n", err)
		os.Exit(1)
	}

	if err := cli.Validate(cfg); err != nil {
		log.Error("%v", err)
		os.Exit(1)
	}

	if cfg.Split {
		log.Error("Split mode ainda não implementado nesta versão refatorada")
		os.Exit(1)
	}

	// Resolve clipboard backend cedo (para mensagem de sucesso)
	cb := clipboard.New(log)

	// Caminho de saída
  out, err := openOutputAtomic(cfg.Output, log)
	if err != nil {
		log.Error("Falha ao criar arquivo de saída '%s': %v", cfg.Output, err)
		os.Exit(1)
	}
	defer out.Cleanup()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

  sigC := make(chan os.Signal, 1)
  signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM, syscall.SIGPIPE)
  go func() {
    <-sigC
    log.Warn("Sinal recebido — cancelando operação e limpando temporários.")
    cancel()
  }()

	start := time.Now()

	fileList, counters, err := scan.List(ctx, cfg, log)
  truncated := false
	if err != nil {
		if errors.Is(err, scan.ErrMaxFilesExceeded) {
			truncated = true
			log.Warn("Limite atingido (--max-files=%d). Processando %d arquivo(s).", cfg.MaxFiles, len(fileList))
		} else {
      log.Error("%v", err)
      os.Exit(1)
    }
  }

	if cfg.DryRun {
    err := doDryRun(out.Writer(), fileList, counters, start, log);
		if err != nil {
			log.Error("Falha no dry-run: %v", err)
			os.Exit(1)
		}
		if !out.IsStdout() {
			if err := out.Commit(); err != nil {
				log.Error("Falha ao finalizar saída: %v", err)
				os.Exit(1)
			}
		}
		if truncated {
      os.Exit(3)
    }
    return
	}

	// Cabeçalho do documento
	if err := format.WriteDocHeader(out.Writer(), cfg); err != nil {
		log.Error("Falha no cabeçalho do documento: %v", err)
		os.Exit(1)
	}

	// Processamento concorrente (determinístico na saída)
	metrics, err := format.ProcessFiles(ctx, out.Writer(), fileList, cfg, log)
	if err != nil {
		log.Error("%v", err)
		os.Exit(1)
	}

	// Rodapé com resumo adicional
  if err := format.WriteSummaryFooter(out.Writer(), cfg, counters.SkippedBin, counters.SkippedSecret); err != nil {
		log.Error("Falha no rodapé do documento: %v", err)
		os.Exit(1)
	}

  if !out.IsStdout() {
    if err := out.Commit(); err != nil {
      log.Error("Falha ao finalizar saída: %v", err)
      os.Exit(1)
    }
  }

	if cfg.Clipboard {
		source := "biblioteca"
    if err := cb.CopyFile(cfg.Output, out.isStdout); err != nil {
			log.Warn("Falha ao copiar para área de transferência: %v", err)
		} else {
			source = cb.Source()
			log.Info("Copiado para área de transferência via: %s", source)
		}
	}

	elapsed := time.Since(start)
	rate := 0.0
  if elapsed > 0 {
    rate = float64(metrics.Files) / elapsed.Seconds()
  }
  log.Info("Resumo: files=%d bytes=%d | binários_ignorados=%d | sensíveis_ignorados=%d | em %s | taxa=%.1f files/s",
    metrics.Files, metrics.Bytes, counters.SkippedBin, counters.SkippedSecret, elapsed, rate)

  if truncated {
    os.Exit(3)
  }
}

func doDryRun(w io.Writer, files []scan.FileMeta, cn *scan.Counters, start time.Time, log *logx.Logger) error {
	for _, fm := range files {
		if _, err := fmt.Fprintf(w, "  %s (%d bytes)\n", util.ToSlash(fm.Path), fm.Size); err != nil {
			return err
		}
	}
	elapsed := time.Since(start)
	rate := 0.0
	if elapsed > 0 {
		rate = float64(len(files)) / elapsed.Seconds()
	}
	log.Info("Resumo (DRY-RUN): files=%d bytes=%d | binários_ignorados=%d | sensíveis_ignorados=%d | em %s | taxa=%.1f files/s",
		len(files), cn.TotalBytes, cn.SkippedBin, cn.SkippedSecret, elapsed, rate)
	return nil
}
