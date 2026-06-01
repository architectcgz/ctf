// ── 兼容 barrel（DEPRECATED）──
// 运行时 consumer 已全部迁至 owning module。此文件仅保留 re-export
// 以兼容仍通过此路径 mock 的测试文件。新代码禁止从此路径 import。
// 测试迁移完成后可删除此文件。
export * from './contest-manage'
export * from './contest-announcements'
export * from './contest-operations'
export * from './contest-awd-admin'
export * from './contest-reviews'
