<script setup lang="ts">
const packageTree = `challenge-package.zip    # 上传的题目包压缩文件
  challenge.yml          # 题目清单配置（必填）
  statement.md           # 题面 Markdown 文件（必填）
  attachments/           # 题目附件目录（可选）
    web-demo.zip         # 题目附件示例
  docker/                # 教师源码与运行扩展目录（可选）
    Dockerfile           # 容器构建文件示例
    app.py               # 教师编写的 Web 服务器入口示例
    requirements.txt     # 运行依赖示例
    check/check.py       # 本地 checker 脚本示例（AWD 推荐随包提供）
    topology.yml         # 拓扑扩展配置（extensions.topology.source）`

const challengeManifest = `api_version: v1 # 固定为 v1
kind: challenge # 固定为 challenge

meta:
  slug: web-demo # 题目标识（全局唯一，建议英文小写+中划线）
  title: Web Demo # 题目展示名称
  category: web # 题目分类，允许值：web/pwn/reverse/crypto/misc/forensics（非法值会回退为 misc）
  difficulty: easy # 难度等级，允许值：beginner/easy/medium/hard/insane（非法值会回退为 easy）
  points: 100 # 分值，必须大于 0
  tags:
    - sqli # 标签（可选）
    - mysql # 标签（可选）

content:
  statement: statement.md # 题面文件路径（题目包内）
  attachments:
    - path: attachments/web-demo.zip # 附件文件路径（题目包内）
      name: web-demo.zip # 附件展示名称（可选，默认文件名）

flag:
  type: static # 判题类型：static/dynamic/regex/manual_review
  value: flag{example} # 当 type=static 或 regex 时必填
  prefix: flag # flag 前缀（可选）

hints:
  - level: 1 # 提示级别（数字越大通常越晚解锁）
    title: Hint 1 # 提示标题
    content: 先看登录流程的请求参数 # 提示内容
  - level: 2 # 第二条提示示例（可继续追加 level: 3、4...）
    title: Hint 2 # 提示标题
    content: 关注 SQL 语句拼接位置 # 提示内容

runtime:
  type: container # 运行方式，如 container/static
  image:
    ref: ctf/web-demo:latest # 容器镜像地址（type=container 时使用）

extensions:
  awd:
    checker:
      type: http_standard # AWD 推荐在题目包中声明 checker
      config:
        put_flag:
          method: PUT
          path: /api/flag
          expected_status: 200
          headers:
            X-AWD-Checker-Token: demo-checker-token
          body_template: "{{FLAG}}"
        get_flag:
          method: GET
          path: /api/flag
          expected_status: 200
          headers:
            X-AWD-Checker-Token: demo-checker-token
          expected_substring: "{{FLAG}}"
        havoc:
          method: GET
          path: /health
          expected_status: 200
  topology:
    source: docker/topology.yml # 拓扑文件路径（可选）
    enabled: false # 是否启用拓扑扩展`

const statementGuide = `statement.md 写法建议

- 不要写 # 题目名，页面外层已经展示标题
- 不要写 ## 题目描述，页面外层已经有该区块标题
- 开头直接写正文背景和任务
- 推荐结构：## 目标 / ## 访问入口(或 ## 获取方式) / ## 补充说明
- hints 已单独配置时，尽量不要再在 statement.md 重复写提示`
</script>

<template>
  <main class="content-pane">
    <div class="hero-copy">
      <div class="workspace-overline">Uploader Guide</div>
      <h1 class="hero-title">
        题目包示例
      </h1>
      <p class="hero-summary">
        上传 zip 之前，先核对目录结构、`challenge.yml` 字段和题面文件路径。教师自己写的 Web
        服务代码通常也放在 `docker/` 目录里，这里的示例与当前导入解析规则保持一致。
      </p>
    </div>

    <div class="journal-divider" />

    <div class="guide-grid">
      <article class="guide-section guide-section--plain">
        <div class="guide-section__label">
          目录结构
        </div>
        <h2 class="guide-section__title">
          建议保留最小目录
        </h2>
        <pre class="guide-code"><code>{{ packageTree }}</code></pre>
      </article>
      <article class="guide-section guide-section--plain">
        <div class="guide-section__label">
          statement.md
        </div>
        <h2 class="guide-section__title">
          题面正文写法
        </h2>
        <pre class="guide-code"><code>{{ statementGuide }}</code></pre>
      </article>
    </div>

    <article class="guide-section guide-section--full guide-section--plain">
      <div class="guide-section__label">
        challenge.yml
      </div>
      <h2 class="guide-section__title">
        最小可用示例
      </h2>
      <pre class="guide-code"><code>{{ challengeManifest }}</code></pre>
    </article>
  </main>
</template>

<style scoped>
.workspace-overline,
.guide-section__label {
  font-size: var(--font-size-0-70);
  font-weight: 700;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.hero-copy {
  max-width: 60rem;
  padding-top: 1.5rem;
}

.hero-summary {
  margin: var(--space-3-5) 0 0;
  color: var(--journal-muted);
  line-height: 1.75;
}

.guide-grid {
  display: grid;
  gap: var(--space-4);
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.guide-section {
  display: grid;
  gap: var(--space-3-5);
  padding: var(--space-4-5);
  border: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  border-radius: 1rem;
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base)),
    color-mix(in srgb, var(--journal-surface-subtle) 94%, var(--color-bg-base))
  );
}

.guide-section--plain {
  padding: 0;
  border: 0;
  border-radius: 0;
  background: transparent;
}

.guide-section--full {
  margin-top: var(--space-4);
}

.guide-section__title {
  margin: 0;
  font-size: var(--font-size-1-12);
  font-weight: 700;
  color: var(--journal-ink);
}

.guide-code {
  overflow-x: auto;
  margin: 0;
  border-radius: 0.85rem;
  border: 1px solid color-mix(in srgb, var(--journal-border) 92%, transparent);
  background: var(--color-bg-surface);
  padding: var(--space-4);
  color: var(--color-text-primary);
  font-family: var(--font-family-mono);
  font-size: var(--font-size-0-92);
  line-height: 1.7;
}

@media (max-width: 960px) {
  .guide-grid {
    grid-template-columns: 1fr;
  }
}
</style>
