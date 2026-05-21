package cli

import (
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func defaultWorkflowPath(prefix, ext string) string {
	if ext == "" {
		ext = ".bin"
	}
	return prefix + "-" + time.Now().UTC().Format("20060102-150405") + ext
}

func outputPathForMedia(outputPath, prefix, ext string) string {
	if strings.TrimSpace(outputPath) == "" {
		return defaultWorkflowPath(prefix, ext)
	}
	info, err := os.Stat(outputPath)
	if err == nil && info.IsDir() {
		return filepath.Join(outputPath, defaultWorkflowPath(prefix, ext))
	}
	if filepath.Ext(outputPath) == "" && ext != "" {
		return outputPath + ext
	}
	return outputPath
}

func writeWorkflowFile(path string, data []byte, force bool) (string, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	if !force {
		if _, err := os.Stat(path); err == nil {
			return "", fmt.Errorf("%s already exists; pass --force to overwrite", path)
		}
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return "", err
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return absolute, nil
}

func extensionForContentType(contentType, fallback string) string {
	if semi := strings.Index(contentType, ";"); semi >= 0 {
		contentType = contentType[:semi]
	}
	if ext, err := mime.ExtensionsByType(strings.TrimSpace(contentType)); err == nil && len(ext) > 0 {
		return ext[0]
	}
	return fallback
}
