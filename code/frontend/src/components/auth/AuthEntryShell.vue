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
        <div class="auth-entry-shell__kicker">
          Teaching Platform Access
        </div>
        <h1 class="auth-entry-shell__title">
          教学平台入口
        </h1>
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
  background-color: var(--color-bg-base);
  color: var(--color-text-primary);
  background-image: 
    radial-gradient(circle at 0% 0%, color-mix(in srgb, var(--color-primary) 8%, transparent) 0%, transparent 50%),
    radial-gradient(circle at 100% 100%, color-mix(in srgb, var(--color-primary) 5%, transparent) 0%, transparent 50%);
}

.auth-entry-shell__ambient {
  position: absolute;
  pointer-events: none;
  border-radius: 999px;
  filter: blur(120px);
  opacity: 0.6;
}

.auth-entry-shell__ambient--top {
  top: -12rem;
  right: -8rem;
  width: 32rem;
  height: 32rem;
  background: var(--color-primary);
}

.auth-entry-shell__ambient--side {
  bottom: -10rem;
  left: -10rem;
  width: 28rem;
  height: 28rem;
  background: color-mix(in srgb, var(--color-primary) 40%, transparent);
}

.auth-entry-shell__frame {
  position: relative;
  z-index: 1;
  display: grid;
  grid-template-columns: 1.1fr 0.9fr;
  width: min(1120px, calc(100% - 3rem));
  min-height: 680px;
  margin: 0 auto;
  border: 1px solid var(--color-border-default);
  border-radius: 2rem;
  background: var(--color-bg-surface);
  box-shadow: 
    0 40px 100px -20px color-mix(in srgb, var(--color-shadow-strong) 40%, transparent),
    0 0 0 1px color-mix(in srgb, var(--color-bg-surface) 80%, white);
  overflow: hidden;
}

.auth-entry-shell__overview {
  padding: 4.5rem;
  background: linear-gradient(135deg, var(--color-bg-elevated), var(--color-bg-surface));
  border-right: 1px solid var(--color-border-subtle);
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.auth-entry-shell__kicker {
  display: inline-flex;
  font-size: var(--font-size-11);
  font-weight: 900;
  letter-spacing: 0.25em;
  text-transform: uppercase;
  color: var(--color-primary);
  margin-bottom: 1.5rem;
}

.auth-entry-shell__title {
  margin: 0 0 1.25rem;
  font-size: clamp(32px, 4vw, 48px);
  font-weight: 900;
  letter-spacing: -0.04em;
  line-height: 1.05;
  color: var(--color-text-primary);
}

.auth-entry-shell__copy {
  font-size: var(--font-size-16);
  line-height: 1.75;
  color: var(--color-text-secondary);
  max-width: 32rem;
}

.auth-entry-shell__signals {
  display: grid;
  gap: 1.25rem;
  margin-top: 3.5rem;
}

.auth-entry-shell__signal {
  padding: 1.5rem;
  border-radius: 1.25rem;
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border-default);
  transition: all 0.3s ease;
}

.auth-entry-shell__signal:hover {
  transform: translateX(8px);
  border-color: var(--color-primary);
  box-shadow: 0 10px 30px -10px color-mix(in srgb, var(--color-primary) 20%, transparent);
}

.auth-entry-shell__signal-title {
  font-size: var(--font-size-15);
  font-weight: 900;
  color: var(--color-text-primary);
}

.auth-entry-shell__signal-copy {
  margin-top: 0.5rem;
  font-size: var(--font-size-13);
  color: var(--color-text-muted);
  line-height: 1.6;
}

.auth-entry-shell__panel {
  padding: 4.5rem;
  display: flex;
  flex-direction: column;
  justify-content: center;
  background: var(--color-bg-surface);
}

.auth-entry-shell__panel-header {
  margin-bottom: 2.5rem;
}

.auth-entry-shell__panel-eyebrow {
  font-size: var(--font-size-11);
  font-weight: 800;
  text-transform: uppercase;
  letter-spacing: 0.2em;
  color: var(--color-text-muted);
}

.auth-entry-shell__panel-title {
  margin: 0.75rem 0 0.5rem;
  font-size: var(--font-size-28);
  font-weight: 900;
  letter-spacing: -0.02em;
  color: var(--color-text-primary);
}

.auth-entry-shell__panel-copy {
  font-size: var(--font-size-14);
  color: var(--color-text-secondary);
}

.auth-entry-shell__panel-footer {
  margin-top: 2.5rem;
  padding-top: 1.5rem;
  border-top: 1px solid var(--color-border-subtle);
}

@media (max-width: 1024px) {
  .auth-entry-shell__frame {
    grid-template-columns: 1fr;
    min-height: auto;
    width: min(640px, calc(100% - 2rem));
  }
  .auth-entry-shell__overview, .auth-entry-shell__panel {
    padding: 3rem 2.5rem;
  }
  .auth-entry-shell__overview {
    border-right: none;
    border-bottom: 1px solid var(--color-border-subtle);
  }
}
</style>