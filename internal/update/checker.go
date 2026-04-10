package update

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const latestReleaseURL = "https://api.github.com/repos/yone-k/zaim-cli/releases/latest"

// CheckResult はアップデート確認の結果
type CheckResult struct {
	CurrentVersion string
	LatestVersion  string
	Message        string
}

type HTTPClient interface {
	Get(url string) (*http.Response, error)
}

// CheckForUpdate はアップデートを確認する
// currentVersion: 現在のバージョン（例: "v0.2.0"）
// cacheFilePath: キャッシュファイルのパス
// httpClient: HTTP GET用のインターフェース（テスト用にモック可能）
func CheckForUpdate(currentVersion string, cacheFilePath string, httpClient HTTPClient) (*CheckResult, error) {
	if currentVersion == "dev" {
		return nil, nil
	}

	if checkedWithin24Hours(cacheFilePath) {
		return nil, nil
	}

	response, err := httpClient.Get(latestReleaseURL)
	if err != nil {
		return nil, nil
	}
	if response == nil || response.Body == nil {
		return nil, nil
	}
	defer response.Body.Close()

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(response.Body).Decode(&release); err != nil {
		return nil, nil
	}
	if release.TagName == "" {
		return nil, nil
	}

	if err := writeCache(cacheFilePath, time.Now().UTC()); err != nil {
		return nil, err
	}

	if normalizeVersion(currentVersion) == normalizeVersion(release.TagName) {
		return nil, nil
	}

	return &CheckResult{
		CurrentVersion: currentVersion,
		LatestVersion:  release.TagName,
		Message: fmt.Sprintf(
			"新しいバージョン %s が利用可能です。現在のバージョン: %s\n  npm install -g @yone_k/zaim-cli",
			release.TagName,
			currentVersion,
		),
	}, nil
}

func checkedWithin24Hours(cacheFilePath string) bool {
	content, err := os.ReadFile(cacheFilePath)
	if err != nil {
		return false
	}

	var cache struct {
		LastCheck string `json:"last_check"`
	}
	if err := json.Unmarshal(content, &cache); err != nil {
		return false
	}

	lastCheck, err := time.Parse(time.RFC3339, cache.LastCheck)
	if err != nil {
		return false
	}

	return time.Since(lastCheck) < 24*time.Hour
}

func writeCache(cacheFilePath string, checkedAt time.Time) error {
	if err := os.MkdirAll(filepath.Dir(cacheFilePath), 0o755); err != nil {
		return err
	}

	content, err := json.Marshal(struct {
		LastCheck string `json:"last_check"`
	}{
		LastCheck: checkedAt.Format(time.RFC3339),
	})
	if err != nil {
		return err
	}

	return os.WriteFile(cacheFilePath, content, 0o600)
}

func normalizeVersion(version string) string {
	if strings.HasPrefix(version, "v") {
		return version
	}
	return "v" + version
}
