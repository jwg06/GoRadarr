package system

import (
	"archive/zip"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

type backupInfo struct {
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	Created time.Time `json:"created"`
}

func (h *handler) listBackups(w http.ResponseWriter, r *http.Request) {
	backupDir := filepath.Join(h.cfg.Data.RootDir, "backups")
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			writeJSON(w, http.StatusOK, []backupInfo{})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "failed to read backup directory"})
		return
	}

	result := make([]backupInfo, 0)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".zip") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		result = append(result, backupInfo{
			Name:    entry.Name(),
			Size:    info.Size(),
			Created: info.ModTime().UTC(),
		})
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *handler) createBackup(w http.ResponseWriter, r *http.Request) {
	backupDir := filepath.Join(h.cfg.Data.RootDir, "backups")
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "failed to create backup directory"})
		return
	}

	name := "goradarr_backup_" + time.Now().UTC().Format("20060102_150405") + ".zip"
	zipPath := filepath.Join(backupDir, name)

	f, err := os.Create(zipPath)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "failed to create backup file"})
		return
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	dbPath := filepath.Join(h.cfg.Data.RootDir, "goradarr.db")
	if err := addFileToZip(zw, dbPath, "goradarr.db"); err != nil && !os.IsNotExist(err) {
		zw.Close()
		f.Close()
		os.Remove(zipPath)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "failed to add db to backup"})
		return
	}

	// config.yaml is optional
	cfgPath := filepath.Join(h.cfg.Data.RootDir, "config.yaml")
	_ = addFileToZip(zw, cfgPath, "config.yaml")

	// Flush zip before stat
	if err := zw.Close(); err != nil {
		f.Close()
		os.Remove(zipPath)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "failed to finalise backup"})
		return
	}
	f.Close()

	var size int64
	if info, err := os.Stat(zipPath); err == nil {
		size = info.Size()
	}

	writeJSON(w, http.StatusCreated, backupInfo{
		Name:    name,
		Size:    size,
		Created: time.Now().UTC(),
	})
}

func (h *handler) deleteBackup(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if name == "" || strings.Contains(name, "/") || strings.Contains(name, "..") {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid backup name"})
		return
	}

	backupPath := filepath.Join(h.cfg.Data.RootDir, "backups", name)
	if err := os.Remove(backupPath); err != nil {
		if os.IsNotExist(err) {
			writeJSON(w, http.StatusNotFound, map[string]string{"message": "backup not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "failed to delete backup"})
		return
	}
	w.WriteHeader(http.StatusOK)
}

func addFileToZip(zw *zip.Writer, filePath, archiveName string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	w, err := zw.Create(archiveName)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, f)
	return err
}
