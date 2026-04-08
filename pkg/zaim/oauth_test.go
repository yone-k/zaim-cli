package zaim

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestNormalizeParameters_shouldPercentEncodeAndSortParameters_whenParametersContainReservedCharacters(t *testing.T) {
	t.Parallel()

	// Given: RFC 5849 style parameters with reserved characters.
	params := map[string]string{
		"b5":                     "=%253D",
		"a3":                     "a",
		"c@":                     "",
		"a2":                     "r b",
		"oauth_consumer_key":     "9djdj82h48djs9d2",
		"oauth_token":            "kkk9d7dh3k39sjv7",
		"oauth_signature_method": "HMAC-SHA1",
		"oauth_timestamp":        "137131201",
		"oauth_nonce":            "7d8f3e4a",
		"c2":                     "",
	}

	const want = "a2=r%20b&a3=a&b5=%3D%25253D&c%40=&c2=&oauth_consumer_key=9djdj82h48djs9d2&oauth_nonce=7d8f3e4a&oauth_signature_method=HMAC-SHA1&oauth_timestamp=137131201&oauth_token=kkk9d7dh3k39sjv7"

	// When: parameters are normalized for signing.
	got := NormalizeParameters(params)

	// Then: keys are sorted and names/values are percent-encoded.
	if got != want {
		t.Fatalf("NormalizeParameters() = %q, want %q", got, want)
	}
}

func TestConstructBaseString_shouldBuildRFC5849BaseString_whenMethodURLAndNormalizedParametersAreProvided(t *testing.T) {
	t.Parallel()

	// Given: the RFC 5849 example request pieces.
	const method = http.MethodGet
	const rawURL = "http://photos.example.net/photos"
	const normalizedParams = "file=vacation.jpg&oauth_consumer_key=dpf43f3p2l4k3l03&oauth_nonce=kllo9940pd9333jh&oauth_signature_method=HMAC-SHA1&oauth_timestamp=1191242096&oauth_token=nnch734d00sl2jdk&oauth_version=1.0&size=original"
	const want = "GET&http%3A%2F%2Fphotos.example.net%2Fphotos&file%3Dvacation.jpg%26oauth_consumer_key%3Ddpf43f3p2l4k3l03%26oauth_nonce%3Dkllo9940pd9333jh%26oauth_signature_method%3DHMAC-SHA1%26oauth_timestamp%3D1191242096%26oauth_token%3Dnnch734d00sl2jdk%26oauth_version%3D1.0%26size%3Doriginal"

	// When: the signature base string is constructed.
	got := ConstructBaseString(method, rawURL, normalizedParams)

	// Then: METHOD&encoded-URL&encoded-params is returned.
	if got != want {
		t.Fatalf("ConstructBaseString() = %q, want %q", got, want)
	}
}

func TestConstructSigningKey_shouldPercentEncodeSecretsAndJoinWithAmpersand_whenSecretsContainReservedCharacters(t *testing.T) {
	t.Parallel()

	// Given: secrets that require percent-encoding.
	const consumerSecret = "consumer&secret"
	const tokenSecret = "token secret="
	const want = "consumer%26secret&token%20secret%3D"

	// When: the signing key is constructed.
	got := ConstructSigningKey(consumerSecret, tokenSecret)

	// Then: both secrets are encoded and joined by '&'.
	if got != want {
		t.Fatalf("ConstructSigningKey() = %q, want %q", got, want)
	}
}

func TestGenerateSignature_shouldGenerateRFC5849HMACSHA1Signature_whenGivenKnownVector(t *testing.T) {
	t.Parallel()

	// Given: the RFC 5849 example request without duplicate parameter names.
	params := map[string]string{
		"file":                   "vacation.jpg",
		"size":                   "original",
		"oauth_consumer_key":     "dpf43f3p2l4k3l03",
		"oauth_token":            "nnch734d00sl2jdk",
		"oauth_signature_method": "HMAC-SHA1",
		"oauth_timestamp":        "1191242096",
		"oauth_nonce":            "kllo9940pd9333jh",
		"oauth_version":          "1.0",
	}

	const method = http.MethodGet
	const rawURL = "http://photos.example.net/photos"
	const consumerSecret = "kd94hf93k423kf44"
	const tokenSecret = "pfkkdhi9sl3r4s00"
	const want = "tR3+Ty81lMeYAr/Fid0kMTYa/WM="

	// When: the OAuth 1.0a signature is generated.
	got := GenerateSignature(method, rawURL, params, consumerSecret, tokenSecret)

	// Then: the RFC signature vector is reproduced.
	if got != want {
		t.Fatalf("GenerateSignature() = %q, want %q", got, want)
	}
}

