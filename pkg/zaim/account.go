package zaim

import (
	"context"
	"encoding/json"
	"net/http"
)

type accountsResponse struct {
	Accounts []Account `json:"accounts"`
}

func (c *Client) ListUserAccounts(ctx context.Context) ([]Account, error) {
	resp, err := c.do(ctx, http.MethodGet, "/v2/home/account", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var body accountsResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	return body.Accounts, nil
}
