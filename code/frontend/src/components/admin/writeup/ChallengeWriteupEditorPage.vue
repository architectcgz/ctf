<script setup lang="ts">
import { RefreshCw, Save, Trash2 } from 'lucide-vue-next'
import { computed } from 'vue'

import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import { useChallengeWriteupEditorPage } from '@/composables/useChallengeWriteupEditorPage'

const props = withDefaults(
  defineProps<{
    challengeId: string
    embedded?: boolean
  }>(),
  {
    embedded: false,
  }
)

const emit = defineEmits<{
  back: []
}>()

const isEmbedded = computed(() => props.embedded)
const pageShellClass = computed(() =>
  isEmbedded.value
    ? 'writeup-embedded-shell'
    : 'workspace-shell journal-shell journal-shell-admin journal-notes-card journal-hero flex min-h-full flex-1 flex-col'
)
const {
  loading,
  saving,
  deleting,
  togglingRecommendation,
  challenge,
  writeup,
  form,
  hasWriteup,
  visibilityLabel,
  loadPage,
  handleSave,
  handleDelete,
  handleToggleRecommendation,
  restoreExistingWriteup,
} = useChallengeWriteupEditorPage(props.challengeId)
</script>

<template>
  <component
    :is="isEmbedded ? 'div' : 'section'"
    :class="pageShellClass"
  >
    <header v-if="!isEmbedded" class="workspace-topbar">
      <div class="topbar-leading">
        <span class="workspace-overline">Challenge Workspace</span>
        <span class="class-chip">题解管理</span>
      </div>
      <div class="writeup-top-actions">
        <button class="ui-btn ui-btn--ghost" type="button" @click="emit('back')">
          返回题目
        </button>
        <button class="ui-btn ui-btn--ghost" type="button" @click="void loadPage()">
          <RefreshCw class="h-4 w-4" />
          刷新
        </button>
      </div>
    </header>

    <PageHeader
      v-if="!isEmbedded"
      class="writeup-page-header"
      eyebrow="Admin Writeup"
      title="题解管理"
      :description="
        challenge
          ? `为《${challenge.title}》维护管理员题解，控制公开范围。`
          : '为题目维护管理员题解，控制公开范围。'
      "
    />
    <div v-else class="list-heading writeup-tab-heading">
      <div>
        <div class="workspace-overline">Admin Writeup</div>
        <h1 class="workspace-page-title">题解管理</h1>
      </div>
      <p class="workspace-page-copy">
        {{
          challenge
            ? `为《${challenge.title}》维护管理员题解，控制公开范围。`
            : '为题目维护管理员题解，控制公开范围。'
        }}
      </p>
    </div>

    <div class="journal-divider" />

    <AppLoading v-if="loading" class="writeup-loading">正在加载题解数据...</AppLoading>

    <main v-else :class="isEmbedded ? 'writeup-workspace' : 'content-pane writeup-workspace'">
      <section class="writeup-main">
        <section class="writeup-section writeup-editor-section">
          <header class="writeup-editor-head">
            <div>
              <div class="journal-note-label">Writeup Editor</div>
              <h2 class="writeup-section-title">编辑器</h2>
            </div>
            <div class="writeup-badges">
              <span
                class="writeup-badge"
                :class="hasWriteup ? 'writeup-badge--ok' : 'writeup-badge--warn'"
              >
                {{ hasWriteup ? '已存在题解' : '尚未创建' }}
              </span>
              <span v-if="writeup?.is_recommended" class="writeup-badge writeup-badge--accent">
                推荐题解
              </span>
            </div>
          </header>

          <div class="writeup-form-grid">
            <label class="writeup-field writeup-field--title">
              <span class="writeup-field-label">题解标题</span>
              <input
                v-model="form.title"
                type="text"
                class="writeup-field-input"
                placeholder="例如：官方解题思路 / 赛后复盘"
              />
            </label>

            <label class="writeup-field writeup-field--visibility">
              <span class="writeup-field-label">可见性</span>
              <select v-model="form.visibility" class="writeup-field-input">
                <option value="private">private</option>
                <option value="public">public</option>
              </select>
            </label>
          </div>

          <div class="writeup-visibility-note">{{ visibilityLabel }}</div>

          <label class="writeup-field writeup-field--content">
            <span class="writeup-field-label">题解正文</span>
            <textarea
              v-model="form.content"
              rows="16"
              class="writeup-content-input"
              placeholder="输入官方题解、赛后复盘或教学讲解内容。"
            />
          </label>

          <div class="writeup-editor-actions" role="group" aria-label="题解编辑操作">
            <button
              :disabled="saving"
              class="ui-btn ui-btn--primary"
              type="button"
              @click="void handleSave()"
            >
              <Save class="h-4 w-4" />
              {{ saving ? '保存中...' : '保存题解' }}
            </button>
            <button
              v-if="hasWriteup"
              :disabled="togglingRecommendation"
              class="ui-btn ui-btn--secondary"
              type="button"
              @click="void handleToggleRecommendation()"
            >
              {{
                togglingRecommendation
                  ? '处理中...'
                  : writeup?.is_recommended
                    ? '取消推荐'
                    : '设为推荐'
              }}
            </button>
            <button
              v-if="hasWriteup"
              class="ui-btn ui-btn--ghost"
              type="button"
              @click="restoreExistingWriteup"
            >
              恢复已保存版本
            </button>
            <button
              v-if="hasWriteup"
              :disabled="deleting"
              class="ui-btn ui-btn--danger"
              type="button"
              @click="void handleDelete()"
            >
              <Trash2 class="h-4 w-4" />
              {{ deleting ? '删除中...' : '删除题解' }}
            </button>
          </div>
        </section>

        <section class="writeup-section writeup-snapshot-section">
          <header class="writeup-subsection-head">
            <div class="journal-note-label">Snapshot</div>
            <h2 class="writeup-section-title">当前已保存版本</h2>
          </header>

          <template v-if="writeup">
            <dl class="writeup-snapshot-grid">
              <div class="writeup-snapshot-item">
                <dt>标题</dt>
                <dd>{{ writeup.title }}</dd>
              </div>
              <div class="writeup-snapshot-item">
                <dt>可见性</dt>
                <dd>{{ writeup.visibility }}</dd>
              </div>
              <div class="writeup-snapshot-item">
                <dt>推荐状态</dt>
                <dd>{{ writeup.is_recommended ? '推荐题解' : '未推荐' }}</dd>
              </div>
              <div class="writeup-snapshot-item">
                <dt>创建时间</dt>
                <dd>{{ writeup.created_at }}</dd>
              </div>
              <div class="writeup-snapshot-item">
                <dt>更新时间</dt>
                <dd>{{ writeup.updated_at }}</dd>
              </div>
            </dl>
          </template>

          <AppEmpty
            v-else
            title="当前还没有管理员题解"
            description="填写表单后点击保存，即可创建题解并控制公开范围。"
            icon="BookOpen"
          />
        </section>
      </section>

      <aside class="context-rail writeup-rail">
        <div class="writeup-rail-card">
          <div class="journal-note-label">Challenge</div>
          <h2 class="writeup-rail-title">题目信息</h2>

          <dl v-if="challenge" class="writeup-rail-meta">
            <div>
              <dt>标题</dt>
              <dd>{{ challenge.title }}</dd>
            </div>
            <div>
              <dt>分类</dt>
              <dd>{{ challenge.category }}</dd>
            </div>
            <div>
              <dt>状态</dt>
              <dd>{{ challenge.status }}</dd>
            </div>
            <div>
              <dt>难度</dt>
              <dd>{{ challenge.difficulty }}</dd>
            </div>
            <div>
              <dt>分值</dt>
              <dd>{{ challenge.points }}</dd>
            </div>
          </dl>
          <div v-else class="writeup-rail-empty">题目信息加载中。</div>
        </div>
      </aside>
    </main>
  </component>
