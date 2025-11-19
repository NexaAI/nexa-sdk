package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/bytedance/sonic"
	goversion "github.com/hashicorp/go-version"
	"github.com/spf13/cobra"

	"github.com/NexaAI/nexa-sdk/runner/internal/downloader"
	"github.com/NexaAI/nexa-sdk/runner/internal/render"
	"github.com/NexaAI/nexa-sdk/runner/internal/store"
)

const (
	githubAPIURL = "https://api.github.com/repos/NexaAI/nexa-sdk/releases/latest"
	userAgent    = "Nexa-Updater/1.0"

	updateCheckInterval  = 24 * time.Hour
	notificationInterval = 8 * time.Hour
)

func update() *cobra.Command {
	updateCmd := &cobra.Command{
		GroupID: "management",
		Use:     "update",
		Short:   "update nexa",
		Long:    "Update nexa to the latest version",
	}

	updateCmd.Run = func(cmd *cobra.Command, args []string) {
		err := func() error {
			// check platform
			assetMap := map[string]map[string]string{
				"darwin": {
					"amd64": "nexa-cli_macos_x86_64.pkg",
					"arm64": "nexa-cli_macos_arm64.pkg",
				},
				"windows": {
					"amd64": "nexa-cli_windows_x86_64.exe",
					"arm64": "nexa-cli_windows_arm64.exe",
				},
			}
			assetName, ok := assetMap[runtime.GOOS][runtime.GOARCH]
			if !ok {
				return fmt.Errorf("current platform is not supported, please manually download")
			}

			rls, err := getLastRelease()
			if err != nil {
				return err
			}

			updateAvailable, err := hasUpdate(Version, rls.TagName)
			if err != nil {
				return err
			}

			if !updateAvailable {
				fmt.Println("Already up-to-date.")
				return nil
			}

			// find asset
			var ast asset
			for _, asset := range rls.Assets {
				if asset.Name == assetName {
					ast = asset
					break
				}
			}
			if ast.Name == "" {
				return fmt.Errorf("asset %s not found in release", assetName)
			}

			fmt.Println(
				render.GetTheme().Warning.Sprint("New version found, file: "),
				render.GetTheme().Success.Sprint(ast.Name),
				render.GetTheme().Warning.Sprint(", version: "),
				render.GetTheme().Success.Sprint(rls.TagName))

			// download
			dst := filepath.Join(os.TempDir(), ast.Name)
			progress := make(chan int64)
			bar := render.NewProgressBar(int64(ast.Size), "downloading")
			go func() {
				defer bar.Exit()
				for pg := range progress {
					bar.Add(pg)
				}
			}()
			if err = downloadPkg(ast.BrowserDownloadURL, dst, int64(ast.Size), progress); err != nil {
				return err
			}

			if err = installPkg(dst); err != nil {
				return err
			}
			fmt.Println("update package is ready to install")

			return nil
		}()
		if err != nil {
			fmt.Println(render.GetTheme().Error.Sprintf("Update failed: %s", err.Error()))
			os.Exit(1)
		}
	}

	return updateCmd
}

// util functions

func getLastRelease() (release, error) {
	var rls release

	client := &http.Client{Timeout: 10 * time.Second}

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
	return rls, err
}

func hasUpdate(cur, latest string) (bool, error) {
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

// downloadPkg a file from url to dst with progress
func downloadPkg(url, dst string, size int64, progress chan int64) error {
	defer close(progress)

	file, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	downloader := downloader.NewDownloader("")

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

	return nil
}

func installPkg(pkgPath string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", pkgPath)
	case "windows":
		cmd = exec.Command(pkgPath)
	default:
		return errors.New("update is not supported on this platform")
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Start()
}

// Update check store

type updateCheck struct {
	LastCheck     time.Time `json:"last_check"`
	LastNotify    time.Time `json:"last_notify"`
	LatestVersion string    `json:"latest_version"`
}

func getUpdateCheck() (updateCheck, error) {
	ck := updateCheck{}

	data, err := os.ReadFile(filepath.Join(store.Get().DataPath(), "update_check"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = setUpdateCheck(ck)
		}
		return ck, err
	}

	err = sonic.Unmarshal(data, &ck)
	return ck, err
}

func setUpdateCheck(ck updateCheck) error {
	slog.Info("setting last update check", "update_check", ck)

	data, _ := sonic.Marshal(ck)
	return os.WriteFile(filepath.Join(store.Get().DataPath(), "update_check"), data, 0644)
}

// check and notify

func checkUpdate() {
	slog.Info("checking for updates")

	ck, err := getUpdateCheck()
	slog.Info("last update check", "update_check", ck, "error", err)
	if err != nil {
		return
	}

	if time.Since(ck.LastCheck) < updateCheckInterval {
		slog.Info("skip update check, interval not reached", "last_check_time", ck.LastCheck)
		return
	}

	rls, err := getLastRelease()
	slog.Debug("latest release", "release", rls, "error", err)
	if err != nil {
		return
	}

	ck.LastCheck = time.Now()
	ck.LatestVersion = rls.TagName
	setUpdateCheck(ck)
}

func notifyUpdate() {
	ck, _ := getUpdateCheck()
	slog.Info("notifying update", "update_check", ck)

	upAvail, _ := hasUpdate(Version, ck.LatestVersion)
	if !upAvail || time.Since(ck.LastNotify) < notificationInterval {
		slog.Info("skip update notification", "update_available", upAvail, "time_since_last_notify", time.Since(ck.LastNotify))
		return
	}

	ck.LastNotify = time.Now()
	setUpdateCheck(ck)

	fmt.Fprintf(os.Stderr, "\n\n%s %s → %s\n",
		render.GetTheme().Warning.Sprintf("A new version of nexa-cli is available:"),
		render.GetTheme().Success.Sprint(Version),
		render.GetTheme().Success.Sprint(ck.LatestVersion))

	fmt.Fprintf(os.Stderr, "%s\n\n",
		render.GetTheme().Warning.Sprint("To update, run: `nexa update`"),
	)
}
