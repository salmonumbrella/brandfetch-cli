package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func computeSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file for checksum: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("failed to read file for checksum: %w", err)
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func verifySHA256(path, expected string) (bool, error) {
	sum, err := computeSHA256(path)
	if err != nil {
		return false, err
	}

	return strings.EqualFold(sum, strings.TrimSpace(expected)), nil
}

type checksumEntry struct {
	Path string
	Sum  string
}

func parseSHA256Manifest(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	entries := make(map[string]string)
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		parts := strings.Fields(trimmed)
		if len(parts) < 2 {
			continue
		}
		hash := parts[0]
		filename := parts[1]
		filename = strings.TrimPrefix(filename, "*")
		filename = strings.TrimPrefix(filename, "./")
		if filename != "" {
			entries[filename] = hash
		}
	}

	return entries, nil
}

func verifySHA256ManifestEntry(path, root string, manifest map[string]string) error {
	if manifest == nil {
		return nil
	}
	keys := []string{filepath.Base(path)}
	if root != "" {
		if rel, err := filepath.Rel(root, path); err == nil && rel != "" && rel != "." {
			keys = append([]string{rel}, keys...)
		}
	}
	var expected string
	var ok bool
	for _, key := range keys {
		if expected, ok = manifest[key]; ok {
			break
		}
	}
	if !ok {
		return fmt.Errorf("no manifest entry for %s", filepath.Base(path))
	}
	okSum, err := verifySHA256(path, expected)
	if err != nil {
		return err
	}
	if !okSum {
		return fmt.Errorf("expected %s", expected)
	}
	return nil
}

func buildChecksumEntry(path, root string) (checksumEntry, error) {
	sum, err := computeSHA256(path)
	if err != nil {
		return checksumEntry{}, err
	}
	entryPath := filepath.Base(path)
	if root != "" {
		if rel, err := filepath.Rel(root, path); err == nil && rel != "" && rel != "." {
			entryPath = rel
		}
	}
	return checksumEntry{Path: entryPath, Sum: sum}, nil
}

func writeSHA256Manifest(path string, entries []checksumEntry, appendExisting bool) error {
	if len(entries) == 0 {
		return fmt.Errorf("no downloaded files to write")
	}

	merged := make(map[string]string)
	if appendExisting {
		if existing, err := parseSHA256Manifest(path); err == nil {
			for k, v := range existing {
				merged[k] = v
			}
		}
	}

	for _, entry := range entries {
		if entry.Path == "" || entry.Sum == "" {
			continue
		}
		merged[entry.Path] = entry.Sum
	}

	paths := make([]string, 0, len(merged))
	for path := range merged {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	var sb strings.Builder
	for _, name := range paths {
		sb.WriteString(fmt.Sprintf("%s  %s\n", merged[name], name))
	}

	return os.WriteFile(path, []byte(sb.String()), 0o644)
}
