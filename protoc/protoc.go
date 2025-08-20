package protoc

import (
	"archive/zip"
	"bytes"
	"fmt"
	"go/build"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	protocUrl                    = "https://github.com/protocolbuffers/protobuf/releases/download"
	cacheDir                     = "protoc-cache"
	protocExecutableRelativePath = "bin/protoc"
)

func EnsureProtoc(protocVersion string) (string, error) {
	exPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("error getting executable path: %w", err)
	}

	cachePath := filepath.Join(filepath.Dir(exPath), cacheDir)
	protocExecutablePath := filepath.Join(filepath.Dir(cachePath), cacheDir, protocExecutableRelativePath)
	if _, err := os.Stat(protocExecutablePath); err == nil {
		return protocExecutablePath, nil
	}

	var protocOsName string
	switch runtime.GOOS {
	case "linux":
		protocOsName = "linux"
	case "windows":
		protocOsName = "win"
	case "darwin":
		protocOsName = "osx"
	default:
		return "", fmt.Errorf("unsupported os")
	}

	var protocArchName string
	switch runtime.GOARCH {
	case "amd64":
		protocArchName = "x86_64"
	case "386":
		protocArchName = "x86_32"
	case "arm64":
		protocArchName = "aarch_64"
	default:
		return "", fmt.Errorf("unsupported architecture")
	}

	var protocInfo string
	if protocOsName == "win" {
		switch protocArchName {
		case "x86_64":
			protocInfo = "win64"
		case "x86_32":
			protocInfo = "win32"
		default:
			return "", fmt.Errorf("unsupported os and architecture combination")
		}
	} else {
		protocInfo = protocOsName + "-" + protocArchName
	}

	downloadUrl := fmt.Sprintf("%s/v%s/protoc-%s-%s.zip", protocUrl, protocVersion, protocVersion, protocInfo)

	resp, err := http.Get(downloadUrl)
	if err != nil {
		return "", fmt.Errorf("error downloading protoc: %w", err)
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("error downloading protoc: %d %s, %s", resp.StatusCode, resp.Status, downloadUrl)
	}

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading protoc response: %w", err)
	}

	zipReader, err := zip.NewReader(bytes.NewReader(respBytes), resp.ContentLength)
	if err != nil {
		return "", fmt.Errorf("error creating zip reader: %w", err)
	}

	for _, f := range zipReader.File {
		path := filepath.Join(cachePath, f.Name)
		if f.FileInfo().IsDir() {
			err := os.MkdirAll(path, 0o755)
			if err != nil {
				return "", fmt.Errorf("error creating cache directory: %w", err)
			}
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return "", fmt.Errorf("error opening zip file: %w", err)
		}

		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return "", err
		}

		out, err := os.Create(path)
		if err != nil {
			return "", fmt.Errorf("error creating file: %w", err)
		}

		if _, err := io.Copy(out, rc); err != nil {
			return "", fmt.Errorf("error copying file from zip: %w", err)
		}

		_ = out.Close()
		_ = rc.Close()
	}

	err = os.Chmod(protocExecutablePath, 0o755)
	if err != nil {
		return "", fmt.Errorf("error setting executable permissions: %w", err)
	}

	return protocExecutablePath, nil
}

func EnsureProtocPlugins() ([]string, error) {
	plugins := []string{
		"google.golang.org/protobuf/cmd/protoc-gen-go@latest",
		"google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest",
	}

	pluginsExecutables := make([]string, len(plugins))
	for i, plugin := range plugins {
		pluginsExecutables[i] = strings.Split(filepath.Base(plugin), "@")[0]
	}

	goBin := os.Getenv("GOBIN")
	if goBin == "" {
		goPath := os.Getenv("GOPATH")
		if goPath == "" {
			goPath = build.Default.GOPATH
		}
		goBin = filepath.Join(goPath, "bin")
	}

	for i, plugin := range plugins {
		if _, err := os.Stat(filepath.Join(goBin, pluginsExecutables[i])); err != nil {
			fmt.Println("Installing", plugin)
			cmd := exec.Command("go", "install", plugin)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Env = os.Environ()
			if err := cmd.Run(); err != nil {
				return nil, err
			}
		}
	}

	pluginsPath := make([]string, len(plugins))
	for i, plugin := range pluginsExecutables {
		pluginsPath[i] = filepath.Join(goBin, plugin)
	}

	return pluginsPath, nil
}
