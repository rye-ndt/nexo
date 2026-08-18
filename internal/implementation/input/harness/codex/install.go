package codex

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"hexago/internal/helpers/custom_error"
	"hexago/internal/implementation/input/harness/harness_helper"
)

type githubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Digest             string `json:"digest"`
	} `json:"assets"`
}

func (c *codex) release() (*harness_helper.Release, error) {
	platform, archiveExt, err := codexPlatform()
	if err != nil {
		return nil, err
	}

	latest := &githubRelease{}
	if err := c.httpCli.GetJSON(c.cfg.ReleaseBase+"/releases/latest", latest); err != nil {
		return nil, custom_error.Critical("resolve latest release: %v", err)
	}

	want := c.cfg.BinName + "-" + platform + archiveExt
	for _, asset := range latest.Assets {
		if asset.Name != want {
			continue
		}

		version := strings.TrimPrefix(strings.TrimPrefix(latest.TagName, "rust-v"), "v")
		return &harness_helper.Release{
			Version:  version,
			Platform: platform,
			URL:      asset.BrowserDownloadURL,
			Checksum: strings.TrimPrefix(asset.Digest, "sha256:"),
		}, nil
	}

	return nil, custom_error.Critical("no codex build for platform %s", platform)
}

func codexPlatform() (string, string, error) {
	arch := map[string]string{"arm64": "aarch64", "amd64": "x86_64"}[runtime.GOARCH]
	if arch == "" {
		return "", "", custom_error.Critical("unsupported platform %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	switch runtime.GOOS {
	case "darwin":
		return arch + "-apple-darwin", ".tar.gz", nil
	case "linux":
		return arch + "-unknown-linux-musl", ".tar.gz", nil
	case "windows":
		return arch + "-pc-windows-msvc", ".exe.zip", nil
	default:
		return "", "", custom_error.Critical("unsupported platform %s/%s", runtime.GOOS, runtime.GOARCH)
	}
}

func extractBinary(archive, dest string) error {
	if strings.HasSuffix(strings.ToLower(archiveName(archive)), ".zip") || runtime.GOOS == "windows" {
		return extractZip(archive, dest)
	}

	return extractTarGz(archive, dest)
}

func archiveName(path string) string {
	if strings.HasSuffix(path, ".download") {
		if runtime.GOOS == "windows" {
			return "codex.zip"
		}
		return "codex.tar.gz"
	}

	return path
}

func extractTarGz(archive, dest string) error {
	file, err := os.Open(archive)
	if err != nil {
		return custom_error.Critical("open archive: %v", err)
	}
	defer file.Close()

	gz, err := gzip.NewReader(file)
	if err != nil {
		return custom_error.Critical("open gzip stream: %v", err)
	}
	defer gz.Close()

	reader := tar.NewReader(gz)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return custom_error.Critical("read archive: %v", err)
		}
		if header.Typeflag == tar.TypeReg && strings.HasPrefix(filepath.Base(header.Name), "codex") {
			return writeBinary(reader, dest)
		}
	}

	return custom_error.Critical("codex binary not found in archive")
}

func extractZip(archive, dest string) error {
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return custom_error.Critical("open archive: %v", err)
	}
	defer reader.Close()

	for _, file := range reader.File {
		name := strings.ToLower(filepath.Base(file.Name))
		if !strings.HasPrefix(name, "codex") || !strings.HasSuffix(name, ".exe") {
			continue
		}

		source, err := file.Open()
		if err != nil {
			return custom_error.Critical("read %s from archive: %v", file.Name, err)
		}
		err = writeBinary(source, dest)
		source.Close()
		return err
	}

	return custom_error.Critical("codex binary not found in archive")
}

func writeBinary(source io.Reader, dest string) error {
	tmp := dest + ".tmp"
	file, err := os.Create(tmp)
	if err != nil {
		return custom_error.Critical("create %s: %v", tmp, err)
	}

	_, copyErr := io.Copy(file, source)
	closeErr := file.Close()
	if copyErr != nil {
		err = copyErr
	} else if closeErr != nil {
		err = closeErr
	} else {
		err = os.Rename(tmp, dest)
	}

	if err != nil {
		os.Remove(tmp)
		return custom_error.Critical("extract codex binary: %v", err)
	}

	return nil
}
