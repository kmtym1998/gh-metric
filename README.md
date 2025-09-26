# gh-metric

GitHub リポジトリのメトリクスを収集・分析するための GitHub CLI 拡張機能です。

## 概要

`gh-metric` は GitHub API を活用して、Pull Request のリードタイムやその他の開発メトリクスを計算・出力するツールです。現在は **マージされた Pull Request のリードタイム分析** 機能を提供しています。

## 機能

### merged-pr コマンド

マージされた Pull Request について以下のメトリクスを計算します：

- **基本情報**: PR 番号、タイトル、作成者、作成日時、URL
- **リードタイム**:
  - `until_first_review`: レビュワーアサインから最初のレビューまでの時間（時間）
  - `until_first_approve`: レビュワーアサインから最初の Approval までの時間（時間）
  - `until_merge`: レビュワーアサインからマージまでの時間（時間）
- **変更情報**: コメント数、変更ファイル数、追加行数、削除行数
- **レビュー情報**: レビュワー一覧（カンマ区切り）

## インストール

### 前提条件

- Go 1.25.0 以上
- [GitHub CLI (gh)](https://cli.github.com/) がインストールされ、認証が完了していること

### ビルドとインストール

```bash
# リポジトリをクローン
git clone https://github.com/kmtym1998/gh-metric.git
cd gh-metric

# ビルド
go build -o gh-metric .

# PATH が通った場所にコピー（例：~/.local/bin/）
cp gh-metric ~/.local/bin/
```

### GitHub CLI 拡張として使用する場合

GitHub CLI の拡張として使用するには、バイナリを `gh-metric` という名前で PATH の通った場所に配置してください：

```bash
# 例：~/.local/bin/ に配置
cp gh-metric ~/.local/bin/

# gh CLI から実行
gh metric merged-pr --help
```

## 使い方

### 基本的な使用法

```bash
# 現在のディレクトリのリポジトリを分析（2024年1月1日以降）
./gh-metric merged-pr --since 2024-01-01

# 特定のリポジトリを分析
./gh-metric merged-pr microsoft/vscode --since 2024-01-01

# 日付範囲を指定
./gh-metric merged-pr spf13/cobra --since 2024-01-01 --until 2024-12-31
```

### オプション

| オプション              | 説明                                 | 例                        |
| ----------------------- | ------------------------------------ | ------------------------- |
| `--since <YYYY-MM-DD>`  | 指定日以降にマージされた PR のみ対象 | `--since 2024-01-01`      |
| `--until <YYYY-MM-DD>`  | 指定日以前にマージされた PR のみ対象 | `--until 2024-12-31`      |
| `--output <json\|csv>`  | 出力形式（デフォルト：csv）          | `--output json`           |
| `--exclude-weekends`    | 週末を除外してリードタイムを計算     | `--exclude-weekends`      |
| `--target-user <users>` | 特定ユーザーが作成した PR のみ対象   | `--target-user john,jane` |

### 使用例

#### CSV 出力（デフォルト）

```bash
./gh-metric merged-pr spf13/cobra --since 2024-01-01 --output csv
```

出力例：

```csv
number,title,author,created_at,url,until_first_review,until_first_approve,until_merge,comment_count,changed_files,additions,deletions,reviewers
2305,chore: upgrade pflags v1.0.9,jpmcb,2025-09-01T12:25:08Z,https://github.com/spf13/cobra/pull/2305,3.33,3.64,3.83,4,2,3,3,marckhouzam
```

#### JSON 出力

```bash
./gh-metric merged-pr spf13/cobra --since 2025-09-01 --output json
```

出力例：

```json
[
  {
    "number": 2305,
    "title": "chore: upgrade pflags v1.0.9",
    "author": "jpmcb",
    "createdAt": "2025-09-01T12:25:08Z",
    "url": "https://github.com/spf13/cobra/pull/2305",
    "until_first_review": 3.33,
    "until_first_approve": 3.64,
    "until_merge": 3.83,
    "comment_count": 4,
    "changed_files": 2,
    "additions": 3,
    "deletions": 3,
    "reviewers": "marckhouzam"
  }
]
```

#### 特定ユーザーの PR のみ分析

```bash
./gh-metric merged-pr --target-user alice,bob --since 2024-01-01
```

#### 週末を除外してリードタイム計算

```bash
./gh-metric merged-pr --exclude-weekends --since 2024-01-01
```

## 開発者向け情報

### プロジェクト構造

```
.
├── cmd/                    # CLI コマンド定義
│   ├── root.go            # ルートコマンド
│   └── mergedpr.go        # merged-pr サブコマンド
├── internal/
│   ├── github/            # GitHub API 関連
│   │   ├── client.go      # GraphQL クライアント
│   │   ├── types.go       # データ型定義
│   │   ├── query.go       # GraphQL クエリ
│   │   ├── service.go     # サービス層
│   │   ├── metrics.go     # メトリクス計算
│   │   └── repo.go        # リポジトリ検出
│   └── output/            # 出力フォーマット
│       └── formatter.go   # CSV/JSON フォーマッター
├── main.go                # エントリーポイント
└── README.md
```

### 開発環境のセットアップ

```bash
# 依存関係の取得
go mod tidy

# ビルド
go build -o gh-metric .

# フォーマット
go fmt ./...

# Vet チェック
go vet ./...
```

### デバッグ方法

#### ログレベルの設定

アプリケーションは構造化ログ（slog）を使用しています。デバッグ情報を確認する場合：

```bash
# ログレベルをDEBUGに設定（環境変数で制御可能にする場合）
SLOG_LEVEL=DEBUG ./gh-metric merged-pr --since 2024-01-01
```

#### GraphQL クエリのデバッグ

実行される GraphQL クエリは INFO レベルでログ出力されます：

```
time=2025-09-26T19:00:32.832+09:00 level=INFO msg="Executing search query" query="is:merged repo:spf13/cobra merged:>=2024-12-01"
```

#### API レスポンスのデバッグ

GitHub API のレスポンス詳細をデバッグしたい場合は、`internal/github/client.go` の `ExecuteQuery` メソッドでログレベルを DEBUG に設定してください。

#### 一般的なトラブルシューティング

**認証エラーが発生する場合：**

```bash
# gh CLI の認証状態を確認
gh auth status

# 必要に応じて再認証
gh auth login
```

**リポジトリが検出されない場合：**

```bash
# 現在のディレクトリの git remote を確認
git remote -v

# origin リモートが正しく設定されていることを確認
```

**GraphQL API エラーが発生する場合：**

- レート制限に引っかかっている可能性があります
- 一時的にリクエスト頻度を下げるか、時間を空けて再実行してください

### テスト

現在、単体テストは実装されていませんが、コードは以下の方針でテストしやすい設計になっています：

- HTTP クライアントは抽象化されており、モックに置き換え可能
- メトリクス計算ロジックは純粋関数として実装
- 各パッケージは適切に分離されている

### 貢献方法

1. このリポジトリをフォーク
2. フィーチャーブランチを作成 (`git checkout -b feature/amazing-feature`)
3. 変更をコミット (`git commit -m 'Add amazing feature'`)
4. ブランチにプッシュ (`git push origin feature/amazing-feature`)
5. Pull Request を作成

## 技術仕様

- **言語**: Go 1.25.0
- **CLI フレームワーク**: [spf13/cobra](https://github.com/spf13/cobra)
- **GitHub API**: GraphQL API v4
- **認証**: GitHub CLI (gh) の認証を利用
- **ログ**: Go 標準ライブラリの `log/slog`

## ライセンス

このプロジェクトのライセンス情報については、LICENSE ファイルを参照してください。
