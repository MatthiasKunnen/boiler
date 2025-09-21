package boiler

import (
	"errors"
	"os"
	"path/filepath"
)

// overwriteSymlink creates a symlink and overwrites an existing file if it exists.
// Atomicity depends on the system, see [os.Rename].
func overwriteSymlink(target string, linkName string) error {
	var tempPath string
	for i := 0; i < 5; i++ {
		tempPath = filepath.Join(os.TempDir(), tempPath)

		err := os.Symlink(
			target,
			tempPath,
		)
		switch {
		case errors.Is(err, os.ErrExist):
			continue
		case err != nil:
			return err
		}
		break
	}

	return os.Rename(tempPath, linkName)
}
