package agents

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	SessionTarName = "agent-session.tar.gz"
	manifestName   = "manifest.json"
)

// ArchiveReader reads and extracts uploaded agent session archives.
//
// Implementations must reject unsafe archive entries because session archives
// are downloaded before being restored into user-controlled directories.
type ArchiveReader interface {
	// ReadManifest reads manifest.json from the archive.
	ReadManifest(path string) (*SessionManifest, error)
	// ExtractSubtree extracts the requested archive subtree into dst.
	ExtractSubtree(path, archivePath, dst string) error
	// Contains reports whether the requested archive subtree exists.
	Contains(path, archivePath string) (bool, error)
}

// TarGzipArchiveReader reads the tar.gz session archive format uploaded by
// agent runtimes.
type TarGzipArchiveReader struct{}

func (in TarGzipArchiveReader) ReadManifest(path string) (*SessionManifest, error) {
	err := in.walkArchive(path, func(header *tar.Header, reader io.Reader) error {
		if filepath.ToSlash(header.Name) != manifestName {
			return nil
		}
		if header.Typeflag != tar.TypeReg {
			return fmt.Errorf("manifest entry is not a regular file")
		}
		var manifest SessionManifest
		if err := json.NewDecoder(reader).Decode(&manifest); err != nil {
			return fmt.Errorf("decode session manifest: %w", err)
		}
		return foundManifest{manifest: &manifest}
	})
	if found, ok := errors.AsType[foundManifest](err); ok {
		return found.manifest, nil
	}
	if err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("session archive does not contain %s", manifestName)
}

func (in TarGzipArchiveReader) Contains(path, archivePath string) (bool, error) {
	archivePath = in.cleanArchivePath(archivePath)
	if archivePath == "" {
		return false, nil
	}
	err := in.walkArchive(path, func(header *tar.Header, _ io.Reader) error {
		if in.entryInSubtree(header.Name, archivePath) {
			return errFoundEntry
		}
		return nil
	})
	if errors.Is(err, errFoundEntry) {
		return true, nil
	}
	return false, err
}

func (in TarGzipArchiveReader) ExtractSubtree(path, archivePath, dst string) error {
	archivePath = in.cleanArchivePath(archivePath)
	if archivePath == "" {
		return fmt.Errorf("archive path is required")
	}
	if err := os.MkdirAll(dst, 0755); err != nil {
		return fmt.Errorf("create restore directory: %w", err)
	}

	var extracted bool
	err := in.walkArchive(path, func(header *tar.Header, reader io.Reader) error {
		if !in.entryInSubtree(header.Name, archivePath) {
			return nil
		}
		rel, ok := in.subtreeRel(header.Name, archivePath)
		if !ok {
			return nil
		}
		extracted = true
		return in.extractEntry(dst, rel, header, reader)
	})
	if err != nil {
		return err
	}
	if !extracted {
		return fmt.Errorf("session archive does not contain %q", archivePath)
	}
	return nil
}

func (in TarGzipArchiveReader) walkArchive(path string, visit func(*tar.Header, io.Reader) error) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open session archive: %w", err)
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("open gzip session archive: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return fmt.Errorf("read session archive: %w", err)
		case header == nil:
			continue
		}
		if err := visit(header, tr); err != nil {
			return err
		}
	}
}

func (in TarGzipArchiveReader) extractEntry(dst, rel string, header *tar.Header, reader io.Reader) error {
	if rel == "" || rel == "." {
		if header.Typeflag == tar.TypeDir {
			return nil
		}
		return nil
	}
	target, err := in.safeTarget(dst, rel)
	if err != nil {
		return err
	}
	mode := os.FileMode(header.Mode)
	switch header.Typeflag {
	case tar.TypeDir:
		return os.MkdirAll(target, mode.Perm())
	case tar.TypeReg:
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		file, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode.Perm())
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(file, reader)
		closeErr := file.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	case tar.TypeSymlink, tar.TypeLink:
		return fmt.Errorf("session archive contains unsupported link entry %q", header.Name)
	default:
		return fmt.Errorf("session archive contains unsupported entry %q", header.Name)
	}
}

func (in TarGzipArchiveReader) safeTarget(dst, rel string) (string, error) {
	rel = filepath.Clean(filepath.FromSlash(rel))
	if filepath.IsAbs(rel) || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("unsafe archive path %q", rel)
	}
	target := filepath.Join(dst, rel)
	cleanDst, err := filepath.Abs(dst)
	if err != nil {
		return "", err
	}
	cleanTarget, err := filepath.Abs(target)
	if err != nil {
		return "", err
	}
	if cleanTarget != cleanDst && !strings.HasPrefix(cleanTarget, cleanDst+string(filepath.Separator)) {
		return "", fmt.Errorf("archive path escapes restore directory: %q", rel)
	}
	return target, nil
}

func (in TarGzipArchiveReader) cleanArchivePath(path string) string {
	path = filepath.ToSlash(filepath.Clean(strings.TrimSpace(path)))
	path = strings.TrimPrefix(path, "/")
	if path == "." {
		return ""
	}
	return path
}

func (in TarGzipArchiveReader) entryInSubtree(name, archivePath string) bool {
	name = in.cleanArchivePath(name)
	return name == archivePath || strings.HasPrefix(name, archivePath+"/")
}

func (in TarGzipArchiveReader) subtreeRel(name, archivePath string) (string, bool) {
	name = in.cleanArchivePath(name)
	if name == archivePath {
		return ".", true
	}
	return strings.TrimPrefix(name, archivePath+"/"), strings.HasPrefix(name, archivePath+"/")
}

type foundManifest struct {
	manifest *SessionManifest
}

func (f foundManifest) Error() string { return "found manifest" }

var errFoundEntry = fmt.Errorf("found archive entry")
