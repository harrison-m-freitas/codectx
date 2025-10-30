package logx

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	LvlError = 0
	LvlWarn  = 1
	LvlInfo  = 2
	LvlDebug = 3
	LvlTrace = 4
)

type Logger struct {
	mu        sync.Mutex
	level     int
	withTS    bool
	colorMode string // auto|always|never
	jsonMode  bool
	filePath  string
	fileOut   io.Writer
	progName  string
}

const timeLayout = "2006-01-02 15:04:05"

func New() *Logger {
	return &Logger{
		level:     LvlInfo,
		withTS:    true,
		colorMode: "auto",
		jsonMode:  false,
		progName:  prog(),
	}
}

func (l *Logger) WithEnv() *Logger {
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		switch v {
		case "0", "1", "2", "3", "4":
			l.level = int(v[0] - '0')
		}
	}
	if v := os.Getenv("LOG_TS"); v == "0" {
		l.withTS = false
	}
	if v := os.Getenv("LOG_COLOR"); v != "" {
		l.colorMode = v
	}
	if v := os.Getenv("LOG_JSON"); v == "1" {
		l.jsonMode = true
	}
	if v := os.Getenv("LOG_FILE"); v != "" {
		l.filePath = v
	}
	return l
}
func (l *Logger) WithQuiet(q bool) *Logger {
	if q {
		l.level = LvlWarn
	}
	return l
}
func (l *Logger) WithVerbose(v bool) *Logger {
	if v {
		l.level = LvlDebug
	}
	return l
}

func (l *Logger) Init() error {
	if l.filePath != "" {
		f, err := os.OpenFile(l.filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			l.Warn("Não foi possível escrever em LOG_FILE='%s'", l.filePath)
			return nil
		}
		l.fileOut = f
	}
	l.Debug("Log inicializado | level=%d ts=%v color=%s json=%v file=%s",
		l.level, l.withTS, l.colorMode, l.jsonMode, or(l.filePath, "stderr"))
	return nil
}

func or[T ~string](s T, alt T) T {
	if s == "" {
		return alt
	}
	return s
}

func prog() string {
	if os.Args[0] == "" {
		return "codectx"
	}
	return filepath.Base(os.Args[0])
}

func (l *Logger) enabled(w int) bool { return l.level >= w }

func (l *Logger) log(level int, name, msg string, args ...any) {
	if !l.enabled(level) {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	text := fmt.Sprintf(msg, args...)
	now := time.Now()

	if l.jsonMode {
		l.writeJSON(now, name, text)
		return
	}

	ts := ""
	if l.withTS {
		ts = now.Format(timeLayout)
	}
	line := ""
	if ts != "" {
		line = fmt.Sprintf("[%s] %s %s: %s", ts, l.progName, name, text)
	} else {
		line = fmt.Sprintf("%s %s: %s", l.progName, name, text)
	}

	l.writeText(name, line)
}

func colorize(mode, name, line string) string {
	if mode == "never" {
		return line
	}
	if mode == "auto" {
		if os.Getenv("NO_COLOR") != "" {
			return line
		}
		fi, _ := os.Stderr.Stat()
		if (fi.Mode() & os.ModeCharDevice) == 0 {
			return line
		}
	}
	switch name {
	case "ERROR":
		return "\033[31m" + line + "\033[0m"
	case "WARN":
		return "\033[33m" + line + "\033[0m"
	case "DEBUG":
		return "\033[36m" + line + "\033[0m"
	case "TRACE":
		return "\033[2m" + line + "\033[0m"
	default:
		return line
	}
}

func (l *Logger) Error(msg string, a ...any) { l.log(LvlError, "ERROR", msg, a...) }
func (l *Logger) Warn(msg string, a ...any)  { l.log(LvlWarn, "WARN", msg, a...) }
func (l *Logger) Info(msg string, a ...any)  { l.log(LvlInfo, "INFO", msg, a...) }
func (l *Logger) Debug(msg string, a ...any) { l.log(LvlDebug, "DEBUG", msg, a...) }
func (l *Logger) Trace(msg string, a ...any) { l.log(LvlTrace, "TRACE", msg, a...) }

func (l *Logger) writeJSON(now time.Time, levelName, text string) {
	payload := map[string]any{
		"time":  now.Format(time.RFC3339),
		"prog":  l.progName,
		"level": levelName,
		"msg":   text,
	}
	enc, _ := json.Marshal(payload)
	fmt.Fprintln(os.Stderr, string(enc))
	if l.fileOut != nil {
		fmt.Fprintln(l.fileOut, string(enc))
	}
}

func (l *Logger) writeText(levelName, line string) {
	out := colorize(l.colorMode, levelName, line)
	fmt.Fprintln(os.Stderr, out)
	if l.fileOut != nil {
		fmt.Fprintln(l.fileOut, line)
	}
}
