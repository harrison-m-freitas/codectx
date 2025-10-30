package scan

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/harrison-m-freitas/codectx/internal/cli"
	"github.com/harrison-m-freitas/codectx/internal/filters"
	"github.com/harrison-m-freitas/codectx/internal/gitx"
	"github.com/harrison-m-freitas/codectx/internal/logx"
	"github.com/harrison-m-freitas/codectx/internal/util"
)

var ErrMaxFilesExceeded = errors.New("max-files exceeded")

type FileMeta struct {
	Path  string
	Size  int64
	MTime int64
	Ext   string
	Key   string // chave de ordenação
	Index int    // posição determinística pós-sort
}

type Counters struct {
	SkippedBin    int
	SkippedSecret int
	TotalBytes    int64
}

func List(ctx context.Context, cfg cli.Config, log *logx.Logger) ([]FileMeta, *Counters, error) {
	var all []string
	for _, p := range cfg.Paths {
		abs, _ := filepath.Abs(p)
		files, err := listPath(abs, cfg, log)
		if err != nil {
			return nil, nil, err
		}
		all = append(all, files...)
	}

	cn := &Counters{}
	selected := make([]FileMeta, 0, len(all))
	for _, fp := range all {
		sz := util.FileSize(fp)
		cn.TotalBytes += sz
		d := filters.Decide(fp, cfg)
		if !d.Include {
			switch d.Reason {
			case "binary":
				cn.SkippedBin++
			case "secret":
				cn.SkippedSecret++
			}
			continue
		}
		selected = append(selected, FileMeta{
			Path:  fp,
			Size:  sz,
			MTime: util.FileMTime(fp),
			Ext:   extLower(fp),
		})
	}

	switch cfg.Order {
	case "path":
		sort.Slice(selected, func(i, j int) bool {
			return util.ToSlash(selected[i].Path) < util.ToSlash(selected[j].Path)
		})
	case "ext":
		sort.Slice(selected, func(i, j int) bool {
			if selected[i].Ext == selected[j].Ext {
				return util.ToSlash(selected[i].Path) < util.ToSlash(selected[j].Path)
			}
			return selected[i].Ext < selected[j].Ext
		})
	case "size":
		sort.Slice(selected, func(i, j int) bool {
			if selected[i].Size == selected[j].Size {
				return util.ToSlash(selected[i].Path) < util.ToSlash(selected[j].Path)
			}
			return selected[i].Size < selected[j].Size
		})
	case "mtime":
		sort.Slice(selected, func(i, j int) bool {
			if selected[i].MTime == selected[j].MTime {
				return util.ToSlash(selected[i].Path) < util.ToSlash(selected[j].Path)
			}
			return selected[i].MTime < selected[j].MTime
		})
	default:
		return nil, nil, errors.New("ordenação inválida")
	}

  var limErr error
  if cfg.MaxFiles > 0 && len(selected) > cfg.MaxFiles {
    selected = selected[:cfg.MaxFiles]
    limErr = ErrMaxFilesExceeded
  }

	for i := range selected {
		selected[i].Index = i
	}
	return selected, cn, limErr
}

func extLower(p string) string {
	e := filepath.Ext(p)
	if e == "" {
		return ""
	}
	return strings.TrimPrefix(strings.ToLower(e), ".")
}

func listPath(path string, cfg cli.Config, log *logx.Logger) ([]string, error) {
	if files, err := gitx.List(path); err == nil && len(files) > 0 {
		log.Debug("git-aware ativo em: %s", path)
		return absAll(files), nil
	}

  var res []string
	maxDepth := cfg.Depth
	rootDepth := depthOf(path)

	err := filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		relDepth := depthOf(p) - rootDepth
		if maxDepth > 0 && relDepth > maxDepth {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			if isExcludedDir(util.Base(p), cfg.Excludes, cfg.CaseInsensitive) {
        return filepath.SkipDir
      }
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		res = append(res, p)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return absAll(res), nil
}

func isExcludedDir(base string, excludes []string, insensitive bool) bool {
  b := base
  if insensitive {
    b = strings.ToLower(b)
  }
	for _, ex := range excludes {
		ex = strings.TrimSpace(ex)
		if ex == "" || strings.ContainsAny(ex, "*/?/") {
			continue
	}
    needle := ex
    if insensitive {
      needle = strings.ToLower(needle)
    }
		if b == needle {
			return true
		}
	}
	return false
}

func absAll(l []string) []string {
	out := make([]string, 0, len(l))
	for _, p := range l {
		ap, err := filepath.Abs(p)
		if err == nil {
			out = append(out, ap)
		}
	}
	return out
}

func depthOf(p string) int {
	n := 0
	for _, r := range filepath.ToSlash(p) {
		if r == '/' {
			n++
		}
	}
	return n
}