</template>

<style scoped>
.journal-shell,
.writeup-embedded-shell {
  --journal-topbar-padding-bottom: var(--space-3);
  --journal-shell-hero-radial-strength: 7%;
  --journal-shell-hero-radial-size: 22rem;
  --journal-shell-hero-end: var(--journal-surface);
  --journal-shell-hero-shadow: 0 22px 50px var(--color-shadow-soft);
}

.writeup-embedded-shell {
  display: grid;
  gap: var(--space-5);
}

.writeup-tab-heading {
  margin-bottom: var(--space-1);
}

.list-heading {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: var(--space-3);
}

.workspace-overline {
  font-size: var(--font-size-0-70);
  font-weight: 700;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.writeup-top-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.writeup-top-actions > .ui-btn,
.writeup-editor-actions > .ui-btn {
  --ui-btn-height: 2.25rem;
  --ui-btn-radius: 0.65rem;
  --ui-btn-padding: var(--space-2) var(--space-3-5);
  --ui-btn-font-size: var(--font-size-0-84);
  --ui-btn-font-weight: 600;
  --ui-btn-focus-ring: color-mix(in srgb, var(--journal-accent) 18%, transparent);
}

.writeup-top-actions > .ui-btn.ui-btn--ghost,
.writeup-editor-actions > .ui-btn.ui-btn--ghost {
  --ui-btn-border: var(--journal-border);
  --ui-btn-background: color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base));
  --ui-btn-color: var(--journal-ink);
  --ui-btn-hover-border: color-mix(in srgb, var(--journal-accent) 28%, transparent);
  --ui-btn-hover-background: color-mix(in srgb, var(--journal-accent) 4%, var(--journal-surface));
  --ui-btn-hover-color: var(--journal-accent);
}

