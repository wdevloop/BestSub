package checker

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// ResourceFile defines the structure for managing a downloadable resource.
type ResourceFile struct {
	Name       string // Filename, e.g., "province.json"
	RemotePath string // Relative path on the remote server, e.g., "xykt/NetQuality/main/ref/province.json"
	LocalPath  string // Absolute local path, determined at runtime
}

var requiredResourceFiles = []ResourceFile{
	{
		Name:       "province.json",
		RemotePath: "xykt/NetQuality/main/ref/province.json",
	},
	{
		Name:       "iso3166.json",
		RemotePath: "xykt/NetQuality/main/ref/iso3166.json",
	},
	{
		Name:       "provider.json",
		RemotePath: "xykt/NetQuality/main/ref/provider.json",
	},
	{
		Name:       "target.json",
		RemotePath: "xykt/NetQuality/main/ref/target.json",
	},
	{
		Name:       "useragent.txt",
		RemotePath: "xykt/NetQuality/main/ref/useragent.txt",
	},
	{
		Name:       "cookies.txt",
		RemotePath: "xykt/NetQuality/main/ref/cookies.txt",
	},
	{
		Name:       "iata-icao.csv",
		RemotePath: "xykt/NetQuality/main/ref/iata-icao.csv",
	},
	{
		Name:       "AS_Mapping.txt",
		RemotePath: "xykt/NetQuality/main/ref/AS_Mapping.txt",
	},
	{
		Name:       "iperf.json",
		RemotePath: "xykt/NetQuality/main/ref/iperf.json",
	},
	{
		Name:       "speedtest_cn.json",
		RemotePath: "xykt/NetQuality/main/ref/speedtest_cn.json",
	},
}

// Base URL for downloading mirrored resources.
const mirrorBaseURL = "https://ghfast.top/https://raw.githubusercontent.com/"

// appSubDir is the subdirectory within the system's temp directory for storing app-specific data.
const appSubDir = "BestSub"
const refDataSubDir = "ref_data"

var appTempDir string // Cached application temporary directory

// getAppTempDir initializes and returns the application-specific temporary directory path.
// It creates the directory if it doesn't exist.
func getAppTempDir() (string, error) {
	if appTempDir != "" {
		return appTempDir, nil
	}

	baseTempDir := os.TempDir()
	path := filepath.Join(baseTempDir, appSubDir, refDataSubDir)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, 0755)
		if err != nil {
			return "", fmt.Errorf("failed to create temp directory %s: %w", path, err)
		}
		log.Printf("Created application temp directory: %s", path)
	} else if err != nil {
		return "", fmt.Errorf("failed to stat temp directory %s: %w", path, err)
	}
	appTempDir = path
	return appTempDir, nil
}

// downloadFile downloads a file from a URL and saves it to a local path.
func downloadFile(localFilepath string, url string) error {
	log.Printf("Downloading %s to %s", url, localFilepath)
	resp, err := http.Get(url) //nolint:gosec // User-provided mirror URL is a known source pattern
	if err != nil {
		return fmt.Errorf("failed to download %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download %s: status code %d", url, resp.StatusCode)
	}

	out, err := os.Create(localFilepath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", localFilepath, err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save content to %s: %w", localFilepath, err)
	}
	log.Printf("Successfully downloaded %s", localFilepath)
	return nil
}

// EnsureResourceFiles checks for the existence of required resource files in the app temp directory.
// If a file is missing, it attempts to download it from the mirror URL.
// It populates the LocalPath field for each ResourceFile.
func EnsureResourceFiles() error {
	tempDir, err := getAppTempDir()
	if err != nil {
		return fmt.Errorf("could not get app temp directory: %w", err)
	}

	for i := range requiredResourceFiles {
		res := &requiredResourceFiles[i]
		res.LocalPath = filepath.Join(tempDir, res.Name)
		fullURL := mirrorBaseURL + res.RemotePath

		if _, err := os.Stat(res.LocalPath); os.IsNotExist(err) {
			log.Printf("Resource file %s not found locally. Attempting download...", res.LocalPath)
			if err := downloadFile(res.LocalPath, fullURL); err != nil {
				// For now, log error and continue. Critical file handling can be added later.
				log.Printf("Error downloading %s from %s: %v. Proceeding without this file if possible.", res.Name, fullURL, err)
				// Depending on file criticality, you might want to return err here.
				// Example for critical files:
				// if res.Name == "province.json" || res.Name == "target.json" || res.Name == "provider.json" {
				// 	 return fmt.Errorf("failed to download critical resource file %s: %w", res.Name, err)
				// }
			}
		} else if err != nil {
			return fmt.Errorf("failed to check status of resource file %s: %w", res.LocalPath, err)
		} else {
			log.Printf("Resource file %s found locally.", res.LocalPath)
		}
	}
	return nil
}

// GetResourcePath returns the absolute local path for a given resource name.
// EnsureResourceFiles must be called successfully before this function.
func GetResourcePath(name string) (string, error) {
	// Ensure appTempDir is initialized, even if EnsureResourceFiles hasn't been called explicitly
	// by the top-level main, though it's expected to be.
	if _, err := getAppTempDir(); err != nil {
		return "", fmt.Errorf("app temp directory not initialized: %w", err)
	}

	for _, res := range requiredResourceFiles {
		if res.Name == name {
			if res.LocalPath == "" {
				// This case should ideally not happen if EnsureResourceFiles was called.
				return "", fmt.Errorf("LocalPath for resource %s not set. EnsureResourceFiles might not have run or completed successfully", name)
			}
			return res.LocalPath, nil
		}
	}
	return "", fmt.Errorf("resource file %s not defined in requiredResourceFiles", name)
}

// Future resource management code will go here. 