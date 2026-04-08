package cli

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/yone/zaim-cli/internal/config"
	"github.com/yone/zaim-cli/pkg/zaim"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Zaim OAuth認証を管理",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Zaimにログインしアクセストークンを取得",
	RunE: func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()

		var err error
		consumerKey, err = promptCredential("Consumer Key", consumerKey)
		if err != nil {
			return err
		}

		consumerSecret, err = promptCredential("Consumer Secret", consumerSecret)
		if err != nil {
			return err
		}

		port, _ := cmd.Flags().GetInt("port")
		listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
		if err != nil {
			return fmt.Errorf("callback listenerの起動に失敗しました: %w", err)
		}
		defer listener.Close()

		callbackURL := fmt.Sprintf("http://localhost:%d/callback", port)
		oauthToken, oauthTokenSecret, err := zaim.RequestToken(ctx, consumerKey, consumerSecret, callbackURL)
		if err != nil {
			return fmt.Errorf("リクエストトークンの取得に失敗しました: %w", err)
		}

		authorizeURL := zaim.GetAuthorizeURL(oauthToken)
		fmt.Fprintf(cmd.OutOrStdout(), "ブラウザで認証を進めてください: %s\n", authorizeURL)

		if runtime.GOOS == "darwin" {
			if err := exec.Command("open", authorizeURL).Start(); err != nil {
				return fmt.Errorf("ブラウザ起動に失敗しました: %w", err)
			}
		}

		callbackResult, err := waitOAuthCallback(ctx, listener)
		if err != nil {
			return err
		}
		if callbackResult.oauthToken != oauthToken {
			return fmt.Errorf("callbackのoauth_tokenが一致しません")
		}

		accessToken, accessTokenSecret, err := zaim.ExchangeAccessToken(
			ctx,
			consumerKey,
			consumerSecret,
			oauthToken,
			oauthTokenSecret,
			callbackResult.oauthVerifier,
		)
		if err != nil {
			return fmt.Errorf("アクセストークンの取得に失敗しました: %w", err)
		}

		if err := config.Save(&config.Config{
			ConsumerKey:       consumerKey,
			ConsumerSecret:    consumerSecret,
			AccessToken:       accessToken,
			AccessTokenSecret: accessTokenSecret,
		}); err != nil {
			return fmt.Errorf("設定保存に失敗しました: %w", err)
		}

		fmt.Fprintln(cmd.OutOrStdout(), "認証情報を保存しました。")

		return nil
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "認証状態を確認",
	RunE: func(cmd *cobra.Command, _ []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("設定読み込みに失敗しました: %w", err)
		}

		client := zaim.New(zaim.OAuthConfig{
			ConsumerKey:       cfg.ConsumerKey,
			ConsumerSecret:    cfg.ConsumerSecret,
			AccessToken:       cfg.AccessToken,
			AccessTokenSecret: cfg.AccessTokenSecret,
		})

		user, err := client.VerifyAuth(cmd.Context())
		if err != nil {
			return fmt.Errorf("認証確認に失敗しました: %w", err)
		}

		name := user.Name
		if name == "" {
			name = user.Login
		}

		fmt.Fprintf(cmd.OutOrStdout(), "認証済み: %s (ID: %d)\n", name, user.ID)

		return nil
	},
}

type oauthCallbackResult struct {
	oauthToken    string
	oauthVerifier string
}

func init() {
	authLoginCmd.Flags().Int("port", 8080, "コールバック用ローカルサーバーのポート番号")
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authStatusCmd)
}

func promptCredential(label, current string) (string, error) {
	if current != "" {
		return current, nil
	}

	fmt.Printf("%s: ", label)

	var value string
	if _, err := fmt.Scan(&value); err != nil {
		return "", fmt.Errorf("%sの入力に失敗しました: %w", label, err)
	}
	if value == "" {
		return "", fmt.Errorf("%sが空です", label)
	}

	return value, nil
}

func waitOAuthCallback(ctx context.Context, listener net.Listener) (*oauthCallbackResult, error) {
	callbackCh := make(chan oauthCallbackResult, 1)
	serverErrCh := make(chan error, 1)

	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			if r.URL.Path != "/callback" {
				http.NotFound(w, r)
				return
			}

			oauthToken := r.URL.Query().Get("oauth_token")
			oauthVerifier := r.URL.Query().Get("oauth_verifier")
			if oauthToken == "" || oauthVerifier == "" {
				http.Error(w, "missing oauth callback parameters", http.StatusBadRequest)
				return
			}

			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = fmt.Fprint(w, "<html><body>認証成功。ブラウザを閉じてください。</body></html>")

			callbackCh <- oauthCallbackResult{
				oauthToken:    oauthToken,
				oauthVerifier: oauthVerifier,
			}
		}),
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	go func() {
		err := server.Serve(listener)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrCh <- err
		}
	}()

	select {
	case result := <-callbackCh:
		if err := server.Shutdown(ctx); err != nil {
			return nil, fmt.Errorf("callback serverの停止に失敗しました: %w", err)
		}
		return &result, nil
	case err := <-serverErrCh:
		return nil, fmt.Errorf("callback serverでエラーが発生しました: %w", err)
	case <-ctx.Done():
		_ = server.Shutdown(context.Background())
		return nil, ctx.Err()
	}
}