.writeup-editor-actions > .ui-btn.ui-btn--primary {
  --ui-btn-primary-border: color-mix(in srgb, var(--journal-accent) 18%, transparent);
  --ui-btn-primary-background: var(--journal-accent);
  --ui-btn-primary-hover-background: color-mix(in srgb, var(--journal-accent) 88%, black);
}

.writeup-editor-actions > .ui-btn.ui-btn--secondary {
  --ui-btn-secondary-border: color-mix(in srgb, var(--journal-accent) 30%, transparent);
  --ui-btn-secondary-background: color-mix(in srgb, var(--journal-accent) 12%, transparent);
  --ui-btn-secondary-color: var(--journal-accent);
  --ui-btn-secondary-hover-border: color-mix(in srgb, var(--journal-accent) 38%, transparent);
  --ui-btn-secondary-hover-background: color-mix(in srgb, var(--journal-accent) 18%, transparent);
  --ui-btn-secondary-hover-color: color-mix(in srgb, var(--journal-accent) 92%, white);
}

.writeup-editor-actions > .ui-btn.ui-btn--danger {
  --ui-btn-danger-border: color-mix(in srgb, var(--color-danger) 28%, transparent);
  --ui-btn-danger-background: color-mix(in srgb, var(--color-danger) 10%, transparent);
  --ui-btn-danger-color: var(--color-danger);
  --ui-btn-danger-hover-border: color-mix(in srgb, var(--color-danger) 34%, transparent);
  --ui-btn-danger-hover-background: color-mix(in srgb, var(--color-danger) 14%, transparent);
}

.writeup-loading {
  padding-block: var(--space-7);
}

.writeup-workspace {
  display: grid;
  gap: var(--space-6);
  grid-template-columns: minmax(0, 1fr) minmax(16.5rem, 18.5rem);
  align-items: start;
}

.writeup-main {
  display: grid;
  gap: var(--space-6);
}

.writeup-section {
  display: grid;
  gap: var(--space-4);
}

.writeup-snapshot-section {
  border-top: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  padding-top: var(--space-5);
}

.writeup-editor-head {
  display: flex;
  flex-wrap: wrap;
  justify-content: space-between;
  align-items: flex-start;
  gap: var(--space-3);
}

.writeup-subsection-head {
  display: grid;
  gap: var(--space-1);
}

.writeup-section-title {
  margin: 0;
  font-size: var(--font-size-1-08);
  font-weight: 700;
  color: var(--journal-ink);
}

