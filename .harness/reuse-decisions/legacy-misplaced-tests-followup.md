# Reuse Decision

## Change type
+test / cleanup / docs

## Existing code searched
- `code/frontend/src/components/platform/awd-service/__tests__/AWDChallengeEditorDialog.test.ts`
- `code/frontend/src/components/platform/awd-service/__tests__/AWDChallengeLibraryPage.test.ts`
- `code/frontend/src/components/teacher/reports/__tests__/ClassReportExportDialog.test.ts`
- `code/frontend/src/features/platform/awd-challenges/ui/*`
- `code/frontend/src/features/teaching/class-report-export/ui/*`

## Similar implementations found
- `AWDChallengeEditorDialog` 与 `AWDChallengeLibraryPage` 的运行时 owner 已经在 `features/platform/awd-challenges/ui/*`。
- `ClassReportExportDialog` 的运行时 owner 已经在 `features/teaching/class-report-export/ui/*`。
- 这三份测试本身仍然有效，但目录位置属于历史路径残留，不应继续挂在 `components/platform/awd-service/__tests__` 或 `components/teacher/reports/__tests__`。

## Decision
refactor_existing

## Reason
- 这次继续做“测试跟随当前 owner”收口，而不是继续保留旧目录测试壳。
- 最小正确动作是把测试迁到 feature 当前落点附近，不改变覆盖内容。

## Files to modify
- `.harness/reuse-decisions/legacy-misplaced-tests-followup.md`
- `code/frontend/src/components/platform/awd-service/__tests__/AWDChallengeEditorDialog.test.ts`
- `code/frontend/src/components/platform/awd-service/__tests__/AWDChallengeLibraryPage.test.ts`
- `code/frontend/src/components/teacher/reports/__tests__/ClassReportExportDialog.test.ts`
- `code/frontend/src/features/platform/awd-challenges/ui/AWDChallengeEditorDialog.test.ts`
- `code/frontend/src/features/platform/awd-challenges/ui/AWDChallengeLibraryPage.test.ts`
- `code/frontend/src/features/teaching/class-report-export/ui/ClassReportExportDialog.test.ts`

## After implementation
- `components/platform/awd-service/__tests__/` 和 `components/teacher/reports/__tests__/` 不再承载这批业务测试。
- 相关测试只跟随 `features/platform/awd-challenges/ui/*` 与 `features/teaching/class-report-export/ui/*` 当前 owner。
