package utils

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

func Untar(dst string, r io.Reader) error {
	tr := tar.NewReader(r)
	madeDir := map[string]bool{}
	for {
		header, err := tr.Next()
		switch {
		// if no more files are found return
		case err == io.EOF:
			return nil

		// return any other error
		case err != nil:
			return err

		// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue
		}

		// the target location where the dir/file should be created
		target := filepath.Join(dst, header.Name)
		// the following switch could also be done using fi.Mode(), not sure if there
		// a benefit of using one vs. the other.
		// fi := header.FileInfo()

		// check the file type
		switch header.Typeflag {
		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if err := makeDir(target, madeDir); err != nil {
				return err
			}

		// if it's a file create it
		case tar.TypeReg:
			if err := makeDir(filepath.Dir(target), madeDir); err != nil {
				return err
			}

			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				fmt.Println("could not open file")
				return err
			}

			// copy over contents
			_, err = copyBuffered(f, tr)
			if err1 := f.Close(); err == nil {
				err = err1
			}
			if err != nil {
				return err
			}
		}
	}
}

func makeDir(target string, made map[string]bool) error {
	if made[target] {
		return nil
	}

	if _, err := os.Stat(target); err != nil {
		if err := os.MkdirAll(target, 0755); err != nil {
			return err
		}
	}

	made[target] = true
	return nil
}

var bufPool = &sync.Pool{
	New: func() interface{} {
		buffer := make([]byte, 64*1024)
		return &buffer
	},
}

func copyBuffered(dst io.Writer, src io.Reader) (written int64, err error) {
	buf := bufPool.Get().(*[]byte)
	defer bufPool.Put(buf)

	for {
		nr, er := src.Read(*buf)
		if nr > 0 {
			nw, ew := dst.Write((*buf)[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return written, err
}