.writeup-badges {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.writeup-badge {
  display: inline-flex;
  align-items: center;
  min-height: 2rem;
  border-radius: 999px;
  border: 1px solid transparent;
  padding: 0 var(--space-3);
  font-size: var(--font-size-0-76);
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.writeup-badge--ok {
  border-color: color-mix(in srgb, var(--color-success) 30%, transparent);
  background: color-mix(in srgb, var(--color-success) 12%, transparent);
  color: var(--color-success);
}

.writeup-badge--warn {
  border-color: color-mix(in srgb, var(--color-warning) 30%, transparent);
  background: color-mix(in srgb, var(--color-warning) 12%, transparent);
  color: var(--color-warning);
}

.writeup-badge--accent {
  border-color: color-mix(in srgb, var(--journal-accent) 30%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 12%, transparent);
  color: var(--journal-accent);
}

.writeup-form-grid {
  display: grid;
  gap: var(--space-4);
  grid-template-columns: minmax(0, 1fr) minmax(11rem, 12.5rem);
}

.writeup-field {
  display: grid;
  gap: var(--space-2);
}

.writeup-field--schedule,
.writeup-field--content,
.writeup-field--title {
  grid-column: 1 / -1;
}

.writeup-field-label {
  font-size: var(--font-size-0-84);
  font-weight: 600;
  color: var(--journal-ink);
}

.writeup-field-input,
.writeup-content-input {
  width: 100%;
  border-radius: 0.95rem;
  border: 1px solid var(--journal-border);
  background: color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base));
  color: var(--journal-ink);
  padding: var(--space-3) var(--space-3-5);
  font-size: var(--font-size-0-90);
  outline: none;
  transition:
    border-color 0.16s ease,
    box-shadow 0.16s ease;
}

.writeup-content-input {
  min-height: 22rem;
  line-height: 1.68;
}

.writeup-field-input:focus,
.writeup-content-input:focus {
  border-color: color-mix(in srgb, var(--journal-accent) 44%, transparent);
  box-shadow: 0 0 0 4px color-mix(in srgb, var(--journal-accent) 16%, transparent);
}

.writeup-visibility-note {
  border-left: 2px solid color-mix(in srgb, var(--journal-accent) 36%, transparent);
  padding: var(--space-2) var(--space-3);
  font-size: var(--font-size-0-86);
  line-height: 1.65;
  color: var(--journal-muted);
  background: color-mix(in srgb, var(--journal-surface-subtle) 88%, var(--color-bg-base));
}

.writeup-editor-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.writeup-snapshot-grid {
  display: grid;
  gap: 0;
  margin: 0;
  border-top: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.writeup-snapshot-item {
  display: grid;
  gap: var(--space-1);
  padding: var(--space-3-5) 0;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.writeup-snapshot-item dt {
  font-size: var(--font-size-0-76);
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.writeup-snapshot-item dd {
  margin: 0;
  font-size: var(--font-size-0-90);
  font-weight: 600;
  color: var(--journal-ink);
  word-break: break-word;
}

.writeup-rail {
  border-left: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  padding-left: var(--space-4);
}

.writeup-rail-card {
  display: grid;
  gap: var(--space-3);
}

.writeup-rail-title {
  margin: 0;
  font-size: var(--font-size-1-02);
  color: var(--journal-ink);
}

.writeup-rail-meta {
  display: grid;
  gap: 0;
  margin: 0;
  border-top: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.writeup-rail-meta > div {
  display: grid;
  gap: var(--space-1);
  padding: var(--space-3) 0;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.writeup-rail-meta dt {
  font-size: var(--font-size-0-74);
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.writeup-rail-meta dd {
  margin: 0;
  font-size: var(--font-size-0-88);
  color: var(--journal-ink);
  word-break: break-word;
}

.writeup-rail-empty {
  font-size: var(--font-size-0-84);
  color: var(--journal-muted);
}

@media (max-width: 1100px) {
  .writeup-workspace {
    grid-template-columns: 1fr;
  }

  .writeup-rail {
    border-left: 0;
    border-top: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
    padding-left: 0;
    padding-top: var(--space-4);
  }
}

@media (max-width: 760px) {
  .list-heading {
    align-items: flex-start;
    flex-direction: column;
  }

  .writeup-form-grid {
    grid-template-columns: 1fr;
  }

  .writeup-top-actions {
    width: 100%;
  }

  .writeup-top-actions .ui-btn {
    flex: 1 1 auto;
  }
}
</style>
