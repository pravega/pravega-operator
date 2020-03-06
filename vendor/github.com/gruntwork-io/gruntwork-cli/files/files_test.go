package files

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
)

const TEST_FILE_EXPECTED_CONTENTS = "Hello, World!"

func TestFileExistsValidPath(t *testing.T) {
	t.Parallel()

	assert.True(t, FileExists("../fixtures/files/test-file.txt"))
}

func TestFileExistsInvalidPath(t *testing.T) {
	t.Parallel()

	assert.False(t, FileExists("../fixtures/files/not-a-valid-file-path.txt"))
}

func TestIsDirOnDir(t *testing.T) {
	t.Parallel()

	assert.True(t, IsDir("../fixtures/files"))
}

func TestIsDirOnFile(t *testing.T) {
	t.Parallel()

	assert.False(t, IsDir("../fixtures/files/test-file.txt"))
}

func TestIsDirOnInvalidPath(t *testing.T) {
	t.Parallel()

	assert.False(t, IsDir("../fixtures/files/not-a-valid-file-path.txt"))
}

func TestReadFileAsString(t *testing.T) {
	t.Parallel()

	contents, err := ReadFileAsString("../fixtures/files/test-file.txt")
	assert.Nil(t, err, "Unexpected error: %v", err)
	assert.Equal(t, TEST_FILE_EXPECTED_CONTENTS, contents)
}

func TestCopyFile(t *testing.T) {
	t.Parallel()

	destPath, err := ioutil.TempFile("", "test-copy-file")
	if err != nil {
		t.Fatal(err)
	}

	err = CopyFile("../fixtures/files/test-file.txt", destPath.Name())
	if err != nil {
		t.Fatal(err)
	}

	contents, err := ReadFileAsString(destPath.Name())
	assert.Nil(t, err, "Unexpected error: %v", err)
	assert.Equal(t, TEST_FILE_EXPECTED_CONTENTS, contents)
}
