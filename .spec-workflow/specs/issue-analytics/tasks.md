# タスク一覧

イシュー集計機能の実装タスク。各タスク完了後は`go build`が成功することを確認する。

- [x] 1. GraphQL型定義の実装

  - File: internal/github/types.go
  - listIssueAndProjectFieldsQueryのレスポンス型を追加
  - プライベート型（小文字）として実装
  - Purpose: GraphQLレスポンスのマッピング構造を定義
  - _Leverage: 既存のSearchResponse, PullRequest型の実装パターン_
  - _Requirements: 要求2（イシュー詳細情報の取得）_
  - _Prompt: Implement the task for spec issue-analytics, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in GraphQL API integration | Task: Add listIssuesResponse and issueNode types to internal/github/types.go following the structure defined in listIssueAndProjectFieldsQuery, using private types (lowercase) for internal use only | Restrictions: Do not modify existing types, maintain JSON tag consistency, follow existing code patterns | Success: Types compile without errors, JSON tags match GraphQL field names exactly, supports all fields from the query. After completion, edit tasks.md to mark this task as complete [-] to [x], use log-implementation tool to record the implementation details._

- [x] 2. UseCase層の型定義

  - File: internal/usecase/aggregate_issues.go (新規作成)
  - AggregateIssuesInput, AggregateIssuesOutput, IssueMetric型を定義
  - ProjectFieldValue, Period, Repository型を含む
  - Purpose: ビジネスロジック層のデータ構造を確立
  - _Leverage: internal/usecase/aggregate_merged_pr.goのパターン_
  - _Requirements: 要求1,2,3_
  - _Prompt: Implement the task for spec issue-analytics, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Backend Developer with expertise in Clean Architecture | Task: Create internal/usecase/aggregate_issues.go with Input/Output types and IssueMetric structure including week/month calculations and project field values, following patterns from aggregate_merged_pr.go | Restrictions: Maintain consistent naming with existing usecase types, include proper JSON tags, follow Clean Architecture principles | Success: All types are properly defined with JSON tags, follows existing usecase patterns, supports all required fields from requirements. After completion, edit tasks.md to mark this task as complete [-] to [x], use log-implementation tool to record the implementation details._

- [x] 3. GraphQLサービス拡張

  - File: internal/github/service.go
  - FetchIssuesメソッドを追加
  - listIssueAndProjectFieldsQueryを使用した実装
  - Purpose: GitHub APIとの通信層を実装
  - _Leverage: 既存のFetchPullRequestsメソッドのパターン_
  - _Requirements: 要求1,2_
  - _Prompt: Implement the task for spec issue-analytics, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with GraphQL expertise | Task: Add FetchIssues method to internal/github/service.go using listIssueAndProjectFieldsQuery, implementing pagination and error handling following the pattern of existing FetchPullRequests method | Restrictions: Must handle pagination correctly, maintain error handling consistency, use existing GraphQL client | Success: Method fetches issues with project fields, handles pagination properly, returns typed responses correctly. After completion, edit tasks.md to mark this task as complete [-] to [x], use log-implementation tool to record the implementation details._

- [x] 4. UseCase実装

  - File: internal/usecase/aggregate_issues.go (Task 2で作成済み)
  - Executeメソッドの実装
  - Issue型からIssueMetric型への変換ロジック
  - Purpose: ビジネスロジックの実装
  - _Leverage: internal/usecase/aggregate_merged_pr.goのExecuteメソッド, internal/usecase/utils.goの週・月計算関数_
  - _Requirements: 要求1,2,3_
  - _Prompt: Implement the task for spec issue-analytics, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Backend Developer with business logic expertise | Task: Implement Execute method in aggregate_issues.go to fetch issues via GitHub service, transform to IssueMetric with week/month calculations and project field value extraction, following patterns from aggregate_merged_pr.go | Restrictions: Must reuse existing utility functions for date calculations, handle nil values properly, maintain error handling consistency | Success: Execute method processes issues correctly, calculates week/month values, extracts project fields, handles all edge cases. After completion, edit tasks.md to mark this task as complete [-] to [x], use log-implementation tool to record the implementation details._

- [x] 5. CLIコマンド実装

  - File: cmd/issues.go (新規作成)
  - Cobraコマンドの定義と実装
  - フラグの定義（owner, repo, since, until, state, limit）
  - Purpose: ユーザーインターフェースの実装
  - _Leverage: cmd/merged_pr.goのパターン, cmd/root.goへの登録_
  - _Requirements: 要求1,3,4_
  - _Prompt: Implement the task for spec issue-analytics, first run spec-workflow-guide to get the workflow guide then implement the task: Role: CLI Developer with Cobra framework expertise | Task: Create cmd/issues.go with Cobra command implementation including flags (owner, repo, since, until, state, limit), validation, and usecase invocation following patterns from cmd/merged_pr.go | Restrictions: Must validate date formats, handle default values correctly (state=all, limit=100), support current repo detection | Success: Command parses arguments correctly, validates inputs, invokes usecase properly, handles errors gracefully. After completion, edit tasks.md to mark this task as complete [-] to [x], use log-implementation tool to record the implementation details._

