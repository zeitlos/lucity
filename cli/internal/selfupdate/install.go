package selfupdate

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

const (
	maxChecksums = 64 << 10
	maxSignature = 1 << 10
	maxArchive   = 200 << 20
	maxBinary    = 400 << 20
)

var publicKey = "x6sDeaO1mw08QGJYz9Xmy9+ZMJ4nsNTeL6WQ3Ew3L6Q="

// Update installs the latest release when it is newer than current. It returns
// the latest release and whether it was installed.
func Update(ctx context.Context, current string) (string, bool, error) {
	if _, ok := parse(current); !ok {
		return "", false, fmt.Errorf("lucity %s is a development build: install a release to update it", current)
	}
	latest, err := Refresh(ctx)
	if err != nil {
		return "", false, fmt.Errorf("find the latest release: %w", err)
	}
	if !Newer(current, latest) {
		return latest, false, nil
	}
	err = install(ctx, latest)
	previous := load()
	previous.FailedTo = ""
	if err != nil {
		previous.FailedTo = latest
	}
	save(previous)
	if err != nil {
		return latest, false, err
	}
	return latest, true, nil
}

// Lock keeps a second updater from running at the same time. It returns false
// when another one holds the lock.
func Lock() (func(), bool, error) {
	directory, err := cacheDirectory()
	if err != nil {
		return nil, false, err
	}
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return nil, false, err
	}
	file, err := os.OpenFile(filepath.Join(directory, "update.lock"), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, false, err
	}
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		file.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return func() { file.Close() }, true, nil
}

// StartBackground runs 'lucity update --background' detached, so it finishes
// after the command that started it has exited.
func StartBackground() error {
	executable, err := executablePath()
	if err != nil {
		return err
	}
	command := exec.Command(executable, "update", "--background")
	command.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := command.Start(); err != nil {
		return err
	}
	return command.Process.Release()
}

func install(ctx context.Context, version string) error {
	key, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil || len(key) != ed25519.PublicKeySize {
		return errors.New("this build has no release signing key, so it can't verify updates")
	}
	executable, err := executablePath()
	if err != nil {
		return err
	}
	directory := filepath.Dir(executable)
	if !Writable() {
		return fmt.Errorf("can't replace %s: run 'sudo lucity update'", executable)
	}

	base := downloadURL + "/v" + version + "/"
	checksums, err := download(ctx, base+"checksums.txt", maxChecksums)
	if err != nil {
		return err
	}
	signature, err := download(ctx, base+"checksums.txt.sig", maxSignature)
	if err != nil {
		return err
	}
	if !ed25519.Verify(key, checksums, signature) {
		return fmt.Errorf("the signature of release %s doesn't match", version)
	}

	asset := fmt.Sprintf("lucity_%s_%s_%s.tar.gz", version, runtime.GOOS, runtime.GOARCH)
	want, err := checksum(checksums, asset)
	if err != nil {
		return err
	}
	archive, err := download(ctx, base+asset, maxArchive)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(archive)
	if hex.EncodeToString(sum[:]) != want {
		return fmt.Errorf("%s doesn't match its checksum", asset)
	}
	binary, err := extract(archive)
	if err != nil {
		return err
	}

	temporary, err := os.CreateTemp(directory, ".lucity-update-")
	if err != nil {
		return err
	}
	defer os.Remove(temporary.Name())
	_, writeErr := temporary.Write(binary)
	closeErr := temporary.Close()
	if err := errors.Join(writeErr, closeErr); err != nil {
		return err
	}
	if err := os.Chmod(temporary.Name(), 0o755); err != nil {
		return err
	}
	if err := checkRuns(ctx, temporary.Name(), version); err != nil {
		return err
	}
	return os.Rename(temporary.Name(), executable)
}

func download(ctx context.Context, url string, limit int64) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "lucity-cli")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download %s: %s", url, response.Status)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("download %s: larger than expected", url)
	}
	return data, nil
}

func checksum(checksums []byte, asset string) (string, error) {
	for line := range strings.Lines(string(checksums)) {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[1] == asset {
			return fields[0], nil
		}
	}
	return "", fmt.Errorf("the release has no checksum for %s", asset)
}

func extract(archive []byte) ([]byte, error) {
	compressed, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, err
	}
	reader := tar.NewReader(compressed)
	for {
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			return nil, errors.New("the release archive has no lucity binary")
		}
		if err != nil {
			return nil, err
		}
		if header.Name == "lucity" && header.Typeflag == tar.TypeReg {
			return io.ReadAll(io.LimitReader(reader, maxBinary))
		}
	}
}

func checkRuns(ctx context.Context, binary, version string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary, "version")
	command.Env = append(os.Environ(), "LUCITY_NO_UPDATE_CHECK=1")
	output, err := command.Output()
	if err != nil {
		return fmt.Errorf("the downloaded binary doesn't run: %w", err)
	}
	if strings.TrimSpace(string(output)) != "lucity "+version {
		return fmt.Errorf("the downloaded binary reports %q instead of lucity %s", strings.TrimSpace(string(output)), version)
	}
	return nil
}

func executablePath() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(executable)
}
