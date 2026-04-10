#!/usr/bin/env bash
set -euo pipefail

# ---------------------------------------------------------------------------
# publish-npm.sh
#
# npm パッケージを公開するスクリプト。
# バイナリはパッケージに含めず、postinstall で GitHub Releases からダウンロードする。
# CIから呼び出されることを想定。
#
# Usage:
#   ./scripts/publish-npm.sh v0.1.0 [--dry-run]
# ---------------------------------------------------------------------------

# ---- プロジェクトルート特定 ------------------------------------------------
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# ---- 引数パース ----------------------------------------------------------
VERSION=""
DRY_RUN=false

for arg in "$@"; do
  case "$arg" in
    --dry-run)
      DRY_RUN=true
      ;;
    *)
      if [[ -z "$VERSION" ]]; then
        VERSION="$arg"
      fi
      ;;
  esac
done

if [[ -z "$VERSION" ]]; then
  echo "Error: version argument is required (e.g. v0.1.0)" >&2
  exit 1
fi

# v プレフィックスを除去
VERSION="${VERSION#v}"

echo "Publishing version: ${VERSION}"
echo "Dry run: ${DRY_RUN}"

# ---- 環境変数チェック ----------------------------------------------------
if [[ "$DRY_RUN" == "false" && -z "${NODE_AUTH_TOKEN:-}" ]]; then
  echo "Error: NODE_AUTH_TOKEN is not set" >&2
  exit 1
fi

# ---- 一時ディレクトリ ----------------------------------------------------
WORK_DIR="$(mktemp -d)"
trap 'rm -rf "$WORK_DIR"' EXIT

echo "Working directory: ${WORK_DIR}"

# ---- メインパッケージの準備 ------------------------------------------------
echo ""
echo "=== Preparing @yone-k/zaim-cli ==="

MAIN_PKG_DIR="${WORK_DIR}/zaim-cli"
cp -r "${PROJECT_ROOT}/npm/zaim-cli" "$MAIN_PKG_DIR"

# package.json の version を更新
jq --arg v "$VERSION" '.version = $v' "${MAIN_PKG_DIR}/package.json" > "${MAIN_PKG_DIR}/package.json.tmp"
mv "${MAIN_PKG_DIR}/package.json.tmp" "${MAIN_PKG_DIR}/package.json"

echo "  version updated to ${VERSION}"

# ---- 公開 ----------------------------------------------------------------
if [[ "$DRY_RUN" == "true" ]]; then
  echo ""
  echo "=== Dry run: showing package contents ==="
  echo ""
  echo "--- package.json ---"
  cat "${MAIN_PKG_DIR}/package.json"
  echo ""
  echo "--- Files ---"
  find "${MAIN_PKG_DIR}" -type f | sort
  echo ""
  echo "Dry run complete. No packages were published."
  exit 0
fi

echo ""
echo "=== Publishing @yone-k/zaim-cli@${VERSION} ==="
cd "$MAIN_PKG_DIR"
if ! output=$(npm publish --access public 2>&1); then
  if echo "$output" | grep -q "You cannot publish over the previously published versions"; then
    echo "  Already published, skipping."
  else
    echo "$output" >&2
    exit 1
  fi
else
  echo "  Published successfully."
fi

echo ""
echo "Package published successfully!"
