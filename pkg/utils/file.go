package utils

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pluralsh/plural/pkg/utils/pathing"
)

func CopyFile(src, dest string) error {
	bytesRead, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(dest, bytesRead, 0644)
}

func EmptyDirectory(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()

	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}

	for _, name := range names {
		if err := os.RemoveAll(pathing.SanitizeFilepath(filepath.Join(dir, name))); err != nil {
			return err
		}
	}

	return nil
}

func IsEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}

func WriteFile(name string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
		return err
	}
	return ioutil.WriteFile(name, content, 0644)
}

func ReadFile(name string) (string, error) {
	content, err := ioutil.ReadFile(name)
	return string(content), err
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
