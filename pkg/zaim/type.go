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
	ProfileImageURL string `json:"profile_image_url"`
	InputCount      int    `json:"input_count"`
	RepeatCount     int    `json:"repeat_count"`
	Day             string `json:"day"`
}

type Money struct {
	ID            int    `json:"id"`
	Mode          string `json:"mode"`
	UserID        int    `json:"user_id"`
	Date          string `json:"date"`
	CategoryID    int    `json:"category_id"`
	GenreID       int    `json:"genre_id"`
	AccountID     int    `json:"account_id"`
	Amount        int    `json:"amount"`
	Comment       string `json:"comment"`
	Active        int    `json:"active"`
	Created       string `json:"created"`
	CurrencyCode  string `json:"currency_code"`
	FromAccountID int    `json:"from_account_id"`
	ToAccountID   int    `json:"to_account_id"`
}

type Category struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Mode     string `json:"mode"`
	Sort     int    `json:"sort"`
	Active   int    `json:"active"`
	Created  string `json:"created"`
	Modified string `json:"modified"`
}

type Genre struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	CategoryID int    `json:"category_id"`
	Mode       string `json:"mode"`
	Sort       int    `json:"sort"`
	Active     int    `json:"active"`
	Created    string `json:"created"`
	Modified   string `json:"modified"`
}

type Account struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Mode     string `json:"mode"`
	Sort     int    `json:"sort"`
	Active   int    `json:"active"`
	Created  string `json:"created"`
	Modified string `json:"modified"`
}

type Currency struct {
	Code string `json:"code"`
	Name string `json:"name"`
}
