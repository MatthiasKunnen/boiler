package filecasing_test

import (
	"github.com/MatthiasKunnen/boiler/pkg/filecasing"
	"github.com/stretchr/testify/assert"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCaseRoundtrip(t *testing.T) {
	testPath := filepath.Join("testdata", t.Name())
	dirs := []string{
		"tEst/hello/There",
		"tEst2",
		"tEsu/hey/hey",
	}
	files := []string{
		"Foo",
		"tEst/hello/There/bR",
		"tEst/hello/There/hmm.txt",
		"tEsu/hey/FOO",
	}
	defer func() {
		err := os.RemoveAll(testPath)
		assert.NoError(t, err)
	}()
	for _, dir := range dirs {
		assert.NoError(t, os.MkdirAll(filepath.Join(testPath, dir), 0755))
	}
	for _, file := range files {
		f, err := os.Create(filepath.Join(testPath, file))
		assert.NoError(t, err)
		assert.NoError(t, f.Close())
	}

	var originals []string
	err := filecasing.MakeLowerCase(testPath, func(original string) {
		originals = append(originals, original)
	})
	assert.NoError(t, err)

	for _, dir := range dirs {
		assert.NoDirExists(t, filepath.Join(testPath, dir))
		assert.DirExists(t, filepath.Join(testPath, strings.ToLower(dir)))
	}
	for _, file := range files {
		assert.NoFileExists(t, filepath.Join(testPath, file))
		assert.FileExists(t, filepath.Join(testPath, strings.ToLower(file)))
	}

	for _, original := range originals {
		assert.NoError(t, filecasing.RestoreCase(testPath, original))
	}

	var found []string
	err = filepath.WalkDir(testPath, func(path string, d fs.DirEntry, err error) error {
		rel, err := filepath.Rel(testPath, path)
		if err != nil {
			return err
		}
		found = append(found, rel)
		return nil
	})
	assert.NoError(t, err)

	for _, dir := range dirs {
		assert.DirExists(t, filepath.Join(testPath, dir))
	}
	for _, file := range files {
		assert.FileExists(t, filepath.Join(testPath, file))
	}
}

func TestMakeLowerCase(t *testing.T) {
	testPath := filepath.Join("testdata", t.Name())
	dirs := []string{
		"tEst/hello/There",
		"tEst2",
		"tEsu/hey/hey",
	}
	files := []string{
		"Foo",
		"tEst/hello/There/bR",
		"tEst/hello/There/hmm.txt",
		"tEsu/hey/FOO",
	}
	expectedOriginals := []string{
		"Foo",
		"tEst/hello/There/bR",
		"tEst/hello/There",
		"tEst",
		"tEst2",
		"tEsu/hey/FOO",
		"tEsu",
	}
	defer func() {
		err := os.RemoveAll(testPath)
		assert.NoError(t, err)
	}()
	for _, dir := range dirs {
		assert.NoError(t, os.MkdirAll(filepath.Join(testPath, dir), 0755))
	}
	for _, file := range files {
		f, err := os.Create(filepath.Join(testPath, file))
		assert.NoError(t, err)
		assert.NoError(t, f.Close())
	}

	var actual []string
	err := filecasing.MakeLowerCase(testPath, func(original string) {
		actual = append(actual, original)
	})
	assert.NoError(t, err)
	assert.Equal(t, expectedOriginals, actual)

	for _, dir := range dirs {
		assert.NoDirExists(t, filepath.Join(testPath, dir))
		assert.DirExists(t, filepath.Join(testPath, strings.ToLower(dir)))
	}
	for _, file := range files {
		assert.NoFileExists(t, filepath.Join(testPath, file))
		assert.FileExists(t, filepath.Join(testPath, strings.ToLower(file)))
	}
}

func TestWalkDfs(t *testing.T) {
	var actual []string
	expected := []string{
		"a.txt",
		"tEst/HuH/eZe.txt",
		"tEst/HuH",
		"tEst/test.txt",
		"tEst",
		"tEst2/test.txt",
		"tEst2/tesu.txt",
		"tEst2",
	}
	err := filecasing.WalkDfs("testdata/walktest1", func(path string, d fs.DirEntry) error {
		actual = append(actual, path)
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}
