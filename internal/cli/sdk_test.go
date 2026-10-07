package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"sync/atomic"
	"testing"

	"github.com/yone-k/go-zaim"
)

func TestMoneyListWithExternalSDK(t *testing.T) {
	for _, tc := range []struct {
		name       string
		status     int
		body       string
		wantOutput []map[string]any
		wantError  string
	}{
		{
			name:   "filtered records retain JSON output and limit",
			status: http.StatusOK,
			body:   `{"money":[{"id":1,"amount":1200,"mode":"payment","name":"lunch"},{"id":2,"amount":900}]}`,
			wantOutput: []map[string]any{{
				"id": float64(1), "amount": float64(1200), "mode": "payment", "name": "lunch",
			}},
		},
		{
			name:      "API failures retain the error and produce no JSON output",
			status:    http.StatusForbidden,
			body:      `{"error":"permission denied"}`,
			wantError: `request failed: status=403 body={"error":"permission denied"}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				if r.Method != http.MethodGet || r.URL.Path != "/v2/home/money" {
					t.Errorf("request = %s %s", r.Method, r.URL.Path)
				}
				wantQuery := url.Values{
					"mapping": {"1"}, "mode": {"payment"}, "category_id": {"101"}, "limit": {"1"}, "page": {"2"},
				}
				if !reflect.DeepEqual(r.URL.Query(), wantQuery) {
					t.Errorf("query = %v, want %v", r.URL.Query(), wantQuery)
				}
				if r.Header.Get("Authorization") == "" {
					t.Error("OAuth authorization header is missing")
				}
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()

			previousClient, previousFormat := Client, OutputFormat
			previousMode, previousCategory := moneyListMode, moneyListCategoryID
			previousLimit, previousPage := moneyListLimit, moneyListPage
			previousStart, previousEnd := moneyListStartDate, moneyListEndDate
			previousWriter, previousContext := moneyListCmd.OutOrStdout(), moneyListCmd.Context()
			t.Cleanup(func() {
				Client, OutputFormat = previousClient, previousFormat
				moneyListMode, moneyListCategoryID = previousMode, previousCategory
				moneyListLimit, moneyListPage = previousLimit, previousPage
				moneyListStartDate, moneyListEndDate = previousStart, previousEnd
				moneyListCmd.SetOut(previousWriter)
				moneyListCmd.SetContext(previousContext)
			})
			Client = zaim.NewWithOptions(zaim.OAuthConfig{
				ConsumerKey: "key", ConsumerSecret: "secret", AccessToken: "token", AccessTokenSecret: "token-secret",
			}, zaim.ClientOptions{BaseURL: server.URL, HTTPClient: server.Client()})
			OutputFormat = "json"
			moneyListMode, moneyListCategoryID = "payment", 101
			moneyListLimit, moneyListPage = 1, 2
			moneyListStartDate, moneyListEndDate = "", ""
			var output bytes.Buffer
			moneyListCmd.SetOut(&output)
			moneyListCmd.SetContext(context.Background())

			err := moneyListCmd.RunE(moneyListCmd, nil)
			if calls.Load() != 1 {
				t.Fatalf("API calls = %d, want 1", calls.Load())
			}
			if tc.wantError != "" {
				if err == nil || err.Error() != tc.wantError || output.Len() != 0 {
					t.Fatalf("error = %v, output = %q", err, output.String())
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			var got []map[string]any
			if err := json.Unmarshal(output.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if len(got) != len(tc.wantOutput) {
				t.Fatalf("records = %d, want %d", len(got), len(tc.wantOutput))
			}
			for key, want := range tc.wantOutput[0] {
				if got[0][key] != want {
					t.Errorf("record.%s = %v, want %v", key, got[0][key], want)
				}
			}
		})
	}
}
