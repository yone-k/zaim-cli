package zaim

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	oauthRequestTokenURL = "https://api.zaim.net/v2/auth/request"
	oauthAccessTokenURL  = "https://api.zaim.net/v2/auth/access"
)

func NormalizeParameters(params map[string]string) string {
	type pair struct {
		name  string
		value string
	}

	pairs := make([]pair, 0, len(params))
	for key, value := range params {
		pairs = append(pairs, pair{
			name:  percentEncode(key),
			value: percentEncode(value),
		})
	}

	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].name == pairs[j].name {
			return pairs[i].value < pairs[j].value
		}
		return pairs[i].name < pairs[j].name
	})

	parts := make([]string, 0, len(pairs))
	for _, pair := range pairs {
		parts = append(parts, pair.name+"="+pair.value)
	}

	return strings.Join(parts, "&")
}

func ConstructBaseString(method, rawURL, normalizedParams string) string {
	return strings.ToUpper(method) + "&" + percentEncode(normalizeBaseURL(rawURL)) + "&" + percentEncode(normalizedParams)
}

func ConstructSigningKey(consumerSecret, tokenSecret string) string {
	return percentEncode(consumerSecret) + "&" + percentEncode(tokenSecret)
}

func GenerateSignature(method, rawURL string, params map[string]string, consumerSecret, tokenSecret string) string {
	normalizedParams := NormalizeParameters(params)
	baseString := ConstructBaseString(method, rawURL, normalizedParams)
	signingKey := ConstructSigningKey(consumerSecret, tokenSecret)

	mac := hmac.New(sha1.New, []byte(signingKey))
	_, _ = mac.Write([]byte(baseString))

	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func RequestToken(ctx context.Context, consumerKey, consumerSecret, callbackURL string) (oauthToken, oauthTokenSecret string, err error) {
	params, err := newOAuthParams(consumerKey)
	if err != nil {
		return "", "", err
	}
	params["oauth_callback"] = callbackURL
	params["oauth_signature"] = GenerateSignature(http.MethodPost, oauthRequestTokenURL, params, consumerSecret, "")

	values, err := doOAuthPOST(ctx, oauthRequestTokenURL, params)
	if err != nil {
		return "", "", err
	}

	return values.Get("oauth_token"), values.Get("oauth_token_secret"), nil
}

func GetAuthorizeURL(oauthToken string) string {
	return "https://auth.zaim.net/users/auth?oauth_token=" + oauthToken
}

func ExchangeAccessToken(ctx context.Context, consumerKey, consumerSecret, oauthToken, oauthTokenSecret, oauthVerifier string) (accessToken, accessTokenSecret string, err error) {
	params, err := newOAuthParams(consumerKey)
	if err != nil {
		return "", "", err
	}
	params["oauth_token"] = oauthToken
	params["oauth_verifier"] = oauthVerifier
	params["oauth_signature"] = GenerateSignature(http.MethodPost, oauthAccessTokenURL, params, consumerSecret, oauthTokenSecret)

	values, err := doOAuthPOST(ctx, oauthAccessTokenURL, params)
	if err != nil {
		return "", "", err
	}

	return values.Get("oauth_token"), values.Get("oauth_token_secret"), nil
}

func newOAuthParams(consumerKey string) (map[string]string, error) {
	nonce, err := generateNonce()
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"oauth_consumer_key":     consumerKey,
		"oauth_nonce":            nonce,
		"oauth_signature_method": "HMAC-SHA1",
		"oauth_timestamp":        strconv.FormatInt(time.Now().Unix(), 10),
		"oauth_version":          "1.0",
	}, nil
}

func generateNonce() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate oauth nonce: %w", err)
	}
	return hex.EncodeToString(buf), nil
}

func doOAuthPOST(ctx context.Context, endpoint string, authParams map[string]string) (url.Values, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build oauth request: %w", err)
	}
	req.Header.Set("Authorization", buildAuthorizationHeader(authParams))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send oauth request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read oauth response: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("oauth request failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	values, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, fmt.Errorf("parse oauth response: %w", err)
	}

	return values, nil
}

func buildAuthorizationHeader(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+`="`+percentEncode(params[key])+`"`)
	}

	return "OAuth " + strings.Join(parts, ", ")
}

func normalizeBaseURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	scheme := strings.ToLower(parsed.Scheme)
	host := strings.ToLower(parsed.Hostname())
	port := parsed.Port()

	if port != "" && !isDefaultPort(scheme, port) {
		host += ":" + port
	}

	path := parsed.EscapedPath()
	if path == "" {
		path = "/"
	}

	return scheme + "://" + host + path
}

func isDefaultPort(scheme, port string) bool {
	return (scheme == "http" && port == "80") || (scheme == "https" && port == "443")
}

func percentEncode(value string) string {
	var builder strings.Builder
	builder.Grow(len(value) * 3)

	for i := 0; i < len(value); i++ {
		b := value[i]
		if isUnreserved(b) {
			builder.WriteByte(b)
			continue
		}
		builder.WriteByte('%')
		builder.WriteByte(upperHex(b >> 4))
		builder.WriteByte(upperHex(b & 0x0F))
	}

	return builder.String()
}

func isUnreserved(b byte) bool {
	return (b >= 'A' && b <= 'Z') ||
		(b >= 'a' && b <= 'z') ||
		(b >= '0' && b <= '9') ||
		b == '-' || b == '.' || b == '_' || b == '~'
}

func upperHex(n byte) byte {
	if n < 10 {
		return '0' + n
	}
	return 'A' + (n - 10)
}
