package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yone/zaim-cli/internal/config"
	"github.com/yone/zaim-cli/internal/version"
	"github.com/yone/zaim-cli/pkg/zaim"
)

var (
	Client       *zaim.Client
	OutputFormat string

	consumerKey       string
	consumerSecret    string
	accessToken       string
	accessTokenSecret string
)

var rootCmd = &cobra.Command{
	Use:          "zaim",
	Short:        "Zaim API CLI tool",
	SilenceUsage: true,
	Version:      buildVersion(),
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
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
			return zaim.OAuthConfig{}, err
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
	return fmt.Sprintf("%s (commit: %s, date: %s)", version.Version, version.Commit, version.Date)
}
