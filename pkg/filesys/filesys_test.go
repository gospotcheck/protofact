package filesys

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_CreateAndDelete_TmpDir(t *testing.T) {
	fs := &FS{}
	path, err := fs.CreateUniqueTmpDir("/tmp")
	if err != nil {
		t.Error(err)
	}
	err = fs.DeleteDir(path)
	if err != nil {
		t.Error(err)
	}
}

func Test_GetSubDirectories_ReturnsTopLevelDirectories(t *testing.T) {
	path, _ := filepath.Abs("./test-resources")

	fs := &FS{}

	dirs, _ := fs.GetSubDirectories(path)

	assert.Contains(t, dirs, "subdir-a")
	assert.Contains(t, dirs, "subdir-b")
	assert.Len(t, dirs, 2)
}

func Test_GetSubDirectories_DoesNotReturnFiles(t *testing.T) {
	path, _ := filepath.Abs("./test-resources")

	fs := &FS{}

	dirs, _ := fs.GetSubDirectories(path)

	assert.NotContains(t, dirs, "itsafile.txt")
}

func Test_GetFileNames_ReturnsFileNamesOnly(t *testing.T) {
	path, _ := filepath.Abs("./test-resources")

	fs := &FS{}

	fileNames, _ := fs.GetFileNames(path)

	assert.Contains(t, fileNames, "itsafile.txt")
	assert.Len(t, fileNames, 1)
}
