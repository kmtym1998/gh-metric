# 指示

## 概要

これは gh-cli の拡張機能です。Go で実装します。
GitHub API を利用して、GitHub のリポジトリに関する様々なメトリクスを収集・出力します。

## 機能

gh metric コマンドを提供します。このコマンドは以下のサブコマンドを持ちます。

- `merged-pr`: マージされた Pull Request のリードタイムを計算します。

### merged-pr

Pull Request に関連する以下の指標を計算します。

- PR 番号
- PR タイトル
- PR 作成者の GitHub ユーザー名
- PR 作成日時
- URL
- until_first_review: Pull Request にレビュワーがアサインされてから最初のレビューまでの時間 (hours)
- until_first_approve: Pull Request にレビュワーがアサインされてから最初の approve までの時間 (hours)
- until_merge: Pull Request にレビュワーがアサインされてからマージされるまでの時間 (hours)
- PR のコメント数
- 変更ファイル数
- 追加行数
- 削除行数
- レビュワーの一覧 (カンマ区切りで id を出力)

以下のオプションをサポートします。

- `--since <YYYY-MM-DD>`: 指定した日付以降にマージされた Pull Request のみを対象とします。日付の形式は `YYYY-MM-DD` です。
- `--until <YYYY-MM-DD>`: 指定した日付以前にマージされた Pull Request のみを対象とします。日付の形式は `YYYY-MM-DD` です。
- `--output <json|csv>`: 出力形式を指定します。`json` または `csv` を指定できます。デフォルトは `csv` です。標準出力に対して出力します。
- `--exclude-weekends`: 週末を除外してリードタイムを計算します。
- `--target-user <username1,username2,...>`: 指定したユーザーが作成した Pull Request のみを対象とします。カンマ区切りで複数のユーザーを指定できます。

## 実装

### GitHub GraphQL API クエリ

PR 指標の取得には GitHub GraphQL API を使用します。以下のクエリを使用します。ページネーションに対応するため、`$cursor` 変数を使用しています。また、`$query` 変数には検索クエリを渡します。オプションに指定された日付やユーザー名を組み合わせてクエリを生成します。マージされた Pull Request のみを対象とするため、`is:merged` フィルターを使用します。途中で API からエラーレスポンスが返ってきた場合は、エラーレスポンスを出力し、処理を中断します。

```gql
query PullRequestLeadTime($query: String!, $cursor: String) {
  search(query: $query, type: ISSUE, first: 50, after: $cursor) {
    pageInfo {
      startCursor
      endCursor
      hasNextPage
    }
    nodes {
      ... on PullRequest {
        number
        title
        createdAt
        mergedAt
        closedAt
        url
        additions
        deletions
        changedFiles
        comments {
          totalCount
        }
        timelineItems(first: 30, itemTypes: [REVIEW_REQUESTED_EVENT]) {
          pageInfo {
            startCursor
            endCursor
            hasNextPage
          }
          nodes {
            __typename
            ... on ReviewRequestedEvent {
              __typename
              createdAt
              requestedReviewer {
                ... on Bot {
                  __typename
                  login
                }
                ... on Mannequin {
                  __typename
                  login
                }
                ... on Team {
                  __typename
                  name
                  members {
                    nodes {
                      login
                    }
                  }
                }
                ... on User {
                  __typename
                  login
                }
              }
            }
          }
        }
        author {
          login
        }
        reviews(first: 50) {
          nodes {
            author {
              login
            }
            state
            submittedAt
          }
        }
      }
    }
  }
}
```

### リードタイム計算

リードタイムの計算の起点となる時刻は以下の通りです。

- Pull Request にレビュワーがアサインされた時刻:
  - `timelineItems` の中で最も古い `createdAt` を使用します。
- 最初のレビューが行われた時刻:
  - `reviews` の中で最も古い `submittedAt` を使用します。
- 最初の approve が行われた時刻:
  - `reviews` の中で最も古い `submittedAt` で、かつ `state` が `APPROVED` のものを使用します。
- Pull Request がマージされた時刻:
  - `mergedAt` を使用します。

リードタイムはすべて単位は (hours) とします。いちど unixtime に変換してから計算し、最後に時間に変換します。

### 使用するツール

- 実装言語は Go (1.25.0)
- http リクエストには標準の net/http パッケージを使用します
- ロギングは log/slog パッケージを使用してください
- CLI の実装には spf13/cobra パッケージを使用してください

### コーディングスタイル

- テストは書かなくてもいいですが、後からテストが書きやすいように HTTP リクエストのメソッドなどをモックしたり、httptest に置き換えやすいような設計をしてください
- Go のコーディングスタイルに従ってください
- ファイル、関数、メソッドなどは適切に分割してください。

### その他

すべてを一気に実装するのではなく、すこしずつ進めて私に都度方針を確認しながら進めてください。
