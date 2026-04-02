<script setup lang="ts">
interface AuthSignalItem {
  key: string
  title: string
  description: string
}

defineProps<{
  panelEyebrow: string
  panelTitle: string
  panelDescription: string
}>()

const signals: AuthSignalItem[] = [
  {
    key: 'training',
    title: '训练空间',
    description: '进入靶场、竞赛、实例与排行榜视图。',
  },
  {
    key: 'teaching',
    title: '教学协同',
    description: '班级、学生分析与报告导出在同一平台完成。',
  },
  {
    key: 'operations',
    title: '系统值守',
    description: '支持管理员进入审计、告警与风险研判链路。',
  },
]
</script>

<template>
  <div class="auth-entry-shell">
    <div class="auth-entry-shell__ambient auth-entry-shell__ambient--top" />
    <div class="auth-entry-shell__ambient auth-entry-shell__ambient--side" />

    <div class="auth-entry-shell__frame">
      <section class="auth-entry-shell__overview">
        <div class="auth-entry-shell__kicker">Teaching Platform Access</div>
        <h1 class="auth-entry-shell__title">教学平台入口</h1>
        <p class="auth-entry-shell__copy">
          学生训练、教师教学和管理员值守共用同一套账号体系。登录后按角色进入对应工作台。
        </p>

        <div class="auth-entry-shell__signals">
          <article
            v-for="signal in signals"
            :key="signal.key"
            class="auth-entry-shell__signal"
          >
            <div class="auth-entry-shell__signal-title">
              {{ signal.title }}
            </div>
            <p class="auth-entry-shell__signal-copy">
              {{ signal.description }}
            </p>
          </article>
        </div>
      </section>

      <section class="auth-entry-shell__panel">
        <header class="auth-entry-shell__panel-header">
          <div class="auth-entry-shell__panel-eyebrow">
            {{ panelEyebrow }}
          </div>
          <h2 class="auth-entry-shell__panel-title">
            {{ panelTitle }}
          </h2>
          <p class="auth-entry-shell__panel-copy">
            {{ panelDescription }}
          </p>
        </header>

        <div class="auth-entry-shell__panel-body">
          <slot />
        </div>

        <footer
          v-if="$slots.footer"
          class="auth-entry-shell__panel-footer"
        >
          <slot name="footer" />
        </footer>
      </section>
    </div>
  </div>
</template>

<style scoped>
.auth-entry-shell {
  position: relative;
  min-height: 100vh;
  display: flex;
  align-items: center;
  overflow: hidden;
  padding: 1rem 0;
  background: var(--color-bg-base);
  color: var(--color-text-primary);
}

.auth-entry-shell__ambient {
  position: absolute;
  pointer-events: none;
  border-radius: 999px;
  filter: blur(80px);
}

.auth-entry-shell__ambient--top {
  top: -9rem;
  right: -5rem;
  width: 22rem;
  height: 22rem;
  background: color-mix(in srgb, var(--color-primary) 18%, transparent);
}

.auth-entry-shell__ambient--side {
  bottom: -8rem;
  left: -6rem;
  width: 20rem;
  height: 20rem;
  background: color-mix(in srgb, var(--color-primary) 10%, transparent);
}

.auth-entry-shell__frame {
  position: relative;
  z-index: 1;
  display: grid;
  grid-template-columns: minmax(0, 1.02fr) minmax(0, 0.98fr);
  gap: 0;
  width: min(1040px, calc(100% - 2rem));
  min-height: min(640px, calc(100vh - 2rem));
  margin: 0 auto;
  border: 1px solid color-mix(in srgb, var(--color-border-default) 78%, transparent);
  border-radius: 1.5rem;
  background:
    linear-gradient(180deg, color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base)), color-mix(in srgb, var(--color-bg-surface) 86%, var(--color-bg-base))),
    radial-gradient(circle at top right, color-mix(in srgb, var(--color-primary) 12%, transparent), transparent 18rem);
  box-shadow: 0 22px 48px var(--color-shadow-strong);
  overflow: hidden;
}

