package exporter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	logMu          sync.Mutex
	onceInitLog    sync.Once
	currentLogFile string
	initLogErr     error
	logFile        *os.File
)

func fixPath(path string) (string, error) {
	resolvedPath := path
	var err error

	if strings.HasPrefix(path, "~/") {
		var home string
		home, err = os.UserHomeDir()
		if err == nil {
			resolvedPath = filepath.Join(home, path[2:])
		}
	}

	return resolvedPath, err
}

func initLogFile(logDir string) (string, error) {
	onceInitLog.Do(func() {
		fixedLogDir, err := fixPath(logDir)
		if err != nil {
			initLogErr = err
			return
		}

		timestamp := time.Now().UTC().Format("20060102T150405Z")
		filename := fmt.Sprintf("%s-gotelemetry.log", timestamp)
		currentLogFile = filepath.Join(fixedLogDir, filename)
	})

	if initLogErr != nil {
		return "", initLogErr
	}

	return currentLogFile, nil
}

func ensureLogFileOpen(fullPath string) (*os.File, error) {
	if logFile != nil {
		return logFile, nil
	}

	// crear el path si no existe
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("creating log directory %q: %w", dir, err)
	}

	f, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	logFile = f
	return logFile, nil
}

func LogExport(payload []byte, logDir string) error {
	fullPath, err := initLogFile(logDir)
	if err != nil {
		return err
	}

	logMu.Lock()
	defer logMu.Unlock()

	f, err := ensureLogFileOpen(fullPath)
	if err != nil {
		return err
	}

	_, err = f.Write(append(payload, '\n'))
	return err
}

// CloseLogFile permite cerrar el archivo limpiamente cuando se apaga el sistema.
func CloseLogFile() error {
	logMu.Lock()
	defer logMu.Unlock()

	if logFile == nil {
		return nil
	}

	err := logFile.Close()
	logFile = nil
	return err
}
