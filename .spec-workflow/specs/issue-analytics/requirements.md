# 要求仕様書

- 最終更新日: `2025-11-22`
- バージョン: `8b307c5`

## 概要

イシュー集計機能は、GitHubリポジトリのイシューデータを収集・分析し、指定期間内のイシュー活動を定量的に可視化する機能です。チームや個人がイシューの解決効率やボトルネックを把握し、プロジェクトの健全性を評価できるようにします。

## プロダクトビジョンとの整合性

この機能は、product.mdで定義された「Issue分析」機能の実装であり、以下の目標を達成します：

- **データアクセシビリティ**: GitHub CLIの生データを振り返りに即座に使える形式で提供
- **分析の網羅性**: PR分析に加えてIssue集計機能を追加し、開発活動全体の可視化を実現
- **実用的な出力**: Issueの解決時間、ラベル別の傾向など、チームの振り返りに直接活用できる情報を提供

## 機能要求

### 要求 1: イシューのクローズ数集計

**User Story:** エンジニアリングマネージャーは、指定した期間内にクローズされたイシューの数を把握し、チームの問題解決効率を評価する

#### 受け入れ条件

- コマンドのオプションに開始日と終了日が指定されたとき、その期間内にクローズされたイシューの総数を集計する
- リポジトリの所有者とリポジトリ名を指定してイシューを取得できる
- クローズ日時でフィルタリングし、指定期間内のイシューのみをカウントする
- 結果にはイシューの総数と期間の情報を含める

### 要求 2: イシュー詳細情報の取得

**User Story:** 開発チームリーダーは、クローズされたイシューの詳細情報を確認し、問題解決のパターンを分析する

#### 受け入れ条件

- 各イシューについて、番号、タイトル、作成日時、更新日時、クローズ日時を取得する
- イシューの作成者とアサイン担当者の情報を含める
- ラベル情報を取得し、カテゴリー別の分析を可能にする
- マイルストーン情報がある場合は含める
- プロジェクトフィールドの値をすべて取得する
- 作成日時から週・月、クローズ日時から週・月を計算して含める
- JSON形式の出力フィールド:
  ```json
  {
    "period": { "start": "2025-10-01", "end": "2025-10-31" },
    "repository": { "owner": "myorg", "name": "myrepo" },
    "total_count": 47,
    "issues": [
      {
        "number": 123,
        "title": "バグ修正",
        "state": "CLOSED",
        "created_at": "2025-10-01T10:00:00Z",
        "created_week": "2025-W40",
        "created_month": "2025-10",
        "updated_at": "2025-10-04T12:00:00Z",
        "closed_at": "2025-10-05T14:30:00Z",
        "closed_week": "2025-W40",
        "closed_month": "2025-10",
        "author": "user1",
        "assignees": ["user2"],
        "labels": ["bug", "priority-high"],
        "milestone": "v1.0.0",
        "project_field_values": [
          {
            "field_name": "Status",
            "field_type": "single_select",
            "value": "Done"
          },
          {
            "field_name": "Priority",
            "field_type": "single_select",
            "value": "High"
          },
          {
            "field_name": "Sprint",
            "field_type": "iteration",
            "value": "Sprint 23"
          }
        ]
      }
    ]
  }
  ```

### 要求 3: JSON出力形式サポート

**User Story:** チームメンバーは、集計結果をJSON形式で出力し、プログラマティックに処理する

#### 受け入れ条件

- JSON形式での出力をサポートし、プログラマティックな処理を可能にする
- イシューはリスト形式で出力され、各イシューの詳細情報を含む
- 統計情報の集計は出力側（受け取り側のツール）で行うことを前提とする

## 非機能要求

### パフォーマンス

- 100件のイシューを5秒以内に処理できること
- ページネーションを適切に処理し、大量のイシュー（1000件以上）も取得可能であること
- GitHub APIのレート制限を考慮し、適切な待機処理を実装すること

### セキュリティ

- GitHub CLIの認証機構を使用し、拡張機能独自でトークンを管理しないこと
- APIアクセスには必要最小限の権限（repo:read）のみを要求すること

### 信頼性

- ネットワークエラー時には適切なエラーメッセージを表示すること
- データ取得失敗時は処理全体をエラーとして終了すること
- APIレート制限に達した場合、明確なメッセージとリトライ方法を提示すること

### ユーザビリティ

- コマンドの使用方法を`--help`オプションで確認できること
- 進捗状況を表示し、大量データ処理時でもユーザーが状況を把握できること
- エラーメッセージは具体的で、問題解決のための次のアクションを示すこと

## ユーザー利用のイメージ

### コマンドオプション

```
gh metric issues [flags]

Flags:
  -o, --owner string           リポジトリの所有者（組織名またはユーザー名）
                               指定しない場合は現在のリポジトリを使用
  -r, --repo string           リポジトリ名
                               指定しない場合は現在のリポジトリを使用
  -s, --since string          集計開始日（YYYY-MM-DD形式）
  -u, --until string          集計終了日（YYYY-MM-DD形式）
      --state string          イシューの状態（closed, open, all）（デフォルト: all）
      --limit int             取得するイシューの最大数（デフォルト: 100）
  -h, --help                  このコマンドのヘルプを表示

Examples:
  # 基本的な使用方法（先月のイシューを集計）
  gh metric issues --since 2025-10-01 --until 2025-10-31

  # 特定のリポジトリを指定
  gh metric issues --owner myorg --repo myrepo --since 2025-10-01 --until 2025-10-31

  # クローズ済みのイシューのみ集計
  gh metric issues --since 2025-10-01 --until 2025-10-31 --state closed

  # オープン中のイシューを集計
  gh metric issues --since 2025-10-01 --until 2025-10-31 --state open

  # 最大200件まで取得
  gh metric issues --since 2025-10-01 --until 2025-10-31 --limit 200

  # 結果をファイルに保存
  gh metric issues --since 2025-10-01 --until 2025-10-31 > issues.json
```

### 出力例

JSON形式：
```json
{
  "period": {
    "start": "2025-10-01",
    "end": "2025-10-31"
  },
  "repository": {
    "owner": "myorg",
    "name": "myrepo"
  },
  "total_count": 47,
  "issues": [
    {
      "number": 123,
      "title": "Fix authentication bug in login flow",
      "state": "CLOSED",
      "created_at": "2025-10-01T10:00:00Z",
      "created_week": "2025-W40",
      "created_month": "2025-10",
      "updated_at": "2025-10-04T12:00:00Z",
      "closed_at": "2025-10-05T14:30:00Z",
      "closed_week": "2025-W40",
      "closed_month": "2025-10",
      "author": "user1",
      "assignees": ["user2"],
      "labels": ["bug", "priority-high"],
      "milestone": "v1.0.0",
      "project_field_values": [
        {
          "field_name": "Status",
          "field_type": "single_select",
          "value": "Done"
        },
        {
          "field_name": "Priority",
          "field_type": "single_select",
          "value": "High"
        }
      ]
    },
    {
      "number": 124,
      "title": "Add user profile page",
      "state": "OPEN",
      "created_at": "2025-10-02T09:00:00Z",
      "created_week": "2025-W40",
      "created_month": "2025-10",
      "updated_at": "2025-10-15T16:00:00Z",
      "closed_at": null,
      "closed_week": null,
      "closed_month": null,
      "author": "user2",
      "assignees": ["user1", "user3"],
      "labels": ["enhancement", "frontend"],
      "milestone": "v1.1.0",
      "project_field_values": [
        {
          "field_name": "Status",
          "field_type": "single_select",
          "value": "In Progress"
        }
      ]
    }
  ]
}
```