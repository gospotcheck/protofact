// Package filesys provides a struct, FS, that receives several
// file system related methods.
package filesys

import (
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

// FS is a struct owning several file system related methods.
// This package provides a struct rather than a bunch of public methods
// so that it can be passed in dependency injection.
type FS struct{}

// CreateUniqueTmpDir generates a UUID and creates a directory under
// the passed parent path using the UUID as the directory name.
func (f *FS) CreateUniqueTmpDir(parentPath string) (string, error) {
	u := uuid.NewV4()

	path := fmt.Sprintf("%s/%s", parentPath, u)

	err := os.Mkdir(path, 0750)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("could not create directory at /tmp/%s", u))
	}
	return path, nil
}

// DeleteDir performs the equivalent of a rm -rf
// on the passed path.
func (f *FS) DeleteDir(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("could not delete directory at path %s", path))
	}
	return nil
}

// GetSubDirectories returns a list of directory names in a given directory.
// Does not include file names.
func (f *FS) GetSubDirectories(baseDir string) ([]string, error) {
	// get directory names
	file, err := os.Open(baseDir)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to open directory: %s", baseDir))
	}
	defer file.Close()

	list, err := file.Readdir(0)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to read directory: %s", baseDir))
	}

	// make a slice that includes only the directories, not files
	directories := []string{}
	for _, fileInfo := range list {
		if fileInfo.IsDir() {
			directories = append(directories, fileInfo.Name())
		}
	}

	return directories, nil
}

// GetFileNames returns a slice of filenames in a directory.
// Does not include directory names.
func (f *FS) GetFileNames(baseDir string) ([]string, error) {
	file, err := os.Open(baseDir)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to open directory: %s", baseDir))
	}
	defer file.Close()

	list, err := file.Readdir(0)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to read directory: %s", baseDir))
	}

	// make a slice that includes only the directories, not files
	files := []string{}
	for _, fileInfo := range list {
		if !fileInfo.IsDir() {
			files = append(files, fileInfo.Name())
		}
	}

	return files, nil
}

// CopyFile copies a file from one path to another. Existing files
// are overwritten. Does not copy file attributes.
func (f *FS) CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return out.Close()
}
