package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/NexaAI/nexa-sdk/runner/internal/render"
	"github.com/NexaAI/nexa-sdk/runner/internal/store"
	"github.com/NexaAI/nexa-sdk/runner/internal/types"
	"github.com/bytedance/sonic"
	"github.com/spf13/cobra"
)

const (
	githubAPIURL = "https://api.github.com/repos/NexaAI/nexa-sdk/releases/latest"
	userAgent    = "Nexa-Updater/1.0"

	updateCheckInterval  = 24 * time.Hour
	notificationInterval = 8 * time.Hour
)

type release struct {
	TagName string  `json:"tag_name"`
	Assets  []asset `json:"assets"`
}

type asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int    `json:"size"`
	Digest             string `json:"digest"`
}

func update() *cobra.Command {
	updateCmd := &cobra.Command{}
	updateCmd.Use = "update"
	updateCmd.Short = "update nexa"
	updateCmd.Long = "Update nexa to the latest version"

	updateCmd.Run = func(cmd *cobra.Command, args []string) {
		if err := updateImpl(); err != nil {
			fmt.Println(render.GetTheme().Error.Sprintf("update failed: %s", err.Error()))
			os.Exit(1)
		}
	}

	return updateCmd
}

func updateImpl() error {
	rls, err := getLatestRelease()
	if err != nil {
		return err
	}

	if rls.TagName <= Version {
		fmt.Println("Already up-to-date.")
		return nil
	}

	ast, err := findMatchingAsset(rls)
	if err != nil {
		return err
	}
	fmt.Printf("New version found, file: %s, version: %s\n", ast.Name, rls.TagName)

	dst := filepath.Join(os.TempDir(), "nexa", rls.TagName, ast.Name)
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

	ck := updateCheck{
		CheckTime:     time.Now(),
		LastNotify:    time.Now(),
		LatestVersion: rls.TagName,
	}
	if err = setLastCheck(&ck); err != nil {
		return fmt.Errorf("failed to set last check: %w", err)
	}

	if err = installUpdate(dst); err != nil {
		return err
	}
	fmt.Println("update package is ready to install")
	return nil
}

func getLatestRelease() (release, error) {
	var rls release

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", githubAPIURL, nil)
	if err != nil {
		return rls, err
	}

	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return rls, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return rls, fmt.Errorf("get latest release failed: %d", resp.StatusCode)
	}

	err = sonic.ConfigDefault.NewDecoder(resp.Body).Decode(&rls)
	if err != nil {
		return rls, err
	}

	return rls, nil
}

func findMatchingAsset(rls release) (*asset, error) {
	var assetName string
	switch runtime.GOOS {
	case "darwin":
		switch runtime.GOARCH {
		case "amd64":
			assetName = "nexa-cli_macos_x86_64.pkg"
		case "arm64":
			assetName = "nexa-cli_macos_arm64.pkg"
		}

	case "windows":
		switch runtime.GOARCH {
		case "amd64":
			assetName = "nexa-cli_windows_x86_64.exe"
		}

		// 	curl -fsSL https://raw.githubusercontent.com/NexaAI/nexa-sdk/main/release/linux/install.sh -o install.sh && chmod +x install.sh && ./install.sh
		// case "linux":
	}

	for _, asset := range rls.Assets {
		if asset.Name == assetName {
			return &asset, nil
		}
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
		var ast asset
		if err = sonic.ConfigDefault.NewDecoder(manifest).Decode(&ast); err != nil {
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

	ast := asset{
		Name:               filepath.Base(dst),
		BrowserDownloadURL: url,
		Size:               int(info.Size()),
	}

	manifest, err = os.Create(manifestPath)
	if err != nil {
		return err
	}
	defer manifest.Close()

	return sonic.ConfigFastest.NewEncoder(manifest).Encode(ast)
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

type updateCheck struct {
	CheckTime     time.Time `json:"check_time"`
	LastNotify    time.Time `json:"last_notify"`
	LatestVersion string    `json:"latest_version"`
}

func getLastCheck() (updateCheck, error) {
	var ck updateCheck
	configDir, err := os.UserConfigDir()
	if err != nil {
		return ck, err
	}
	checkFile := filepath.Join(configDir, "nexa-cli", ".updatecheck")

	data, err := os.ReadFile(checkFile)
	if err != nil {
		return ck, err
	}

	if err = sonic.Unmarshal(data, &ck); err != nil {
		return ck, err
	}
	return ck, nil
}

func setLastCheck(ck *updateCheck) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	nexaConfigDir := filepath.Join(configDir, "nexa-cli")
	if err := os.MkdirAll(nexaConfigDir, 0755); err != nil {
		return err
	}

	data, err := sonic.Marshal(ck)
	if err != nil {
		return err
	}
	checkFile := filepath.Join(nexaConfigDir, ".updatecheck")
	return os.WriteFile(checkFile, data, 0644)
}

func notifyUpdate() {
	ck, _ := getLastCheck()
	if time.Since(ck.CheckTime) < updateCheckInterval {
		return
	}

	rls, err := getLatestRelease()
	if err != nil {
		fmt.Println("Failed to check for updates:", err)
		return
	}
	ck.CheckTime = time.Now()
	defer setLastCheck(&ck)

	if rls.TagName > Version {
		if rls.TagName != ck.LatestVersion || time.Since(ck.LastNotify) > notificationInterval {
			ck.LatestVersion = rls.TagName
			ck.LastNotify = time.Now()

			fmt.Fprintf(os.Stderr, "\n\n%s %s → %s\n",
				render.GetTheme().Warning.Sprintf("A new version of nexa-cli is available:"),
				render.GetTheme().Success.Sprint(Version),
				render.GetTheme().Success.Sprint(rls.TagName))

			fmt.Fprintf(os.Stderr, "%s\n\n",
				render.GetTheme().Warning.Sprint("To update, run: `nexa update`"),
			)
		}
	}
}
