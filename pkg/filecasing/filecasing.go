package filecasing

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// MakeLowerCase makes all directories and files in the given root directory lower case and
// reports the original path relative to root when a change has been made.
// To restore the case, call RestoreCase with all the delivered paths in reserve.
func MakeLowerCase(root string, deliver func(original string)) error {
	return WalkDfs(root, func(path string, d os.DirEntry) error {
		basename := filepath.Base(path)
		if basename == strings.ToLower(basename) {
			return nil
		}

		newName := filepath.Join(root, filepath.Dir(path), strings.ToLower(basename))
		err := os.Rename(
			filepath.Join(root, path),
			newName,
		)
		if err != nil {
			return err
		}

		deliver(path)
		return nil
	})
}

// RestoreCase restores the case of the file at the base of the path.
// Example: given path /Foo/Bar/Baz, /foo/bar/baz will become /Foo/Bar/baz.
func RestoreCase(root string, path string) error {
	dir := filepath.Join(root, strings.ToLower(filepath.Dir(path)))
	basename := filepath.Base(path)
	return os.Rename(
		filepath.Join(dir, strings.ToLower(basename)),
		filepath.Join(dir, basename),
	)
}

type WalkDirFunc func(path string, d fs.DirEntry) error

// WalkDfs walks a directory in depth-first order.
// It will call deliver for every directory entry found under the given root directory.
// The reported paths are relative to the root.
// The entries of a directory are delivered in lexical order. However, directories are
// traversed before being delivered.
// Returning an error in deliver stops traversal and makes WalkDfs return the error.
func WalkDfs(root string, deliver WalkDirFunc) error {
	info, err := os.Lstat(root)
	if err != nil {
		return err
	}

	return walkDfs(root, "", fs.FileInfoToDirEntry(info), deliver, true)
}

func walkDfs(root, path string, d fs.DirEntry, deliver WalkDirFunc, isFirst bool) error {
	if !d.IsDir() {
		return deliver(path, d)
	}

	entries, err := os.ReadDir(filepath.Join(root, path))
	if err != nil {
		return err
	}

	for _, d1 := range entries {
		err = walkDfs(root, filepath.Join(path, d1.Name()), d1, deliver, false)
		if err != nil {
			return err
		}
	}

	if isFirst {
		return nil
	}

	err = deliver(path, d)
	if err != nil {
		return err
	}

	return nil
}
