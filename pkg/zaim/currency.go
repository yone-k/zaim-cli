package zaim

import (
	"context"
	"encoding/json"
	"net/http"
)

type currenciesResponse struct {
	Currencies []Currency `json:"currencies"`
}

func (c *Client) ListCurrencies(ctx context.Context) ([]Currency, error) {
	resp, err := c.do(ctx, http.MethodGet, "/v2/currency", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var body currenciesResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	return body.Currencies, nil
}
