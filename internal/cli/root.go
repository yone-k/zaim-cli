package cli

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yone-k/zaim-cli/internal/config"
	"github.com/yone-k/zaim-cli/internal/update"
	"github.com/yone-k/zaim-cli/internal/version"
	"github.com/yone-k/zaim-cli/pkg/zaim"
)

var (
	Client       *zaim.Client
	OutputFormat string

	consumerKey       string
	consumerSecret    string
	accessToken       string
	accessTokenSecret string
	updateResultCh    <-chan *update.CheckResult
)

var rootCmd = &cobra.Command{
	Use:          "zaim-cli",
	Short:        "Zaim API CLI tool",
	SilenceUsage: true,
	Version:      buildVersion(),
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		updateResultCh = startUpdateCheck()

		OutputFormat = strings.ToLower(OutputFormat)
		if OutputFormat != "table" && OutputFormat != "json" {
			return fmt.Errorf("invalid output format %q: must be table or json", OutputFormat)
		}

		Client = nil
		if isAuthCommand(cmd) {
			return nil
		}

		oauthConfig, err := resolveOAuthConfig()
		if err != nil {
			return err
		}
		if !hasCompleteOAuthConfig(oauthConfig) {
			return errors.New("missing authentication credentials: set flags, environment variables, or config file")
		}

		Client = zaim.New(oauthConfig)

		return nil
	},
	PersistentPostRunE: func(_ *cobra.Command, _ []string) error {
		if updateResultCh == nil {
			return nil
		}

		if result := <-updateResultCh; result != nil {
			fmt.Fprintln(os.Stderr, result.Message)
		}

		return nil
	},
}

func init() {
	flags := rootCmd.PersistentFlags()
	flags.StringVar(&consumerKey, "consumer-key", "", "Zaim consumer key")
	flags.StringVar(&consumerSecret, "consumer-secret", "", "Zaim consumer secret")
	flags.StringVar(&accessToken, "access-token", "", "Zaim access token")
	flags.StringVar(&accessTokenSecret, "access-token-secret", "", "Zaim access token secret")
	flags.StringVar(&OutputFormat, "output", "table", "Output format (table or json)")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func isAuthCommand(cmd *cobra.Command) bool {
	return cmd.Name() == "auth" || (cmd.Parent() != nil && cmd.Parent().Name() == "auth")
}

func resolveOAuthConfig() (zaim.OAuthConfig, error) {
	cfg := zaim.OAuthConfig{
		ConsumerKey:       consumerKey,
		ConsumerSecret:    consumerSecret,
		AccessToken:       accessToken,
		AccessTokenSecret: accessTokenSecret,
	}

	applyMissingOAuthConfig(&cfg, zaim.OAuthConfig{
		ConsumerKey:       os.Getenv("ZAIM_CONSUMER_KEY"),
		ConsumerSecret:    os.Getenv("ZAIM_CONSUMER_SECRET"),
		AccessToken:       os.Getenv("ZAIM_ACCESS_TOKEN"),
		AccessTokenSecret: os.Getenv("ZAIM_ACCESS_TOKEN_SECRET"),
	})

	fileConfig, err := config.Load()
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return zaim.OAuthConfig{}, fmt.Errorf("failed to load config: %w", err)
		}
	} else {
		applyMissingOAuthConfig(&cfg, zaim.OAuthConfig{
			ConsumerKey:       fileConfig.ConsumerKey,
			ConsumerSecret:    fileConfig.ConsumerSecret,
			AccessToken:       fileConfig.AccessToken,
			AccessTokenSecret: fileConfig.AccessTokenSecret,
		})
	}

	return cfg, nil
}

func applyMissingOAuthConfig(dst *zaim.OAuthConfig, src zaim.OAuthConfig) {
	if dst.ConsumerKey == "" {
		dst.ConsumerKey = src.ConsumerKey
	}
	if dst.ConsumerSecret == "" {
		dst.ConsumerSecret = src.ConsumerSecret
	}
	if dst.AccessToken == "" {
		dst.AccessToken = src.AccessToken
	}
	if dst.AccessTokenSecret == "" {
		dst.AccessTokenSecret = src.AccessTokenSecret
	}
}

func hasCompleteOAuthConfig(cfg zaim.OAuthConfig) bool {
	return cfg.ConsumerKey != "" &&
		cfg.ConsumerSecret != "" &&
		cfg.AccessToken != "" &&
		cfg.AccessTokenSecret != ""
}

func buildVersion() string {
	return version.Version
}

func startUpdateCheck() <-chan *update.CheckResult {
	resultCh := make(chan *update.CheckResult, 1)

	go func() {
		configDir, err := config.GetConfigDir()
		if err != nil {
			resultCh <- nil
			return
		}

		cacheFilePath := filepath.Join(configDir, "update-check.json")
		result, _ := update.CheckForUpdate(version.Version, cacheFilePath, http.DefaultClient)
		resultCh <- result
	}()

	return resultCh
}
