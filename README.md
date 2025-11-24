# gh-metric

GitHub リポジトリのメトリクスを収集・分析するための GitHub CLI 拡張機能です。

## 機能

- **プルリクエスト分析**: マージされたPRのリードタイム計測とレビュー効率の分析
- **イシュー分析**: イシューの作成傾向とプロジェクトフィールド値の集計

## インストール

```bash
gh extension install kmtym1998/gh-metric
```

## 使い方

### イシュー分析

```bash
# 基本的な使い方
gh metric issues                             # 現在のリポジトリ
gh metric issues --repo owner/repo           # 特定リポジトリ
gh metric issues --state open --limit 50     # オープンイシューのみ
gh metric issues --since 2024-01-01          # 期間指定
```

主なオプション:
- `--repo`: 対象リポジトリ（デフォルト: 現在のディレクトリ）
- `--state`: all, open, closed（デフォルト: all）
- `--since/--until`: 期間指定（YYYY-MM-DD形式）
- `--limit`: 最大取得数（デフォルト: 100）

### PR分析

```bash
# 基本的な使い方
gh metric merged-pr                          # マージ済みPRを分析
gh metric merged-pr --target-user user1      # 特定ユーザーのPR
gh metric merged-pr --output csv             # CSV出力
gh metric merged-pr --exclude-weekends       # 週末を除外
```

主なオプション:
- `--repo`: 対象リポジトリ
- `--since/--until`: 期間指定（YYYY-MM-DD形式）
- `--output`: json, csv（デフォルト: csv）
- `--exclude-weekends`: リードタイム計算時に週末を除外
- `--target-user`: 特定ユーザーのPRのみ（カンマ区切り）
- `--limit`: 最大取得数（デフォルト: 10）
- `--include-bot`: ボットレビューを含める

#### 取得できる指標

**リードタイム関連**
- 初回コミットからマージまでの時間
- PR作成からマージまでの時間
- 初回レビューまでの時間
- レビュー承認からマージまでの時間

**レビュー関連**
- レビュアー数
- レビューコメント数
- 承認レビュー数
- 変更リクエスト数

**その他**
- 追加/削除行数
- 変更ファイル数
- コミット数
