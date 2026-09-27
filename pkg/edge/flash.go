package edge

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// ProgressFunc reports bytes written while flashing.
type ProgressFunc func(written, total int64)

// FlashOptions are the CLI flags for plural edge flash.
type FlashOptions struct {
	Image      string
	Device     string
	Progress   io.Writer
	Log        io.Writer
	OnProgress ProgressFunc
}

var (
	openFlashDevice = func(path string) (*os.File, error) {
		return os.OpenFile(path, os.O_WRONLY, 0644)
	}
	elevateFlash = flashWithSudo
)

// DeviceWritable reports whether the current user can open the device for writing.
func DeviceWritable(path string) bool {
	file, err := openFlashDevice(path)
	if err != nil {
		return false
	}
	_ = file.Close()
	return true
}

func isPermissionErr(err error) bool {
	return errors.Is(err, os.ErrPermission) || errors.Is(err, syscall.EACCES)
}

func imageSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}

// Flash writes an image file onto a storage device.
func Flash(options FlashOptions) error {
	image := options.Image
	device := options.Device
	if image == "" {
		return fmt.Errorf("image file path is required")
	}
	if device == "" {
		return fmt.Errorf("storage device path is required")
	}

	out, err := openFlashDevice(device)
	if err != nil {
		if isPermissionErr(err) {
			return elevateFlash(options, err)
		}
		return fmt.Errorf("could not open device: %w", err)
	}
	defer out.Close()

	in, err := os.Open(image)
	if err != nil {
		return fmt.Errorf("could not open image: %w", err)
	}
	defer in.Close()

	total := imageSize(image)
	counter := &countWriter{total: total, out: options.Progress, fn: options.OnProgress}
	_, err = io.Copy(io.MultiWriter(out, counter), in)
	if err == nil && options.OnProgress != nil {
		options.OnProgress(counter.written, total)
	}
	return err
}

type countWriter struct {
	total, written, lastReported int64
	lastTime                     time.Time
	out                          io.Writer
	fn                           ProgressFunc
}

func (w *countWriter) Write(p []byte) (int, error) {
	n := len(p)
	if w.out != nil {
		var err error
		n, err = w.out.Write(p)
		if err != nil {
			return n, err
		}
	}
	w.written += int64(n)
	if w.fn == nil {
		return n, nil
	}
	now := time.Now()
	if w.written-w.lastReported >= 1<<20 || now.Sub(w.lastTime) >= 100*time.Millisecond {
		w.lastReported = w.written
		w.lastTime = now
		w.fn(w.written, w.total)
	}
	return n, nil
}

func flashLog(options FlashOptions, line string) {
	w := options.Log
	if w == nil {
		return
	}
	_, _ = io.WriteString(w, line+"\n")
}

func flashWithSudo(options FlashOptions, original error) error {
	flashLog(options, "device is not writable; retrying with sudo/pkexec")
	parser := newDDProgress(imageSize(options.Image), options)
	if err := runDD("sudo", []string{"-n", "dd"}, options.Image, options.Device, parser); err == nil {
		parser.Close()
		return nil
	}
	if err := runDD("pkexec", []string{"dd"}, options.Image, options.Device, parser); err == nil {
		parser.Close()
		return nil
	}
	parser.Close()
	return fmt.Errorf("could not open device: %w\nneed root to write %s — retry with: sudo plural tui", original, options.Device)
}

func runDD(name string, prefix []string, image, device string, output io.Writer) error {
	args := append(append([]string{}, prefix...), "if="+image, "of="+device, "bs=4M", "conv=fsync", "status=progress")
	cmd := exec.Command(name, args...)
	cmd.Stdout = output
	cmd.Stderr = output
	return cmd.Run()
}

type ddProgress struct {
	total int64
	buf   strings.Builder
	fn    ProgressFunc
	log   io.Writer
	raw   io.Writer
}

func newDDProgress(total int64, options FlashOptions) *ddProgress {
	raw := io.Writer(nil)
	if options.Log == nil {
		raw = os.Stderr
	}
	return &ddProgress{total: total, fn: options.OnProgress, log: options.Log, raw: raw}
}

func (w *ddProgress) Write(p []byte) (int, error) {
	for _, b := range p {
		if b == '\r' || b == '\n' {
			w.flush()
			continue
		}
		w.buf.WriteByte(b)
	}
	return len(p), nil
}

func (w *ddProgress) Close() {
	w.flush()
}

func (w *ddProgress) flush() {
	line := strings.TrimSpace(w.buf.String())
	w.buf.Reset()
	if line == "" {
		return
	}
	if written, ok := parseDDProgress(line); ok {
		if w.fn != nil {
			w.fn(written, w.total)
		}
		if w.raw != nil {
			_, _ = io.WriteString(w.raw, "\r"+line)
		}
		return
	}
	if w.log != nil {
		_, _ = io.WriteString(w.log, line+"\n")
	} else if w.raw != nil {
		_, _ = io.WriteString(w.raw, line+"\n")
	}
}

func parseDDProgress(line string) (int64, bool) {
	line = strings.TrimSpace(line)
	fields := strings.Fields(line)
	if len(fields) < 3 || fields[1] != "bytes" || !strings.Contains(line, "copied") {
		return 0, false
	}
	n, err := strconv.ParseInt(strings.ReplaceAll(fields[0], ",", ""), 10, 64)
	if err != nil || n < 0 {
		return 0, false
	}
	return n, true
}

// DefaultFlashImage returns the absolute path of image/kairos.img in dir
// (or the working directory) when that file exists.
func DefaultFlashImage(dir string) string {
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return ""
		}
	}
	candidates := []string{
		filepath.Join(dir, "image", "kairos.img"),
		filepath.Join(dir, "kairos.img"),
	}
	for _, path := range candidates {
		info, err := os.Stat(path)
		if err != nil || !info.Mode().IsRegular() {
			continue
		}
		abs, err := filepath.Abs(path)
		if err != nil {
			return path
		}
		return abs
	}
	return ""
}
