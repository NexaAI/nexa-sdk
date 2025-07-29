package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/NexaAI/nexa-sdk/internal/render"
	"github.com/NexaAI/nexa-sdk/internal/store"
	"github.com/NexaAI/nexa-sdk/internal/types"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
)

const (
	GithubAPIURL = "https://api.github.com/repos/zhiyuan8/homebrew-go-release/releases/latest"
	UserAgent    = "Nexa-Updater/1.0"
)

type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int    `json:"size"`
	Digest             string `json:"digest"`
}

type Platform struct {
	OS      string
	Arch    string
	Backend string
	Version string
}

func update() *cobra.Command {
	updateCmd := &cobra.Command{}
	updateCmd.Use = "update"
	updateCmd.Short = "update nexa"
	updateCmd.Long = "Update nexa to the latest version"

	var loop bool
	updateCmd.Flags().BoolVar(&loop, "loop", false, "check update in loop")

	updateCmd.Run = func(cmd *cobra.Command, args []string) {
		if err := updateImpl(); err != nil {
			fmt.Println(text.FgRed.Sprintf("update failed: %s", err.Error()))
			os.Exit(1)
		}
	}

	return updateCmd
}

func updateImpl() error {
	platform, err := detectPlatform()
	if err != nil {
		return err
	}
	fmt.Printf("platform: %+v\n", platform)

	release, err := getLatestRelease()
	if err != nil {
		return err
	}

	// check if need update
	if release.TagName <= platform.Version {
		fmt.Printf("No newer version available: %s\n", release.TagName)
		return nil
	}

	ast, err := findMatchingAsset(release, platform)
	if err != nil {
		return err
	}
	fmt.Printf("New version found, file: %s, version: %s\n", ast.Name, release.TagName)

	// download update package
	dst := filepath.Join(os.TempDir(), "nexa", release.TagName, ast.Name)
	// fmt.Printf("download update package to: %s\n", dst)

	progress := make(chan types.DownloadInfo)
	bar := render.NewProgressBar(int64(ast.Size), "downloading")
	fmt.Println()
	go func() {
		defer bar.Exit()
		for pg := range progress {
			bar.Set(pg.TotalDownloaded)
		}
	}()

	if err = download(ast.BrowserDownloadURL, dst, progress); err != nil {
		return err
	}

	if err = installUpdate(dst); err != nil {
		return err
	}
	fmt.Println("update package is ready to install")
	return nil
}

func getCurrentVersion() (string, string, error) {
	var version, backend string

	// nexa version:
	// NexaSDK Bridge Version: v0.1.2-rc4_llama-cpp-metal
	// NexaSDK CLI Version:    v0.2.16-rc2
	cmd := exec.Command("/usr/local/bin/nexa", "version")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println(err.Error())
		return "", "", err
	}

	versionStr := strings.TrimSpace(string(output))

	lines := strings.SplitSeq(versionStr, "\n")
	for line := range lines {
		if strings.Contains(line, "CLI Version") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				version = strings.TrimSpace(parts[1])
			}
		}

		if strings.Contains(line, "Bridge Version") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				backend = strings.Split(strings.TrimSpace(parts[1]), "_")[1]
			}
		}
	}

	if version == "" || backend == "" {
		return "", "", fmt.Errorf("can not parse version: %s", versionStr)
	}
	return version, backend, nil
}

// 检测当前平台
func detectPlatform() (Platform, error) {
	version, backend, err := getCurrentVersion()
	if err != nil {
		return Platform{}, err
	}

	platform := Platform{
		OS:      runtime.GOOS,
		Arch:    runtime.GOARCH,
		Version: version,
		Backend: backend,
	}

	return platform, nil
}

func getLatestRelease() (Release, error) {
	var release Release

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", GithubAPIURL, nil)
	if err != nil {
		return release, err
	}

	req.Header.Set("User-Agent", UserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return release, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return release, fmt.Errorf("get latest release failed: %d", resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(&release)
	if err != nil {
		return release, err
	}

	return release, nil
}

func findMatchingAsset(release Release, platform Platform) (*Asset, error) {
	var assetName string
	switch platform.OS {
	case "darwin":
		macOSVersion, err := getMacOSVersion()
		if err != nil {
			return nil, err
		}

		assetName = fmt.Sprintf("nexa-cli_macos-%s.pkg", macOSVersion)
		for _, asset := range release.Assets {
			if asset.Name == assetName {
				return &asset, nil
			}
		}

	case "windows":
		assetName = "nexa-cli_windows-setup.exe"
		for _, asset := range release.Assets {
			if asset.Name == assetName {
				return &asset, nil
			}
		}

		// 	curl -fsSL https://raw.githubusercontent.com/NexaAI/nexa-sdk/main/release/linux/install.sh -o install.sh && chmod +x install.sh && ./install.sh
		// case "linux":
	}
	return nil, errors.New("can not find matching asset")
}

// download a file from url to dst with progress
func download(url, dst string, progress chan types.DownloadInfo) error {
	if progress != nil {
		defer close(progress)
	}

	manifestPath := filepath.Join(filepath.Dir(dst), "manifest.json")
	manifest, err := os.Open(manifestPath)
	if err == nil {
		defer manifest.Close()
		var ast Asset
		if err = json.NewDecoder(manifest).Decode(&ast); err != nil {
			return err
		}

		if progress != nil {
			progress <- types.DownloadInfo{
				CurrentName:     ast.Name,
				CurrentSize:     int64(ast.Size),
				TotalDownloaded: int64(ast.Size),
				TotalSize:       int64(ast.Size),
			}
		}

		fmt.Printf("file already exists: %s, size: %d\n", ast.Name, ast.Size)
		return nil
	}

	downloader := store.NewHFDownloader(0, progress)
	if err = downloader.Download(context.Background(), url, dst); err != nil {
		return err
	}

	info, err := os.Stat(dst)
	if err != nil {
		return err
	}

	// 写入 manifest.json
	ast := Asset{
		Name:               filepath.Base(dst),
		BrowserDownloadURL: url,
		Size:               int(info.Size()),
	}

	manifest, err = os.Create(manifestPath)
	if err != nil {
		return err
	}
	defer manifest.Close()
	return json.NewEncoder(manifest).Encode(ast)
}

// X86 -> 13
// ARM >=15 -> 15
// ARM others: 14
func getMacOSVersion() (string, error) {
	if runtime.GOOS != "darwin" {
		return "", errors.New("not macos")
	}

	if runtime.GOARCH != "arm64" {
		return "13", nil
	}

	// 运行 sw_vers 命令获取 macOS 版本
	cmd := exec.Command("sw_vers", "-productVersion")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	versionStr := strings.TrimSpace(string(output))
	if versionStr >= "15" {
		return "15", nil
	}
	return "14", nil
}

func installUpdate(pkgPath string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", pkgPath)

	case "windows":
		cmd = exec.Command(pkgPath)

	case "linux":
		return errors.New("linux upgrade is not supported")
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Start()
}
