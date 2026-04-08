package zaim

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type ListMoneyOptions struct {
	Mode       string
	StartDate  string
	EndDate    string
	CategoryID int
	GenreID    int
	AccountID  int
	Limit      int
	Page       int
}

type moneyListResponse struct {
	Money []Money `json:"money"`
}

type CreatePaymentRequest struct {
	Amount        int
	Date          string
	CategoryID    int
	GenreID       int
	FromAccountID int
	Place         string
	Comment       string
	Name          string
}

type CreateIncomeRequest struct {
	Amount      int
	Date        string
	CategoryID  int
	ToAccountID int
	Place       string
	Comment     string
}

type CreateTransferRequest struct {
	Amount        int
	Date          string
	FromAccountID int
	ToAccountID   int
	Comment       string
}

type UpdateMoneyRequest struct {
	Amount        *int
	Date          *string
	CategoryID    *int
	GenreID       *int
	FromAccountID *int
	ToAccountID   *int
	Place         *string
	Comment       *string
	Name          *string
}

func (c *Client) ListMoney(ctx context.Context, opts *ListMoneyOptions) ([]Money, error) {
	params := map[string]string{}
	if opts != nil {
		addStringParamIfNotEmpty(params, "mode", opts.Mode)
		addStringParamIfNotEmpty(params, "start_date", opts.StartDate)
		addStringParamIfNotEmpty(params, "end_date", opts.EndDate)
		addIntParamIfNotZero(params, "category_id", opts.CategoryID)
		addIntParamIfNotZero(params, "genre_id", opts.GenreID)
		addIntParamIfNotZero(params, "account_id", opts.AccountID)
		addIntParamIfNotZero(params, "limit", opts.Limit)
		addIntParamIfNotZero(params, "page", opts.Page)
	}

	resp, err := c.do(ctx, http.MethodGet, "/v2/home/money", params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var body moneyListResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	return body.Money, nil
}

func (c *Client) CreatePayment(ctx context.Context, req *CreatePaymentRequest) error {
	params := map[string]string{"mapping": "1"}
	if req != nil {
		addIntParamIfNotZero(params, "amount", req.Amount)
		addStringParamIfNotEmpty(params, "date", req.Date)
		addIntParamIfNotZero(params, "category_id", req.CategoryID)
		addIntParamIfNotZero(params, "genre_id", req.GenreID)
		addIntParamIfNotZero(params, "from_account_id", req.FromAccountID)
		addStringParamIfNotEmpty(params, "place", req.Place)
		addStringParamIfNotEmpty(params, "comment", req.Comment)
		addStringParamIfNotEmpty(params, "name", req.Name)
	}

	resp, err := c.do(ctx, http.MethodPost, "/v2/home/money/payment", params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (c *Client) CreateIncome(ctx context.Context, req *CreateIncomeRequest) error {
	params := map[string]string{"mapping": "1"}
	if req != nil {
		addIntParamIfNotZero(params, "amount", req.Amount)
		addStringParamIfNotEmpty(params, "date", req.Date)
		addIntParamIfNotZero(params, "category_id", req.CategoryID)
		addIntParamIfNotZero(params, "to_account_id", req.ToAccountID)
		addStringParamIfNotEmpty(params, "place", req.Place)
		addStringParamIfNotEmpty(params, "comment", req.Comment)
	}

	resp, err := c.do(ctx, http.MethodPost, "/v2/home/money/income", params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (c *Client) CreateTransfer(ctx context.Context, req *CreateTransferRequest) error {
	params := map[string]string{"mapping": "1"}
	if req != nil {
		addIntParamIfNotZero(params, "amount", req.Amount)
		addStringParamIfNotEmpty(params, "date", req.Date)
		addIntParamIfNotZero(params, "from_account_id", req.FromAccountID)
		addIntParamIfNotZero(params, "to_account_id", req.ToAccountID)
		addStringParamIfNotEmpty(params, "comment", req.Comment)
	}

	resp, err := c.do(ctx, http.MethodPost, "/v2/home/money/transfer", params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (c *Client) UpdateMoney(ctx context.Context, id int, mode string, req *UpdateMoneyRequest) error {
	params := map[string]string{}
	if req != nil {
		addIntPointerParam(params, "amount", req.Amount)
		addStringPointerParam(params, "date", req.Date)
		addIntPointerParam(params, "category_id", req.CategoryID)
		addIntPointerParam(params, "genre_id", req.GenreID)
		addIntPointerParam(params, "from_account_id", req.FromAccountID)
		addIntPointerParam(params, "to_account_id", req.ToAccountID)
		addStringPointerParam(params, "place", req.Place)
		addStringPointerParam(params, "comment", req.Comment)
		addStringPointerParam(params, "name", req.Name)
	}

	resp, err := c.do(ctx, http.MethodPut, fmt.Sprintf("/v2/home/money/%s/%d", mode, id), params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (c *Client) DeleteMoney(ctx context.Context, id int, mode string) error {
	resp, err := c.do(ctx, http.MethodDelete, fmt.Sprintf("/v2/home/money/%s/%d", mode, id), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func addStringParamIfNotEmpty(params map[string]string, key string, value string) {
	if value == "" {
		return
	}
	params[key] = value
}

func addIntParamIfNotZero(params map[string]string, key string, value int) {
	if value == 0 {
		return
	}
	params[key] = strconv.Itoa(value)
}

func addStringPointerParam(params map[string]string, key string, value *string) {
	if value == nil {
		return
	}
	params[key] = *value
}

func addIntPointerParam(params map[string]string, key string, value *int) {
	if value == nil {
		return
	}
	params[key] = strconv.Itoa(*value)
}
