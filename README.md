# zaim-cli

Zaim家計簿APIを操作するCLIツール。Go SDKとCLIを提供。

## インストール

```bash
go install github.com/yone/zaim-cli/cmd/zaim@latest
```

## セットアップ

### 1. Zaim Developer登録
[Zaim Developers](https://dev.zaim.net/) でアプリケーションを登録し、Consumer KeyとConsumer Secretを取得。

### 2. 認証
```bash
zaim auth login
```
ブラウザが開き、Zaimの認証ページに遷移。認証後、アクセストークンが `~/.config/zaim/config.json` に保存される。

### 3. 認証状態の確認
```bash
zaim auth status
```

## 使い方

### ユーザー情報
```bash
zaim user
zaim user --output json
```

### 家計簿レコード
```bash
# 一覧
zaim money list --limit 10
zaim money list --mode payment --start-date 2026-01-01 --end-date 2026-01-31

# 支出を記録
zaim money create-payment --amount 1000 --date 2026-01-15 --category-id 101 --genre-id 201

# 収入を記録
zaim money create-income --amount 50000 --date 2026-01-25 --category-id 11

# 振替を記録
zaim money create-transfer --amount 10000 --date 2026-01-20 --from-account-id 1 --to-account-id 2

# 更新
zaim money update 12345 payment --comment "修正済み"

# 削除
zaim money delete 12345 payment
```

### マスターデータ
```bash
zaim category list
zaim category defaults --mode payment
zaim genre list
zaim genre defaults --mode income
zaim account list
zaim currency list
```

### 出力フォーマット
すべてのコマンドで `--output` フラグが使用可能:
- `table`（デフォルト）: テーブル形式
- `json`: JSON形式

## 環境変数
設定ファイルの代わりに環境変数でも認証情報を指定可能:
- `ZAIM_CONSUMER_KEY`
- `ZAIM_CONSUMER_SECRET`
- `ZAIM_ACCESS_TOKEN`
- `ZAIM_ACCESS_TOKEN_SECRET`

## SDK
`pkg/zaim` パッケージとして独立したGoクライアントライブラリも提供:

```go
import (
    "context"
    "fmt"
    "log"

    "github.com/yone/zaim-cli/pkg/zaim"
)

func main() {
    client := zaim.New(zaim.OAuthConfig{
        ConsumerKey:       "your-key",
        ConsumerSecret:    "your-secret",
        AccessToken:       "your-token",
        AccessTokenSecret: "your-token-secret",
    })

    user, err := client.VerifyAuth(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("User: %s (ID: %d)\n", user.Name, user.ID)
}
```

## ライセンス
MIT
