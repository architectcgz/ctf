<script setup lang="ts">
defineProps<{
  panelEyebrow: string
  panelTitle: string
  panelDescription: string
}>()
</script>

<template>
  <div class="auth-entry-shell">
    <div class="auth-entry-shell__ambient" />

    <div class="auth-entry-shell__container">
      <header class="auth-entry-shell__header">
        <div class="workspace-overline">
          Teaching Platform Access
        </div>
        <h1 class="hero-title">
          {{ panelTitle }}
        </h1>
        <p class="hero-summary">
          {{ panelDescription }}
        </p>
      </header>

      <main class="auth-entry-shell__main">
        <div class="auth-entry-shell__card">
          <slot />
        </div>

        <footer
          v-if="$slots.footer"
          class="auth-entry-shell__footer"
        >
          <slot name="footer" />
        </footer>
      </main>

      <aside class="auth-entry-shell__hints">
        <div class="hint-item">
          <strong>训练空间</strong>
          <span>进入靶场、竞赛、实例与排行榜</span>
        </div>
        <div class="hint-divider" />
        <div class="hint-item">
          <strong>教学协同</strong>
          <span>班级、学生分析与报告导出</span>
        </div>
        <div class="hint-divider" />
        <div class="hint-item">
          <strong>系统值守</strong>
          <span>审计、告警与风险研判</span>
        </div>
      </aside>
    </div>
  </div>
</template>

<style scoped>
.auth-entry-shell {
  position: relative;
  min-height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  background-color: var(--color-bg-base);
  color: var(--color-text-primary);
  padding: 4rem 2rem;
  overflow-x: hidden;
}

.auth-entry-shell__ambient {
  position: absolute;
  top: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 100%;
  max-width: 80rem;
  height: 40rem;
  background: radial-gradient(
    circle at 50% 0%,
    color-mix(in srgb, var(--color-primary) 12%, transparent) 0%,
    transparent 70%
  );
  pointer-events: none;
  z-index: 0;
}

.auth-entry-shell__container {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 32rem;
  display: flex;
  flex-direction: column;
  gap: 3rem;
}

.auth-entry-shell__header {
  text-align: center;
}

.hero-title {
  margin: 0.75rem 0 0;
  font-size: var(--font-size-30);
  font-weight: 900;
  letter-spacing: -0.04em;
  color: var(--color-text-primary);
}

.hero-summary {
  margin: 1rem auto 0;
  font-size: var(--font-size-15);
  line-height: 1.7;
  color: var(--color-text-secondary);
  max-width: 24rem;
}

.auth-entry-shell__main {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.auth-entry-shell__card {
  padding: 2.5rem;
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border-default);
  border-radius: 1.5rem;
  box-shadow: 
    0 20px 50px -12px color-mix(in srgb, var(--color-shadow-strong) 15%, transparent),
    0 0 0 1px color-mix(in srgb, var(--color-bg-surface) 80%, white);
}

.auth-entry-shell__footer {
  text-align: center;
  padding: 0.5rem 0;
}

.auth-entry-shell__hints {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 1.25rem;
  margin-top: 1rem;
  opacity: 0.8;
}

.hint-item {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}

.hint-item strong {
  font-size: 11px;
  font-weight: 800;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--color-text-primary);
}

.hint-item span {
  font-size: 10px;
  color: var(--color-text-muted);
  white-space: nowrap;
}

.hint-divider {
  width: 1px;
  height: 1.5rem;
  background: var(--color-border-subtle);
}

@media (max-width: 640px) {
  .auth-entry-shell__hints {
    flex-direction: column;
    gap: 1rem;
    text-align: center;
  }
  .hint-divider {
    display: none;
  }
}
</style>