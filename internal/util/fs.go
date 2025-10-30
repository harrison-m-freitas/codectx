package util

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
  "strings"
	"time"
)

const noHash = "NO_HASH"

func ToSlash(p string) string {
	if p == "" {
		return ""
	}
	p = strings.ReplaceAll(p, "\\", "/")
	if filepath.Separator != '/' {
		p = strings.ReplaceAll(p, string(filepath.Separator), "/")
	}
	return p
}

func FileSize(path string) int64 {
	fi, err := fileInfo(path)
	if err != nil {
		return 0
	}
	return fi.Size()
}

func FileMTime(path string) int64 {
	fi, err := fileInfo(path)
	if err != nil {
		return 0
	}
	return fi.ModTime().Unix()
}

func IsBinary(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()
	buf := make([]byte, 4096)
	n, _ := f.Read(buf)
	if n <= 0 {
		return false, nil
	}
	for _, b := range buf[:n] {
		if b == 0 {
			return true, nil
		}
	}
	return false, nil
}

func Sha256Short8(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return noHash, err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return noHash, err
	}
	sum := h.Sum(nil)
	s := hex.EncodeToString(sum)
	if len(s) >= 8 {
		return s[:8], nil
	}
	return s, nil
}

func OpenRead(path string) (*os.File, error)   { return os.Open(path) }
func CreateWrite(path string) (*os.File, error) { return os.Create(path) }
func Now() time.Time { return time.Now() }
func EnsureDirAll(path string) error { return os.MkdirAll(path, 0o755) }
func Abs(path string) (string, error) { return filepath.Abs(path) }
func Join(elem ...string) string { return filepath.Join(elem...) }
func Base(path string) string { return filepath.Base(path) }
func HasPathPrefix(p, prefix string) bool { return strings.HasPrefix(p, prefix) }

func fileInfo(path string) (os.FileInfo, error) { return os.Stat(path) }
