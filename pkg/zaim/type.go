package zaim

type OAuthConfig struct {
	ConsumerKey       string `json:"consumer_key"`
	ConsumerSecret    string `json:"consumer_secret"`
	AccessToken       string `json:"access_token"`
	AccessTokenSecret string `json:"access_token_secret"`
}

type User struct {
	ID              int    `json:"id"`
	Login           string `json:"login"`
	Name            string `json:"name"`
	InputCount      int    `json:"input_count"`
	DayCount        int    `json:"day_count"`
	RepeatCount     int    `json:"repeat_count"`
	Day             int    `json:"day"`
	Week            int    `json:"week"`
	Month           int    `json:"month"`
	CurrencyCode    string `json:"currency_code"`
	ProfileImageURL string `json:"profile_image_url"`
	CoverImageURL   string `json:"cover_image_url"`
	ProfileModified string `json:"profile_modified"`
}

type Money struct {
	ID            int    `json:"id"`
	Mode          string `json:"mode"`
	UserID        int    `json:"user_id"`
	Date          string `json:"date"`
	CategoryID    int    `json:"category_id"`
	GenreID       int    `json:"genre_id"`
	FromAccountID int    `json:"from_account_id"`
	ToAccountID   int    `json:"to_account_id"`
	Amount        int    `json:"amount"`
	Comment       string `json:"comment"`
	Active        int    `json:"active"`
	Name          string `json:"name"`
	ReceiptID     int    `json:"receipt_id"`
	Place         string `json:"place"`
	Created       string `json:"created"`
	CurrencyCode  string `json:"currency_code"`
}

type Category struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	Mode             string `json:"mode"`
	Sort             int    `json:"sort"`
	ParentCategoryID int    `json:"parent_category_id"`
	Active           int    `json:"active"`
	Modified         string `json:"modified"`
}

type Genre struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Sort          int    `json:"sort"`
	Active        int    `json:"active"`
	CategoryID    int    `json:"category_id"`
	ParentGenreID int    `json:"parent_genre_id"`
	Modified      string `json:"modified"`
}

type Account struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	Modified        string `json:"modified"`
	Sort            int    `json:"sort"`
	Active          int    `json:"active"`
	LocalID         int    `json:"local_id"`
	WebsiteID       int    `json:"website_id"`
	ParentAccountID int    `json:"parent_account_id"`
}

type Currency struct {
	CurrencyCode string `json:"currency_code"`
	Name         string `json:"name"`
	Unit         string `json:"unit"`
	Point        int    `json:"point"`
}
