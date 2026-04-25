<script setup lang="ts">
defineProps<{
  panelEyebrow: string
  panelTitle: string
  panelDescription: string
}>()

const emit = defineEmits<{
  heroProbe: []
}>()
</script>

<template>
  <div class="auth-entry-shell">
    <!-- 背景层：网格与环境光 -->
    <div class="auth-entry-shell__bg">
      <div class="technical-grid" />
      <div class="ambient-glow ambient-glow--1" />
      <div class="ambient-glow ambient-glow--2" />
    </div>

    <div class="auth-entry-shell__container">
      <!-- 左侧：视觉锚点 -->
      <section
        class="auth-entry-shell__hero"
        @click="emit('heroProbe')"
      >
        <header class="hero-branding">
          <div class="branding-overline">CTF Platform Infrastructure</div>
          <h1 class="branding-title">
            SECURE<br />PLATFORM<br /><span>ACCESS</span>
          </h1>
          <div class="branding-decoration" />
        </header>
        
        <div class="hero-features">
          <p class="hero-copy">
            统一身份认证系统。集成训练靶场、教学协同与系统值守链路，为网络安全实战提供全栈技术支持。
          </p>
          <div class="signal-list">
            <div class="signal-item">
              <span class="signal-dot" />
              <span>训练空间</span>
            </div>
            <div class="signal-item">
              <span class="signal-dot" />
              <span>教学协同</span>
            </div>
            <div class="signal-item">
              <span class="signal-dot" />
              <span>系统值守</span>
            </div>
          </div>
        </div>
      </section>

      <!-- 右侧：表单面板 -->
      <main class="auth-entry-shell__content">
        <div class="auth-panel">
          <header class="auth-panel__header">
            <div class="auth-panel__eyebrow">{{ panelEyebrow }}</div>
            <h2 class="auth-panel__title">{{ panelTitle }}</h2>
            <p class="auth-panel__desc">{{ panelDescription }}</p>
          </header>

          <div class="auth-panel__body">
            <slot />
          </div>

          <footer v-if="$slots.footer" class="auth-panel__footer">
            <slot name="footer" />
          </footer>
        </div>
      </main>
    </div>
  </div>
</template>

<style scoped>
.auth-entry-shell {
  position: relative;
  min-height: 100vh;
  width: 100%;
  background-color: var(--color-bg-base);
  color: var(--color-text-primary);
  display: flex;
  overflow: hidden;
}

/* 背景系统 */
.auth-entry-shell__bg {
  position: absolute;
  inset: 0;
  z-index: 0;
}

.technical-grid {
  position: absolute;
  inset: 0;
  background-image: 
    linear-gradient(var(--color-border-subtle) 1px, transparent 1px),
    linear-gradient(90deg, var(--color-border-subtle) 1px, transparent 1px);
  background-size: var(--space-16, 4rem) var(--space-16, 4rem);
  opacity: 0.15;
  mask-image: radial-gradient(circle at 50% 50%, var(--color-bg-base), transparent 80%);
}

.ambient-glow {
  position: absolute;
  border-radius: 50%;
  filter: blur(160px);
  opacity: 0.4;
}

.ambient-glow--1 {
  top: -10%; left: -10%;
  width: 40rem; height: 40rem;
  background: var(--color-primary);
}

.ambient-glow--2 {
  bottom: -5%; right: -5%;
  width: 30rem; height: 30rem;
  background: color-mix(in srgb, var(--color-primary) 30%, transparent);
}

.auth-entry-shell__container {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 1440px;
  margin: 0 auto;
  display: grid;
  grid-template-columns: 1.2fr 0.8fr;
  padding: var(--space-8) var(--space-12);
  align-items: center;
}

/* 左侧设计 */
.auth-entry-shell__hero {
  padding-right: var(--space-12);
}

.hero-branding {
  position: relative;
  margin-bottom: calc(var(--space-8) * 2);
}

.branding-overline {
  font-size: var(--font-size-12);
  font-weight: 800;
  letter-spacing: 0.4em;
  text-transform: uppercase;
  color: var(--color-primary);
  margin-bottom: var(--space-6);
}

.branding-title {
  font-size: clamp(48px, 6vw, 84px);
  font-weight: 900;
  line-height: 0.9;
  letter-spacing: -0.05em;
  margin: 0;
  color: var(--color-text-primary);
}

.branding-title span {
  color: transparent;
  -webkit-text-stroke: 1.5px var(--color-text-primary);
}

.branding-decoration {
  margin-top: var(--space-8);
  width: var(--space-16, 4rem);
  height: var(--space-1);
  background: var(--color-primary);
}

.hero-features {
  max-width: 32rem;
}

.hero-copy {
  font-size: var(--font-size-15);
  line-height: 1.8;
  color: var(--color-text-secondary);
  margin-bottom: var(--space-12);
}

.signal-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
}

.signal-item {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  font-size: var(--font-size-11);
  font-weight: 800;
  letter-spacing: 0.2em;
  color: var(--color-text-muted);
}

.signal-dot {
  width: var(--space-1-5);
  height: var(--space-1-5);
  background: var(--color-primary);
  border-radius: 50%;
  box-shadow: 0 0 10px var(--color-primary);
}

/* 右侧表单面板 */
.auth-panel {
  background: color-mix(in srgb, var(--color-bg-surface) 60%, transparent);
  backdrop-filter: blur(24px);
  border: 1px solid var(--color-border-default);
  border-radius: var(--space-6);
  padding: var(--space-10) var(--space-12);
  box-shadow: 
    0 40px 100px -20px color-mix(in srgb, var(--color-shadow-strong) 40%, transparent),
    inset 0 0 0 1px color-mix(in srgb, white 10%, transparent);
}

.auth-panel__eyebrow {
  font-size: var(--font-size-10, 10px);
  font-weight: 900;
  letter-spacing: 0.3em;
  text-transform: uppercase;
  color: var(--color-text-muted);
  margin-bottom: var(--space-3);
}

.auth-panel__title {
  font-size: var(--font-size-24);
  font-weight: 900;
  letter-spacing: -0.02em;
  margin: 0 0 var(--space-2);
}

.auth-panel__desc {
  font-size: var(--font-size-13);
  color: var(--color-text-secondary);
  line-height: 1.6;
  margin: 0 0 var(--space-12);
}

.auth-panel__footer {
  margin-top: var(--space-8);
  padding-top: var(--space-6);
  border-top: 1px solid var(--color-border-subtle);
  text-align: center;
}

@media (max-width: 1200px) {
  .auth-entry-shell__container {
    grid-template-columns: 1fr;
    padding: var(--space-12) var(--space-8);
  }
  .auth-entry-shell__hero {
    display: none;
  }
  .auth-panel {
    max-width: 32rem;
    margin: 0 auto;
    padding: var(--space-12);
  }
}
</style>
