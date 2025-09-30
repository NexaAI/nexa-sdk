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

	"github.com/NexaAI/nexa-sdk/runner/internal/downloader"
	"github.com/NexaAI/nexa-sdk/runner/internal/render"
	"github.com/bytedance/sonic"
	goversion "github.com/hashicorp/go-version"
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
	updateCmd := &cobra.Command{
		GroupID: "management",
		Use:     "update",
		Short:   "update nexa",
		Long:    "Update nexa to the latest version",
	}

	updateCmd.Run = func(cmd *cobra.Command, args []string) {
		if err := updateImpl(); err != nil {
			fmt.Println(render.GetTheme().Error.Sprintf("update failed: %s", err.Error()))
			os.Exit(1)
		}
	}

	return updateCmd
}

func updateImpl() error {
	ck, rls, err := checkForUpdate(true)
	if err != nil {
		return err
	}

	if !ck.UpdateAvailable {
		fmt.Println("Already up-to-date.")
		return nil
	}

	ast, err := findMatchingAsset(rls)
	if err != nil {
		return err
	}
	fmt.Println(
		render.GetTheme().Warning.Sprint("New version found, file: "),
		render.GetTheme().Success.Sprint(ast.Name),
		render.GetTheme().Warning.Sprint(", version: "),
		render.GetTheme().Success.Sprint(rls.TagName))

	dst := filepath.Join(os.TempDir(), "nexa", rls.TagName, ast.Name)
	progress := make(chan int64)
	bar := render.NewProgressBar(int64(ast.Size), "downloading")
	go func() {
		defer bar.Exit()
		for pg := range progress {
			bar.Add(pg)
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

func needUpdate(cur, latest string) (bool, error) {
	curVer, err := goversion.NewVersion(cur)
	if err != nil {
		return false, fmt.Errorf("invalid SemVer %s: %w", cur, err)
	}
	latestVer, err := goversion.NewVersion(latest)
	if err != nil {
		return false, fmt.Errorf("invalid SemVer %s: %w", latest, err)
	}
	return curVer.Compare(latestVer) < 0, nil
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
		case "arm64":
			assetName = "nexa-cli_windows_arm64.exe"
		}

		// 	curl -fsSL https://raw.githubusercontent.com/NexaAI/nexa-sdk/main/release/linux/install.sh -o install.sh && chmod +x install.sh && ./install.sh
		// case "linux":
	}

	for _, asset := range rls.Assets {
		if asset.Name == assetName {
			return &asset, nil
		}
	}

	return nil, fmt.Errorf("no matching asset found: %s/%s", rls.TagName, assetName)
}

// download a file from url to dst with progress
func download(url, dst string, progress chan int64) error {
	if progress != nil {
		defer close(progress)
	}

	// Ensure destination directory exists before accessing manifest or file
	dir := filepath.Dir(dst)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	manifestPath := filepath.Join(dir, "manifest.json")
	manifest, err := os.Open(manifestPath)
	if err == nil {
		defer manifest.Close()
		var ast asset
		if err = sonic.ConfigDefault.NewDecoder(manifest).Decode(&ast); err != nil {
			return err
		}

		if progress != nil {
			progress <- int64(ast.Size)
		}
		return nil
	}

	file, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	downloader := downloader.NewDownloader("")
	size, err := downloader.GetFileSize(url)
	if err != nil {
		return err
	}
	for offset := int64(0); offset < size; offset += 1024 * 1024 {
		limit := int64(1024 * 1024)
		if offset+limit > size {
			limit = size - offset
		}
		if err = downloader.DownloadChunk(context.Background(), url, offset, limit, file); err != nil {
			return err
		}
		progress <- int64(limit)
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
	CheckTime       time.Time `json:"check_time"`
	LastNotify      time.Time `json:"last_notify"`
	LatestVersion   string    `json:"latest_version"`
	UpdateAvailable bool      `json:"update_available"`
}

func getLastCheck() (updateCheck, error) {
	var ck updateCheck
	configDir, err := os.UserConfigDir()
	if err != nil {
		return ck, err
	}
	checkFile := filepath.Join(configDir, "nexa-cli", ".updatecheck")

	data, err := os.ReadFile(checkFile)
	if errors.Is(err, os.ErrNotExist) {
		err = setLastCheck(&ck)
	}
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

func checkForUpdate(manual bool) (updateCheck, release, error) {
	ck, err := getLastCheck()
	if err != nil {
		return ck, release{}, err
	}

	if !manual && time.Since(ck.CheckTime) < updateCheckInterval {
		return ck, release{}, nil
	}

	rls, err := getLatestRelease()
	if err != nil {
		return ck, rls, err
	}

	upAvail, err := needUpdate(Version, rls.TagName)
	if err != nil {
		return ck, rls, err
	}

	ck.CheckTime = time.Now()
	ck.LatestVersion = rls.TagName
	ck.UpdateAvailable = upAvail
	return ck, rls, setLastCheck(&ck)
}

func notifyUpdate() {
	ck, _ := getLastCheck()
	if !ck.UpdateAvailable || time.Since(ck.LastNotify) < notificationInterval {
		return
	}

	ck.LastNotify = time.Now()
	setLastCheck(&ck)

	fmt.Fprintf(os.Stderr, "\n\n%s %s → %s\n",
		render.GetTheme().Warning.Sprintf("A new version of nexa-cli is available:"),
		render.GetTheme().Success.Sprint(Version),
		render.GetTheme().Success.Sprint(ck.LatestVersion))

	fmt.Fprintf(os.Stderr, "%s\n\n",
		render.GetTheme().Warning.Sprint("To update, run: `nexa update`"),
	)
}
