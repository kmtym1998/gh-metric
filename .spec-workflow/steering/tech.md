# Technology Stack

## ドキュメント最終更新日時

- `2025-11-20 00:03:43`
- `6ed2490`

## プロジェクトの概要

gh-metricは、GitHub CLIの拡張機能として提供されるコマンドラインツールです。GitHubリポジトリの開発メトリクスを収集・分析し、チームや個人の生産性向上に貢献します。

## 技術スタック

### 言語・ランタイム

- **Language**: Go 1.25.0
- **Runtime/Compiler**: Go標準コンパイラ (darwin/arm64)
- **Language-specific tools**: go mod (モジュール管理)、go fmt (コードフォーマット)

### 主要な依存関係・ライブラリ

- **github.com/cli/go-gh/v2**: GitHub CLI統合のためのライブラリ
- **github.com/spf13/cobra**: CLIフレームワーク、コマンド構造の実装
- **github.com/go-git/go-git/v5**: Gitリポジトリ操作

### アーキテクチャパターンの概要

Clean Architectureに基づいた層構造を採用しています（将来的な変更の可能性あり）。詳細な構造についてはstructure.mdを参照してください。

### データストレージ

- **Primary storage**: なし（ステートレスなCLIツール）
- **Caching**: なし（毎回GitHub APIから最新データを取得）

### 外部連携

- **APIs**: GitHub GraphQL/REST API（GitHub CLIを経由）
- **Protocols**: HTTPS
- **Authentication**: GitHub Personal Access Token（GitHub CLIが管理）

### モニタリング、オブザーバビリティ

現在は該当なし（CLIツールのため）

## 開発環境とツール

### ビルド・開発ツール

- **Build System**: go build（Go標準ビルドツール）
- **Task Runner**: Make（開発タスクの実行）
- **Package Management**: go mod（Go標準のモジュール管理）
- **Development workflow**: GitHub Actionsによる自動リリース（タグ付け時）

### Code Quality Tools

- **Static Analysis**: なし（go vetを必要に応じて実行）
- **Formatting**: go fmt（Go標準フォーマッター）
- **Testing Framework**: testing（Go標準テストパッケージ）
- **Documentation**: GoDoc形式のコメント

## デプロイ・ディストリビューション

- **Target Platform(s)**: macOS、Linux、Windows（GitHub CLIがサポートする全プラットフォーム）
- **Distribution Method**: GitHub CLI拡張機能（`gh extension install`コマンド）
- **Installation Requirements**: GitHub CLI 2.0以降がインストール済みであること
- **Update Mechanism**: `gh extension upgrade`コマンドによる更新

## Technical Requirements & Constraints

### Performance Requirements

- GitHub APIのレート制限とレスポンス速度に依存
- 大規模リポジトリでも動作可能だが、API制限の範囲内での動作
- 特定のパフォーマンス目標は設定しない（GitHub API依存のため）

### Compatibility Requirements

- **Platform Support**: GitHub CLIがサポートする全OS・アーキテクチャ
- **Dependency Versions**: Go 1.25.0以降、GitHub CLI 2.0以降
- **Standards Compliance**: GitHub API v4（GraphQL）仕様に準拠

### Security & Compliance

- **Security Requirements**:
  - Personal Access Tokenは最小権限で運用することを推奨
  - 認証情報の管理はGitHub CLIに委譲（拡張機能では管理しない）
- **Compliance Standards**: 特になし
- **Threat Model**: トークンの漏洩防止（ユーザー責任）

## 技術的な制約事項

- **GitHub APIレート制限**: 大量のPR分析時にレート制限に到達する可能性
  - 将来的にはバッチ処理や効率的なクエリで改善予定

- **リアルタイム性の欠如**: 毎回APIを呼び出すためリアルタイムではない
  - 設計上の選択（ステートレス）のため、現時点では対応予定なし

- **テストカバレッジ**: 現在テストコードが未実装
  - 主要機能の安定化後、段階的にテストを追加予定