.auth-entry-shell__overview,
.auth-entry-shell__panel {
  padding: 1.75rem;
}

.auth-entry-shell__overview {
  border-right: 1px solid color-mix(in srgb, var(--color-border-default) 76%, transparent);
}

.auth-entry-shell__kicker,
.auth-entry-shell__panel-eyebrow {
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.18em;
  text-transform: uppercase;
}

.auth-entry-shell__kicker {
  display: inline-flex;
  align-items: center;
  min-height: 2rem;
  padding: 0 0.8rem;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--color-primary) 28%, var(--color-border-default));
  background: color-mix(in srgb, var(--color-primary) 10%, transparent);
  color: color-mix(in srgb, var(--color-primary-hover) 82%, white);
}

.auth-entry-shell__title {
  margin: 1.1rem 0 0.6rem;
  font-size: clamp(1.8rem, 2.6vw, 2.45rem);
  line-height: 1.1;
  font-weight: 700;
  letter-spacing: -0.03em;
}

.auth-entry-shell__copy,
.auth-entry-shell__panel-copy,
.auth-entry-shell__signal-copy {
  color: var(--color-text-secondary);
  line-height: 1.7;
}

.auth-entry-shell__copy {
  max-width: 34rem;
  font-size: 0.92rem;
}

.auth-entry-shell__signals {
  display: grid;
  gap: 0.75rem;
  margin-top: 1.4rem;
}

.auth-entry-shell__signal {
  padding: 0.85rem 0.95rem 0.85rem 1rem;
  border: 1px solid color-mix(in srgb, var(--color-border-default) 80%, transparent);
  border-left: 3px solid color-mix(in srgb, var(--color-primary) 58%, transparent);
  border-radius: 1rem;
  background: color-mix(in srgb, var(--color-bg-elevated) 78%, transparent);
}

.auth-entry-shell__signal-title {
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.auth-entry-shell__signal-copy {
  margin-top: 0.35rem;
  font-size: 0.8rem;
}

.auth-entry-shell__panel {
  display: flex;
  flex-direction: column;
}

.auth-entry-shell__panel-header {
  padding-bottom: 1rem;
  border-bottom: 1px dashed color-mix(in srgb, var(--color-border-default) 76%, transparent);
}

.auth-entry-shell__panel-eyebrow {
  color: var(--color-text-muted);
}

.auth-entry-shell__panel-title {
  margin: 0.6rem 0 0.3rem;
  font-size: 1.6rem;
  line-height: 1.15;
  font-weight: 700;
}

.auth-entry-shell__panel-copy {
  font-size: 0.88rem;
}

.auth-entry-shell__panel-body {
  margin-top: 1.2rem;
}

.auth-entry-shell__panel-footer {
  margin-top: auto;
  padding-top: 1rem;
  border-top: 1px dashed color-mix(in srgb, var(--color-border-default) 76%, transparent);
}

:global([data-theme='light']) .auth-entry-shell__frame {
  box-shadow: 0 18px 40px var(--color-shadow-soft);
}

:global([data-theme='light']) .auth-entry-shell__signal,
:global([data-theme='light']) .auth-entry-shell__frame {
  background:
    linear-gradient(180deg, color-mix(in srgb, var(--color-bg-surface) 96%, white), color-mix(in srgb, var(--color-bg-elevated) 94%, white)),
    radial-gradient(circle at top right, color-mix(in srgb, var(--color-primary) 7%, transparent), transparent 18rem);
}

@media (max-width: 960px) {
  .auth-entry-shell__frame {
    grid-template-columns: 1fr;
    width: min(720px, calc(100% - 1rem));
    min-height: auto;
  }

  .auth-entry-shell__overview,
  .auth-entry-shell__panel {
    padding: 1.25rem;
  }

  .auth-entry-shell__overview {
    border-right: none;
    border-bottom: 1px solid color-mix(in srgb, var(--color-border-default) 76%, transparent);
  }
}
</style>