- [x] 6. ルートコマンドへの登録

  - File: cmd/root.go
  - issuesコマンドをサブコマンドとして追加
  - Purpose: コマンドをCLIツールに統合
  - _Leverage: 既存のmerged-prコマンドの登録パターン_
  - _Requirements: 要求4_
  - _Prompt: Implement the task for spec issue-analytics, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Register issues command in cmd/root.go as a subcommand following the existing pattern for merged-pr command | Restrictions: Must maintain command initialization order, do not break existing commands | Success: Issues command is accessible via 'gh metric issues', help text displays correctly. After completion, edit tasks.md to mark this task as complete [-] to [x], use log-implementation tool to record the implementation details._

- [x] 7. プロジェクトフィールド値の処理実装

  - File: internal/usecase/aggregate_issues.go
  - extractProjectFieldValuesヘルパー関数の実装
  - 各フィールドタイプ（text, number, single_select等）の処理
  - Purpose: プロジェクトフィールド値の抽出ロジック
  - _Leverage: GraphQL型定義の__typename判定_
  - _Requirements: 要求2_
  - _Prompt: Implement the task for spec issue-analytics, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with GraphQL and type assertion expertise | Task: Implement extractProjectFieldValues helper function in aggregate_issues.go to parse different field types (text, number, single_select, date, iteration, etc.) using type assertions on interface{} from GraphQL response | Restrictions: Must handle all field types from the query, safely handle nil values and type assertions, maintain clear error messages | Success: All project field types are correctly extracted and converted to ProjectFieldValue structs, handles missing or null fields gracefully. After completion, edit tasks.md to mark this task as complete [-] to [x], use log-implementation tool to record the implementation details._

- [x] 8. JSON出力フォーマッターの再利用確認

  - File: internal/output/json_formatter.go
  - AggregateIssuesOutputが正しくJSON出力されることを確認
  - 必要に応じて微調整
  - Purpose: 出力形式の統一性を確保
  - _Leverage: 既存のJSONFormatterの汎用実装_
  - _Requirements: 要求3_
  - _Prompt: Implement the task for spec issue-analytics, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Verify that existing JSONFormatter in internal/output/json_formatter.go correctly handles AggregateIssuesOutput structure, make minor adjustments if needed for proper JSON serialization | Restrictions: Do not break existing formatter functionality, maintain backward compatibility | Success: Issues data is properly formatted as JSON with correct field names and structure. After completion, edit tasks.md to mark this task as complete [-] to [x], use log-implementation tool to record the implementation details._

- [x] 9. エラーハンドリングの実装

  - File: cmd/issues.go, internal/usecase/aggregate_issues.go
  - 各種エラーシナリオの処理
  - 適切なエラーメッセージの実装
  - Purpose: 堅牢性とユーザビリティの向上
  - _Leverage: 既存のエラーハンドリングパターン_
  - _Requirements: 非機能要求（信頼性）_
  - _Prompt: Implement the task for spec issue-analytics, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with error handling expertise | Task: Implement comprehensive error handling in cmd/issues.go and aggregate_issues.go for scenarios like invalid dates, API errors, rate limits, network issues following existing error patterns | Restrictions: Must provide clear user-facing error messages, maintain consistent error handling style, use appropriate exit codes | Success: All error scenarios are properly handled with informative messages, graceful degradation where appropriate. After completion, edit tasks.md to mark this task as complete [-] to [x], use log-implementation tool to record the implementation details._

- [ ] 10. 統合テスト

  - File: cmd/issues_test.go (新規作成)
  - コマンドライン引数のパースとバリデーションのテスト
  - モックを使用したエンドツーエンドのフロー確認
  - Purpose: 機能の統合動作を検証
  - _Leverage: cmd/merged_pr_test.goのテストパターン（存在する場合）_
  - _Requirements: 全要求_
  - _Prompt: Implement the task for spec issue-analytics, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer with Go testing expertise | Task: Create integration tests in cmd/issues_test.go covering command parsing, validation, and end-to-end flow with mocked GitHub service, following existing test patterns | Restrictions: Must test both success and error cases, maintain test isolation, use appropriate mocking strategies | Success: Tests cover major use cases and error scenarios, run reliably and quickly, provide good coverage of the feature. After completion, edit tasks.md to mark this task as complete [-] to [x], use log-implementation tool to record the implementation details._

- [ ] 11. 手動テストと最終確認

  - 実際のGitHubリポジトリでの動作確認
  - 各種パラメータの組み合わせテスト
  - ドキュメントの更新（README.mdへのコマンド追加）
  - Purpose: 実環境での動作確認と完成度の向上
  - _Leverage: 既存のREADME.mdの構造_
  - _Requirements: 全要求_
  - _Prompt: Implement the task for spec issue-analytics, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Senior Developer with documentation skills | Task: Perform manual testing with real GitHub repositories, test various parameter combinations, update README.md with new issues command documentation following existing documentation style | Restrictions: Must test against repos with project fields, verify all edge cases, maintain documentation consistency | Success: Feature works correctly in production environment, all use cases are validated, documentation is complete and clear. After completion, edit tasks.md to mark this task as complete [-] to [x], use log-implementation tool to record the implementation details._