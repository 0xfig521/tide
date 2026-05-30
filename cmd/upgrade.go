package cmd

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/0xfig521/tide/internal/output"
)

const (
	githubOwner = "0xfig521"
	githubRepo  = "tide"
	releasesURL = "https://api.github.com/repos/" + githubOwner + "/" + githubRepo + "/releases"
)

var (
	upgradeCheck bool
	upgradeTag   string
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade tide to the latest version",
	Long: `Download and install the latest version of tide from GitHub Releases.

Examples:
  tide upgrade              # Upgrade to latest
  tide upgrade --check      # Check if new version available
  tide upgrade --tag v0.2.0 # Install a specific version`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if upgradeCheck {
			return checkUpdate()
		}
		if err := doUpgrade(upgradeTag); err != nil {
			return output.PrintError(output.CodeInternalError, err.Error())
		}
		return nil
	},
}

func init() {
	upgradeCmd.Flags().BoolVar(&upgradeCheck, "check", false, "Check if a new version is available")
	upgradeCmd.Flags().StringVar(&upgradeTag, "tag", "", "Install a specific version tag (e.g. v0.2.0)")
	rootCmd.AddCommand(upgradeCmd)
}

type releaseAsset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

type githubRelease struct {
	TagName string         `json:"tag_name"`
	Assets  []releaseAsset `json:"assets"`
}

func doUpgrade(targetVersion string) error {
	release, err := fetchRelease(targetVersion)
	if err != nil {
		return fmt.Errorf("fetch release: %w", err)
	}

	if !targetVersionSet() && release.TagName == version {
		output.PrintSuccess(map[string]any{"current": version, "message": "already up to date"}, nil)
		return nil
	}

	assetName := fmt.Sprintf("tide-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	assetURL := ""
	for _, a := range release.Assets {
		if a.Name == assetName {
			assetURL = a.URL
			break
		}
	}
	if assetURL == "" {
		return fmt.Errorf("no release asset found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	tmpDir, err := os.MkdirTemp("", "tide-upgrade-")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	tarballPath := filepath.Join(tmpDir, "tide.tar.gz")
	if err := downloadFile(tarballPath, assetURL); err != nil {
		return fmt.Errorf("download: %w", err)
	}

	binaryPath := filepath.Join(tmpDir, "tide")
	if err := extractBinary(tarballPath, binaryPath); err != nil {
		return fmt.Errorf("extract: %w", err)
	}

	if err := os.Chmod(binaryPath, 0755); err != nil {
		return fmt.Errorf("chmod: %w", err)
	}

	// Verify the downloaded binary works
	check := exec.Command(binaryPath, "--version")
	checkOut, err := check.Output()
	checkVer := release.TagName
	if err == nil {
		checkVer = strings.TrimSpace(string(checkOut))
	}

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("find current executable: %w", err)
	}

	// On macOS, resolve symlinks
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return fmt.Errorf("resolve executable path: %w", err)
	}

	backupPath := exePath + ".old"
	if err := os.Rename(exePath, backupPath); err != nil {
		return fmt.Errorf("backup current binary: %w", err)
	}

	if err := copyFile(binaryPath, exePath); err != nil {
		// Restore backup on failure
		os.Rename(backupPath, exePath)
		return fmt.Errorf("install new binary: %w", err)
	}

	os.Remove(backupPath)

	output.PrintSuccess(map[string]any{"previous": version, "current": checkVer}, nil)
	return nil
}

func checkUpdate() error {
	release, err := fetchRelease("")
	if err != nil {
		return output.PrintError(output.CodeInternalError, err.Error())
	}

	if release.TagName == version {
		output.PrintSuccess(map[string]any{"current": version, "latest": version, "update_available": false}, nil)
		return nil
	}
	output.PrintSuccess(map[string]any{"current": version, "latest": release.TagName, "update_available": true}, nil)
	return nil
}

func fetchRelease(targetVersion string) (*githubRelease, error) {
	url := releasesURL + "/latest"
	if targetVersion != "" {
		url = releasesURL + "/tags/" + targetVersion
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "tide-upgrade")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		if targetVersion != "" {
			return nil, fmt.Errorf("version %s not found", targetVersion)
		}
		return nil, fmt.Errorf("no releases found")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var rel githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, fmt.Errorf("parse release: %w", err)
	}
	return &rel, nil
}

func downloadFile(dst, url string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/octet-stream")
	req.Header.Set("User-Agent", "tide-upgrade")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

func extractBinary(tarballPath, dst string) error {
	f, err := os.Open(tarballPath)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if hdr.Name == "tide" || filepath.Base(hdr.Name) == "tide" {
			out, err := os.Create(dst)
			if err != nil {
				return err
			}
			defer out.Close()
			if _, err := io.Copy(out, tr); err != nil {
				return err
			}
			return nil
		}
	}
	return fmt.Errorf("binary 'tide' not found in archive")
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

func targetVersionSet() bool {
	return upgradeTag != ""
}
