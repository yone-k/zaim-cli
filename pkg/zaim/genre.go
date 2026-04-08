package zaim

import (
	"context"
	"encoding/json"
	"net/http"
)

type genresResponse struct {
	Genres []Genre `json:"genres"`
}

func (c *Client) ListUserGenres(ctx context.Context) ([]Genre, error) {
	resp, err := c.do(ctx, http.MethodGet, "/v2/home/genre", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var body genresResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	return body.Genres, nil
}

func (c *Client) ListDefaultGenres(ctx context.Context, mode string) ([]Genre, error) {
	resp, err := c.do(ctx, http.MethodGet, "/v2/genre", map[string]string{
		"mode": mode,
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var body genresResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	return body.Genres, nil
}
