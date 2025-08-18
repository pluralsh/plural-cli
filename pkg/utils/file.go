package utils

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/pluralsh/plural-cli/pkg/utils/pathing"
	"sigs.k8s.io/yaml"
)

func ListDirectory(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

func IsDir(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}

	return fileInfo.IsDir()
}

func CopyFile(src, dest string) error {
	bytesRead, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dest, bytesRead, 0644)
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
	if errors.Is(err, io.EOF) {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}

func WriteFile(name string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
		return err
	}
	return os.WriteFile(name, content, 0644)
}

func ReadFile(name string) (string, error) {
	content, err := os.ReadFile(name)
	return string(content), err
}

func ReadRemoteFile(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	buffer := new(bytes.Buffer)
	if _, err = buffer.ReadFrom(resp.Body); err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func ReadRemoteFileWithRetries(url, token string, retries int) (io.ReadCloser, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Token "+token)

	for i := 0; i < retries; i++ {
		resp, retriable, err := doRequest(req)
		if err != nil {
			if !retriable {
				return nil, err
			}

			time.Sleep(time.Duration(50*(i+1)) * time.Millisecond)
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("could read file, retries exhaused: %w", err)
}

func doRequest(req *http.Request) (io.ReadCloser, bool, error) {
	client := &http.Client{Timeout: time.Minute, Transport: &http.Transport{ResponseHeaderTimeout: time.Minute}}

	resp, err := client.Do(req)
	if err != nil {
		return nil, false, err
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		errMsg, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, false, fmt.Errorf("could not read response body: %w", err)
		}

		return nil, resp.StatusCode == http.StatusTooManyRequests,
			fmt.Errorf("could not fetch url: %s, error: %s, code: %d", req.URL.String(), string(errMsg), resp.StatusCode)
	}

	return resp.Body, false, nil
}

func YamlFile(name string, out interface{}) error {
	content, err := os.ReadFile(name)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(content, out)
}

func RemoteYamlFile(url string, out interface{}) error {
	content, err := ReadRemoteFile(url)
	if err != nil {
		return err
	}

	return yaml.Unmarshal([]byte(content), out)
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func CompareFileContent(filename, content string) (bool, error) {
	c, err := ReadFile(filename)
	if err != nil {
		return false, err
	}
	return c == content, nil
}

func DownloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func CopyDir(src string, dst string) error {
	var err error
	var srcinfo os.FileInfo

	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}

	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	fds := make([]fs.FileInfo, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return err
		}
		fds = append(fds, info)
	}

	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = CopyDir(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		} else {
			if err = CopyFile(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}

func EnsureDir(dir string) error {
	if dir == "" {
		return fmt.Errorf("directory name cannot be empty")
	}

	if !Exists(dir) {
		return os.MkdirAll(filepath.Dir(dir), 0755)
	}

	if !IsDir(dir) {
		return fmt.Errorf("%s is not a directory", dir)
	}

	return nil
}
