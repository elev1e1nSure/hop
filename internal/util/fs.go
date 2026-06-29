package util

import (
	"fmt"
	"os"
	"path/filepath"

	"hop/internal/apperr"
)

func AtomicWrite(path string, data []byte, mode, dirMode os.FileMode) error {
	realPath := path
	if info, err := os.Lstat(path); err == nil && info.Mode()&os.ModeSymlink != 0 {
		resolved, resolveErr := filepath.EvalSymlinks(path)
		if resolveErr != nil {
			cause := fmt.Errorf("resolve symlink %q: %w", path, resolveErr)
			return apperr.Wrap(apperr.ErrReplaceFile, cause, path)
		}
		realPath = resolved
	}
	dir := filepath.Dir(realPath)
	if err := os.MkdirAll(dir, dirMode); err != nil {
		cause := fmt.Errorf("create parent directory %q: %w", dir, err)
		return apperr.Wrap(apperr.ErrCreateDirectory, cause, dir)
	}

	tmp, err := os.CreateTemp(dir, ".hop-*")
	if err != nil {
		cause := fmt.Errorf("create temporary file in %q: %w", dir, err)
		return apperr.Wrap(apperr.ErrCreateTemporaryFile, cause, dir)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()

	if err := tmp.Chmod(mode); err != nil {
		_ = tmp.Close()
		cause := fmt.Errorf("set mode %o on temporary file %q: %w", mode.Perm(), tmpName, err)
		return apperr.Wrap(apperr.ErrSetFileMode, cause, path)
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		cause := fmt.Errorf("write temporary file %q for %q: %w", tmpName, path, err)
		return apperr.Wrap(apperr.ErrWriteTemporaryFile, cause, path)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		cause := fmt.Errorf("sync temporary file %q for %q: %w", tmpName, path, err)
		return apperr.Wrap(apperr.ErrSyncTemporaryFile, cause, path)
	}
	if err := tmp.Close(); err != nil {
		cause := fmt.Errorf("close temporary file %q for %q: %w", tmpName, path, err)
		return apperr.Wrap(apperr.ErrCloseTemporaryFile, cause, path)
	}
	if err := os.Rename(tmpName, realPath); err != nil {
		cause := fmt.Errorf("replace %q with %q: %w", realPath, tmpName, err)
		return apperr.Wrap(apperr.ErrReplaceFile, cause, realPath)
	}
	return nil
}
