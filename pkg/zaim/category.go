package zaim

import (
	"context"
	"encoding/json"
	"net/http"
)

type categoriesResponse struct {
	Categories []Category `json:"categories"`
}

func (c *Client) ListUserCategories(ctx context.Context) ([]Category, error) {
	resp, err := c.do(ctx, http.MethodGet, "/v2/home/category", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var body categoriesResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	return body.Categories, nil
}

func (c *Client) ListDefaultCategories(ctx context.Context, mode string) ([]Category, error) {
	resp, err := c.do(ctx, http.MethodGet, "/v2/category", map[string]string{
		"mode": mode,
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var body categoriesResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	return body.Categories, nil
}
