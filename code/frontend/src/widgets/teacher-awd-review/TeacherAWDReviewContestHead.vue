<template>
  <div class="teacher-directory-head">
    <span class="teacher-directory-head-cell teacher-directory-head-cell-code">代号</span>
    <span class="teacher-directory-head-cell teacher-directory-head-cell-name">赛事</span>
    <span>轮次</span>
    <span>队伍</span>
    <span>状态</span>
    <span>操作</span>
  </div>
</template>

<style scoped>
.teacher-directory-head {
  display: grid;
  grid-template-columns: var(--awd-review-directory-columns);
  gap: var(--space-4);
  padding: 0 0 var(--space-3);
  border-bottom: 1px dashed var(--teacher-divider);
  color: var(--journal-muted);
  font-size: var(--font-size-0-76);
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

@media (max-width: 1080px) {
  .teacher-directory-head {
    display: none;
  }
}
</style>
