package zaim

import (
	"context"
	"encoding/json"
	"net/http"
)

type userVerifyResponse struct {
	Me *User `json:"me"`
}

func (c *Client) VerifyAuth(ctx context.Context) (*User, error) {
	resp, err := c.do(ctx, http.MethodGet, "/v2/home/user/verify", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var body userVerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	return body.Me, nil
}
