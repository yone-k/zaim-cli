package update

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type mockHTTPClient struct {
	response *http.Response
	err      error
	calls    int
}

func (m *mockHTTPClient) Get(url string) (*http.Response, error) {
	m.calls++
	return m.response, m.err
}

func TestCheckForUpdate_shouldReturnResultWithMessage_whenNewVersionIsAvailable(t *testing.T) {
	t.Parallel()

	// Given: v0.3.0 is published on GitHub and the current version is v0.2.0.
	cacheFilePath := filepath.Join(t.TempDir(), "update-check.json")
	httpClient := &mockHTTPClient{
		response: githubReleaseResponse(`{"tag_name":"v0.3.0"}`),
	}

	// When: checking for an update.
	got, err := CheckForUpdate("v0.2.0", cacheFilePath, httpClient)

	// Then: an update result is returned with a user-facing install message.
	if err != nil {
		t.Fatalf("CheckForUpdate() error = %v, want nil", err)
	}
	if got == nil {
		t.Fatal("CheckForUpdate() result = nil, want non-nil")
	}
	if got.CurrentVersion != "v0.2.0" {
		t.Fatalf("CurrentVersion = %q, want %q", got.CurrentVersion, "v0.2.0")
	}
	if got.LatestVersion != "v0.3.0" {
		t.Fatalf("LatestVersion = %q, want %q", got.LatestVersion, "v0.3.0")
	}
	if !strings.Contains(got.Message, "v0.3.0") {
		t.Fatalf("Message = %q, want it to contain latest version %q", got.Message, "v0.3.0")
	}
	if !strings.Contains(got.Message, "npm install -g @yone_k/zaim-cli") {
		t.Fatalf("Message = %q, want it to contain npm install command", got.Message)
	}
}

func TestCheckForUpdate_shouldReturnNil_whenCurrentVersionIsLatest(t *testing.T) {
	t.Parallel()

	// Given: GitHub latest release matches the current version.
	cacheFilePath := filepath.Join(t.TempDir(), "update-check.json")
	httpClient := &mockHTTPClient{
		response: githubReleaseResponse(`{"tag_name":"v0.2.0"}`),
	}

	// When: checking for an update.
	got, err := CheckForUpdate("v0.2.0", cacheFilePath, httpClient)

	// Then: no update result is returned.
	if err != nil {
		t.Fatalf("CheckForUpdate() error = %v, want nil", err)
	}
	if got != nil {
		t.Fatalf("CheckForUpdate() result = %+v, want nil", got)
	}
}

func TestCheckForUpdate_shouldReturnNil_whenCurrentVersionIsDev(t *testing.T) {
	t.Parallel()

	// Given: the current version is a development build.
	cacheFilePath := filepath.Join(t.TempDir(), "update-check.json")
	httpClient := &mockHTTPClient{
		response: githubReleaseResponse(`{"tag_name":"v0.3.0"}`),
	}

	// When: checking for an update.
	got, err := CheckForUpdate("dev", cacheFilePath, httpClient)

	// Then: update checks are skipped.
	if err != nil {
		t.Fatalf("CheckForUpdate() error = %v, want nil", err)
	}
	if got != nil {
		t.Fatalf("CheckForUpdate() result = %+v, want nil", got)
	}
}

func TestCheckForUpdate_shouldReturnNilWithoutCallingAPI_whenCheckedWithin24Hours(t *testing.T) {
	t.Parallel()

	// Given: the cache says the last update check ran within the past 24 hours.
	cacheFilePath := filepath.Join(t.TempDir(), "update-check.json")
	cacheContent := `{"last_check":"2026-04-11T00:00:00Z"}`
	if err := os.WriteFile(cacheFilePath, []byte(cacheContent), 0o600); err != nil {
		t.Fatalf("WriteFile(cacheFilePath) error = %v", err)
	}
	httpClient := &mockHTTPClient{
		response: githubReleaseResponse(`{"tag_name":"v0.3.0"}`),
	}

	// When: checking for an update.
	got, err := CheckForUpdate("v0.2.0", cacheFilePath, httpClient)

	// Then: the cached check suppresses the API call.
	if err != nil {
		t.Fatalf("CheckForUpdate() error = %v, want nil", err)
	}
	if got != nil {
		t.Fatalf("CheckForUpdate() result = %+v, want nil", got)
	}
	if httpClient.calls != 0 {
		t.Fatalf("httpClient.Get calls = %d, want 0", httpClient.calls)
	}
}

func TestCheckForUpdate_shouldReturnNilAndNil_whenNetworkErrorOccurs(t *testing.T) {
	t.Parallel()

	// Given: GitHub API cannot be reached.
	cacheFilePath := filepath.Join(t.TempDir(), "update-check.json")
	httpClient := &mockHTTPClient{
		err: errors.New("network unavailable"),
	}

	// When: checking for an update.
	got, err := CheckForUpdate("v0.2.0", cacheFilePath, httpClient)

	// Then: the failure is silently skipped.
	if err != nil {
		t.Fatalf("CheckForUpdate() error = %v, want nil", err)
	}
	if got != nil {
		t.Fatalf("CheckForUpdate() result = %+v, want nil", got)
	}
}

func TestCheckForUpdate_shouldReturnNilAndNil_whenGitHubAPIReturnsInvalidJSON(t *testing.T) {
	t.Parallel()

	// Given: GitHub API returns an invalid JSON body.
	cacheFilePath := filepath.Join(t.TempDir(), "update-check.json")
	httpClient := &mockHTTPClient{
		response: githubReleaseResponse(`{"tag_name":`),
	}

	// When: checking for an update.
	got, err := CheckForUpdate("v0.2.0", cacheFilePath, httpClient)

	// Then: the invalid response is silently skipped.
	if err != nil {
		t.Fatalf("CheckForUpdate() error = %v, want nil", err)
	}
	if got != nil {
		t.Fatalf("CheckForUpdate() result = %+v, want nil", got)
	}
}

func githubReleaseResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}
