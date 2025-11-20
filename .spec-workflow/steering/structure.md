# Project Structure

## ディレクトリ構成

```
gh-metric/
├── cmd/                    # CLIコマンド定義
│   ├── merged_pr.go       # merged-prコマンドの実装
│   └── root.go            # ルートコマンドとグローバル設定
├── internal/              # 内部パッケージ（外部から参照不可）
│   ├── github/           # GitHub API関連
│   │   ├── client.go     # GraphQLクライアントの初期化
│   │   ├── query.go      # GraphQLクエリ定義
│   │   ├── repo.go       # リポジトリ情報の取得
│   │   ├── service.go    # GitHub APIサービスのインターフェース
│   │   └── types.go      # GraphQLレスポンスの型定義
│   ├── output/           # 出力フォーマッター
│   │   ├── formatter.go           # フォーマッターインターフェース
│   │   ├── json_formatter.go      # JSON形式での出力
│   │   └── pr_csv_formatter.go    # CSV形式での出力
│   └── usecase/          # ビジネスロジック
│       ├── aggregate_merged_pr.go # PRメトリクス集計ロジック
│       └── utils.go               # 共通ユーティリティ関数
├── main.go                # エントリーポイント
├── go.mod                 # Goモジュール定義
├── go.sum                 # 依存関係のチェックサム
├── Makefile              # 開発タスクの自動化
└── README.md             # プロジェクトドキュメント
```

レイヤー構造（Clean Architecture）：

- **cmd 層**: ユーザー入力の処理とコマンド実行
- **usecase 層**: ビジネスロジックとデータ処理
- **internal/github 層**: 外部 API 通信の詳細を隠蔽
- **internal/output 層**: 出力形式の変換

## 命名規則

### ファイル名・ディレクトリ名

- **パッケージ/モジュール**: `snake_case`（例: `merged_pr.go`, `json_formatter.go`）
- **サービス/ハンドラー**: `snake_case`（例: `service.go`, `client.go`）
- **ユーティリティ/ヘルパー**: `snake_case`（例: `utils.go`, `formatter.go`）
- **テスト**: `[filename]_test.go`（例: `utils_test.go`）※現在未実装

### コード内

- **構造体/型**: `PascalCase`（例: `PRMetric`, `MergedPRInput`）
- **関数/メソッド**:
  - 公開: `PascalCase`（例: `NewService`, `Execute`）
  - 非公開: `camelCase`（例: `validateFlags`, `parseTargetUsers`）
- **定数**: `camelCase`または`PascalCase`
- **変数**: `camelCase`（例: `excludeWeekends`, `targetUsers`）

## Code Organization Principles

1. **Single Responsibility**: 各パッケージは単一の責任を持つ
2. **Modularity**: 再利用可能なモジュール構造
3. **Testability**: テスタブルな設計（インターフェース活用）
4. **Consistency**: 確立されたパターンの一貫した適用

## Module Boundaries

### 依存関係の方向

```
main.go
  ↓
cmd/
  ↓
internal/usecase/ ← internal/output/
  ↓
internal/github/
```

### 責任の分離

- **Public API vs Internal**: `internal/`配下は外部から利用不可
- **Dependencies direction**: 下位層は上位層に依存しない
- **Core vs Extensions**: コア機能（usecase）と拡張機能（output）の分離

## Code Size Guidelines

推奨ガイドライン：

- **File size**: 300 行以下
- **Function/Method size**: 70 行以下
- **Class/Module complexity**: Cyclomatic の複雑度 20 以下
- **Nesting depth**: 最大 5 レベル

## Documentation Standards

- すべての公開関数/型に GoDoc コメントを記載
- 複雑なロジックにはインラインコメントを追加
- 各パッケージに package doc コメントを記載
- README に使用方法とインストール手順を記載
