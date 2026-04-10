"use strict";

const https = require("https");
const http = require("http");
const fs = require("fs");
const path = require("path");
const { execSync } = require("child_process");

const ARCH_MAP = {
  x64: "amd64",
  arm64: "arm64",
};

const PLATFORM_MAP = {
  darwin: "darwin",
  linux: "linux",
};

function getArchiveName(version) {
  const platform = PLATFORM_MAP[process.platform];
  const arch = ARCH_MAP[process.arch];

  if (!platform || !arch) {
    return null;
  }

  return `zaim-cli_${version}_${platform}_${arch}.tar.gz`;
}

function getDownloadUrl(version, archiveName) {
  return `https://github.com/yone-k/zaim-cli/releases/download/v${version}/${archiveName}`;
}

function download(url) {
  return new Promise((resolve, reject) => {
    const get = url.startsWith("https") ? https.get : http.get;
    get(url, (res) => {
      if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
        download(res.headers.location).then(resolve, reject);
        return;
      }
      if (res.statusCode !== 200) {
        reject(new Error(`ダウンロード失敗: HTTP ${res.statusCode} - ${url}`));
        return;
      }
      const chunks = [];
      res.on("data", (chunk) => chunks.push(chunk));
      res.on("end", () => resolve(Buffer.concat(chunks)));
      res.on("error", reject);
    }).on("error", reject);
  });
}

async function main() {
  const pkgJson = JSON.parse(
    fs.readFileSync(path.join(__dirname, "package.json"), "utf8")
  );
  const version = pkgJson.version;

  const archiveName = getArchiveName(version);
  if (!archiveName) {
    console.warn(
      `警告: zaim-cli はこのプラットフォーム (${process.platform}-${process.arch}) をサポートしていません。\n` +
      `サポート対象: darwin-arm64, darwin-x64, linux-x64, linux-arm64`
    );
    process.exit(0);
  }

  const url = getDownloadUrl(version, archiveName);
  console.log(`zaim-cli v${version} をダウンロード中...`);
  console.log(`  URL: ${url}`);

  const data = await download(url);

  // tar.gz を一時ファイルに書き出し
  const tmpFile = path.join(__dirname, ".download.tar.gz");
  fs.writeFileSync(tmpFile, data);

  // バイナリをパッケージルート直下に展開
  const binaryDest = path.join(__dirname, "zaim-cli");
  try {
    execSync(`tar xzf "${tmpFile}" -C "${__dirname}" zaim-cli`, {
      stdio: "pipe",
    });
    fs.chmodSync(binaryDest, 0o755);
    console.log(`  バイナリを配置しました: ${binaryDest}`);
  } finally {
    // 一時ファイルを削除
    try {
      fs.unlinkSync(tmpFile);
    } catch (_) {}
  }
}

main().catch((err) => {
  console.error(`警告: バイナリのダウンロードに失敗しました: ${err.message}`);
  console.error(
    "手動でバイナリをダウンロードするか、ZAIM_BINARY_PATH 環境変数でパスを指定してください。"
  );
  process.exit(0);
});