func TestRequestToken_shouldPostOAuthAuthorizationHeaderAndParseTokenResponse_whenRequestSucceeds(t *testing.T) {
	t.Parallel()

	// Given: a request token endpoint that validates the outbound request.
	const consumerKey = "consumer-key"
	const consumerSecret = "consumer-secret"
	const callbackURL = "https://example.com/callback"
	const responseToken = "request-token"
	const responseSecret = "request-secret"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("request method = %s, want %s", r.Method, http.MethodPost)
		}

		authParams := parseOAuthAuthorizationHeader(t, r.Header.Get("Authorization"))

		if got := authParams["oauth_consumer_key"]; got != consumerKey {
			t.Fatalf("oauth_consumer_key = %q, want %q", got, consumerKey)
		}
		if got := authParams["oauth_callback"]; got != callbackURL {
			t.Fatalf("oauth_callback = %q, want %q", got, callbackURL)
		}
		if got := authParams["oauth_signature_method"]; got != "HMAC-SHA1" {
			t.Fatalf("oauth_signature_method = %q, want %q", got, "HMAC-SHA1")
		}
		if got := authParams["oauth_signature"]; got == "" {
			t.Fatal("oauth_signature is empty")
		}
		if got := authParams["oauth_nonce"]; got == "" {
			t.Fatal("oauth_nonce is empty")
		}
		if got := authParams["oauth_timestamp"]; got == "" {
			t.Fatal("oauth_timestamp is empty")
		}

		w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
		_, _ = w.Write([]byte("oauth_token=" + responseToken + "&oauth_token_secret=" + responseSecret + "&oauth_callback_confirmed=true"))
	}))
	defer server.Close()

	originalRequestTokenURL := oauthRequestTokenURL
	oauthRequestTokenURL = server.URL + "/request_token"
	t.Cleanup(func() {
		oauthRequestTokenURL = originalRequestTokenURL
	})

	// When: a request token is requested.
	gotToken, gotSecret, err := RequestToken(context.Background(), consumerKey, consumerSecret, callbackURL)

	// Then: the response token pair is returned without error.
	if err != nil {
		t.Fatalf("RequestToken() error = %v", err)
	}
	if gotToken != responseToken {
		t.Fatalf("oauthToken = %q, want %q", gotToken, responseToken)
	}
	if gotSecret != responseSecret {
		t.Fatalf("oauthTokenSecret = %q, want %q", gotSecret, responseSecret)
	}
}

func TestGetAuthorizeURL_shouldReturnZaimAuthorizeEndpoint_whenOAuthTokenIsProvided(t *testing.T) {
	t.Parallel()

	// Given: a request token returned from the request-token step.
	const oauthToken = "request-token"
	const want = "https://www.zaim.net/users/auth?oauth_token=request-token"

	// When: the authorize URL is constructed.
	got := GetAuthorizeURL(oauthToken)

	// Then: the user-facing authorization URL is returned.
	if got != want {
		t.Fatalf("GetAuthorizeURL() = %q, want %q", got, want)
	}
}

func TestExchangeAccessToken_shouldPostOAuthAuthorizationHeaderAndParseAccessTokenResponse_whenRequestSucceeds(t *testing.T) {
	t.Parallel()

	// Given: an access token endpoint that validates the outbound request.
	const consumerKey = "consumer-key"
	const consumerSecret = "consumer-secret"
	const oauthToken = "request-token"
	const oauthTokenSecret = "request-secret"
	const oauthVerifier = "verifier-code"
	const responseToken = "access-token"
	const responseSecret = "access-secret"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("request method = %s, want %s", r.Method, http.MethodPost)
		}

		authParams := parseOAuthAuthorizationHeader(t, r.Header.Get("Authorization"))

		if got := authParams["oauth_consumer_key"]; got != consumerKey {
			t.Fatalf("oauth_consumer_key = %q, want %q", got, consumerKey)
		}
		if got := authParams["oauth_token"]; got != oauthToken {
			t.Fatalf("oauth_token = %q, want %q", got, oauthToken)
		}
		if got := authParams["oauth_verifier"]; got != oauthVerifier {
			t.Fatalf("oauth_verifier = %q, want %q", got, oauthVerifier)
		}
		if got := authParams["oauth_signature_method"]; got != "HMAC-SHA1" {
			t.Fatalf("oauth_signature_method = %q, want %q", got, "HMAC-SHA1")
		}
		if got := authParams["oauth_signature"]; got == "" {
			t.Fatal("oauth_signature is empty")
		}

		w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
		_, _ = w.Write([]byte("oauth_token=" + responseToken + "&oauth_token_secret=" + responseSecret))
	}))
	defer server.Close()

	originalAccessTokenURL := oauthAccessTokenURL
	oauthAccessTokenURL = server.URL + "/access_token"
	t.Cleanup(func() {
		oauthAccessTokenURL = originalAccessTokenURL
	})

	// When: an access token is exchanged.
	gotToken, gotSecret, err := ExchangeAccessToken(
		context.Background(),
		consumerKey,
		consumerSecret,
		oauthToken,
		oauthTokenSecret,
		oauthVerifier,
	)

	// Then: the response token pair is returned without error.
	if err != nil {
		t.Fatalf("ExchangeAccessToken() error = %v", err)
	}
	if gotToken != responseToken {
		t.Fatalf("accessToken = %q, want %q", gotToken, responseToken)
	}
	if gotSecret != responseSecret {
		t.Fatalf("accessTokenSecret = %q, want %q", gotSecret, responseSecret)
	}
}

func parseOAuthAuthorizationHeader(t *testing.T, header string) map[string]string {
	t.Helper()

	if header == "" {
		t.Fatal("Authorization header is empty")
	}
	if !strings.HasPrefix(header, "OAuth ") {
		t.Fatalf("Authorization header = %q, want OAuth prefix", header)
	}

	params := make(map[string]string)
	for _, part := range strings.Split(strings.TrimPrefix(header, "OAuth "), ",") {
		pair := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(pair) != 2 {
			t.Fatalf("invalid OAuth header segment: %q", part)
		}

		value := strings.Trim(pair[1], `"`)
		decoded, err := url.QueryUnescape(value)
		if err != nil {
			t.Fatalf("failed to decode OAuth header value %q: %v", value, err)
		}

		params[pair[0]] = decoded
	}

	return params
}
