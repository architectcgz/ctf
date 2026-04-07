<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()

const packageTree = `challenge-package.zip
  challenge.yml
  statement.md
  attachments/
    web-demo.zip
  docker/
    topology.yml`

const challengeManifest = `api_version: v1
kind: challenge

meta:
  slug: web-demo
  title: Web Demo
  category: web
  difficulty: easy
  points: 100
  tags:
    - sqli
    - mysql

content:
  statement: statement.md
  attachments:
    - path: attachments/web-demo.zip
      name: web-demo.zip

flag:
  type: static
  value: flag{example}
  prefix: flag

hints:
  - level: 1
    title: Hint 1
    content: 先看登录流程的请求参数

runtime:
  type: container
  image:
    ref: ctf/web-demo:latest

extensions:
  topology:
    source: docker/topology.yml
    enabled: false`
</script>

<template>
  <section
    class="journal-shell journal-hero flex min-h-full flex-1 flex-col rounded-[24px] border px-6 py-6 md:px-8"
  >
    <header class="workspace-topbar">
      <div class="topbar-leading">
        <span class="workspace-overline">Challenge Package</span>
        <span class="class-chip">上传示例</span>
      </div>
      <button class="nav-back" type="button" @click="router.push({ name: 'ChallengeManage', query: { panel: 'import' } })">
        返回导入题目包
      </button>
    </header>

    <div class="hero-copy">
      <div class="journal-eyebrow">Uploader Guide</div>
      <h1 class="hero-title">题目包示例</h1>
      <p class="hero-summary">
        上传 zip 之前，先核对目录结构、`challenge.yml` 字段和题面文件路径。这里的示例与当前导入解析规则保持一致。
      </p>
    </div>

    <div class="journal-divider" />

    <div class="guide-grid">
      <article class="guide-section">
        <div class="guide-section__label">目录结构</div>
        <h2 class="guide-section__title">建议保留最小目录</h2>
        <pre class="guide-code"><code>{{ packageTree }}</code></pre>
      </article>

      <article class="guide-section">
        <div class="guide-section__label">关键约束</div>
        <h2 class="guide-section__title">上传前先自查</h2>
        <ul class="guide-list">
          <li>`challenge.yml` 必须位于 zip 根目录，或 zip 根目录下唯一子目录的根部。</li>
          <li>`api_version` 当前只支持 `v1`，`kind` 必须为 `challenge`。</li>
          <li>`statement.md` 不能为空；若写了其他路径，必须仍然位于题目包内部。</li>
          <li>`meta.points` 必须大于 0。</li>
          <li>`flag.type` 支持 `static`、`dynamic`、`regex`、`manual_review`。</li>
          <li>若 `flag.type` 为 `static` 或 `regex`，必须提供 `flag.value`。</li>
          <li>附件路径必须指向包内文件，不能越出题目包目录。</li>
        </ul>
      </article>
    </div>

    <article class="guide-section guide-section--full">
      <div class="guide-section__label">challenge.yml</div>
      <h2 class="guide-section__title">最小可用示例</h2>
      <pre class="guide-code"><code>{{ challengeManifest }}</code></pre>
    </article>
  </section>
</template>

<style scoped>
.journal-shell {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-accent: var(--color-primary);
  --journal-border: color-mix(in srgb, var(--color-border-default) 84%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 78%, var(--color-bg-base));
}

.journal-hero {
  border-color: var(--journal-border);
  background:
    radial-gradient(circle at top right, color-mix(in srgb, var(--journal-accent) 7%, transparent), transparent 22rem),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base)),
      var(--journal-surface)
    );
  box-shadow: 0 22px 50px var(--color-shadow-soft);
}

.workspace-topbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem 1rem;
}

.topbar-leading {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.65rem;
}

.workspace-overline,
.journal-eyebrow,
.guide-section__label {
  font-size: 0.7rem;
  font-weight: 700;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.class-chip {
  display: inline-flex;
  align-items: center;
  min-height: 30px;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--journal-accent) 26%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  padding: 0.25rem 0.7rem;
  font-size: 0.76rem;
  font-weight: 600;
  color: var(--journal-accent);
}

.nav-back {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 2.35rem;
  border-radius: 0.75rem;
  border: 1px solid var(--journal-border);
  padding: 0.45rem 0.85rem;
  background: color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base));
  color: var(--journal-ink);
  font-size: 0.88rem;
  font-weight: 600;
}

.hero-copy {
  max-width: 60rem;
  padding-top: 1.5rem;
}

.hero-summary {
  margin: 0.9rem 0 0;
  color: var(--journal-muted);
  line-height: 1.75;
}

.journal-divider {
  margin-block: 1.5rem;
  border-top: 1px dashed color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.guide-grid {
  display: grid;
  gap: 1rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.guide-section {
  display: grid;
  gap: 0.85rem;
  padding: 1.1rem;
  border: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  border-radius: 1rem;
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base)),
    color-mix(in srgb, var(--journal-surface-subtle) 94%, var(--color-bg-base))
  );
}

.guide-section--full {
  margin-top: 1rem;
}

.guide-section__title {
  margin: 0;
  font-size: 1.12rem;
  font-weight: 700;
  color: var(--journal-ink);
}

.guide-code {
  overflow-x: auto;
  margin: 0;
  border-radius: 0.85rem;
  border: 1px solid color-mix(in srgb, var(--journal-border) 92%, transparent);
  background: color-mix(in srgb, var(--color-bg-base) 80%, #0f172a 20%);
  padding: 1rem;
  color: var(--journal-ink);
  font-family: 'IBM Plex Mono', 'JetBrains Mono', 'SFMono-Regular', 'Consolas', monospace;
  font-size: 0.84rem;
  line-height: 1.7;
}

.guide-list {
  display: grid;
  gap: 0.65rem;
  margin: 0;
  padding-left: 1.15rem;
  color: var(--journal-muted);
  line-height: 1.7;
}

@media (max-width: 960px) {
  .guide-grid {
    grid-template-columns: 1fr;
  }
}
</style>
