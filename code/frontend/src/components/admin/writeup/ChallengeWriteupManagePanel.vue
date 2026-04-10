<script setup lang="ts">
import { MoreHorizontal } from 'lucide-vue-next'
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'

import { deleteChallengeWriteup, getChallengeWriteup } from '@/api/admin'
import { getTeacherWriteupSubmissions } from '@/api/teacher'
import type { AdminChallengeWriteupData, TeacherSubmissionWriteupItemData } from '@/api/contracts'
import AdminPaginationControls from '@/components/admin/AdminPaginationControls.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import { confirmDestructiveAction } from '@/composables/useDestructiveConfirm'
import { useToast } from '@/composables/useToast'

const props = defineProps<{
  challengeId: string
  challengeTitle?: string
}>()

const router = useRouter()
const toast = useToast()

const loading = ref(true)
const deleting = ref(false)
const submissionLoading = ref(true)
const actionMenuOpen = ref(false)
const writeup = ref<AdminChallengeWriteupData | null>(null)
const writeupSubmissions = ref<TeacherSubmissionWriteupItemData[]>([])
const submissionPage = ref(1)
const submissionPageSize = ref(6)
const submissionTotal = ref(0)

const submissionTotalPages = computed(() =>
  Math.max(1, Math.ceil(submissionTotal.value / Math.max(1, submissionPageSize.value)))
)

async function loadWriteup() {
  if (!props.challengeId) {
    writeup.value = null
    loading.value = false
    return
  }

  loading.value = true
  try {
    writeup.value = await getChallengeWriteup(props.challengeId)
  } catch {
    toast.error('加载题解目录失败')
  } finally {
    loading.value = false
  }
}

async function loadWriteupSubmissions(targetPage = 1) {
  if (!props.challengeId) {
    writeupSubmissions.value = []
    submissionPage.value = 1
    submissionTotal.value = 0
    submissionLoading.value = false
    return
  }

  submissionLoading.value = true
  try {
    const payload = await getTeacherWriteupSubmissions({
      challenge_id: props.challengeId,
      page: targetPage,
      page_size: submissionPageSize.value,
    })
    writeupSubmissions.value = payload.list
    submissionPage.value = payload.page
    submissionPageSize.value = payload.page_size
    submissionTotal.value = payload.total
  } catch {
    toast.error('加载题解投稿失败')
  } finally {
    submissionLoading.value = false
  }
}

function openWriteup(mode: 'view' | 'edit') {
  if (!props.challengeId) return
  actionMenuOpen.value = false
  void router.push({
    path:
      mode === 'view'
        ? `/platform/challenges/${props.challengeId}/writeup/view`
        : `/platform/challenges/${props.challengeId}/writeup`,
  })
}

function openActionMenu() {
  actionMenuOpen.value = true
}

function closeActionMenu() {
  actionMenuOpen.value = false
}

function toggleActionMenu() {
  actionMenuOpen.value = !actionMenuOpen.value
}

function handleActionMenuFocusout(event: FocusEvent) {
  const currentTarget = event.currentTarget
  if (!(currentTarget instanceof HTMLElement)) {
    closeActionMenu()
    return
  }

  const nextTarget = event.relatedTarget
  if (nextTarget instanceof Node && currentTarget.contains(nextTarget)) {
    return
  }

  closeActionMenu()
}

async function handleDelete() {
  if (!props.challengeId || !writeup.value || deleting.value) {
    return
  }

  const confirmed = await confirmDestructiveAction({
    message: '确定删除当前题解吗？删除后学员将无法继续查看。',
  })
  if (!confirmed) {
    return
  }

  deleting.value = true
  try {
    await deleteChallengeWriteup(props.challengeId)
    writeup.value = null
    closeActionMenu()
    toast.success('题解已删除')
  } catch (error) {
    const message = error instanceof Error && error.message.trim() ? error.message : '删除题解失败'
    toast.error(message)
  } finally {
    deleting.value = false
  }
}

async function changeSubmissionPage(page: number) {
  if (
    page < 1 ||
    page === submissionPage.value ||
    submissionLoading.value ||
    !props.challengeId
  ) {
    return
  }

  await loadWriteupSubmissions(page)
}

function formatDate(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }
  return date.toLocaleString('zh-CN')
}

function submissionStatusLabel(status: TeacherSubmissionWriteupItemData['submission_status']): string {
  return status === 'draft' ? '草稿' : '已发布'
}

function visibilityStatusLabel(status: TeacherSubmissionWriteupItemData['visibility_status']): string {
  return status === 'hidden' ? '已隐藏' : '已公开'
}

function resolveAuthorName(item: TeacherSubmissionWriteupItemData): string {
  const name = item.student_name?.trim()
  return name || item.student_username
}

function resolveStudentNo(item: TeacherSubmissionWriteupItemData): string {
  const studentNo = item.student_no?.trim()
  return studentNo || '未设置学号'
}

function resolveClassName(item: TeacherSubmissionWriteupItemData): string {
  const className = item.class_name?.trim()
  return className || '未分班'
}

watch(
  () => props.challengeId,
  () => {
    void loadWriteup()
    void loadWriteupSubmissions(1)
  }
)

onMounted(() => {
  void loadWriteup()
  void loadWriteupSubmissions(1)
})
</script>

<template>
  <section class="writeup-manage-panel">
    <div class="workspace-tab-heading">
      <div class="workspace-tab-heading__main">
        <div class="journal-note-label">Writeup Directory</div>
        <h1 class="workspace-tab-heading__title">题解管理</h1>
      </div>
      <p class="workspace-tab-copy">
        {{ challengeTitle ? `查看《${challengeTitle}》的题解目录，并进入独立编辑页维护内容。` : '查看当前题目的题解目录，并进入独立编辑页维护内容。' }}
      </p>
    </div>

    <div class="writeup-manage-actions">
      <button class="admin-btn admin-btn-primary" type="button" @click="openWriteup('edit')">
        编写题解
      </button>
    </div>

    <AppLoading v-if="loading && submissionLoading" class="writeup-manage-loading">
      正在加载题解内容...
    </AppLoading>

    <template v-else>
      <section class="writeup-manage-section">
        <header class="writeup-manage-section__head">
          <div>
            <div class="journal-note-label">Official Writeup</div>
            <h2 class="writeup-manage-section__title">官方题解</h2>
          </div>
          <p class="writeup-manage-section__copy">
            {{ challengeTitle ? `维护《${challengeTitle}》的官方题解，并保留独立查看与编辑入口。` : '维护当前题目的官方题解，并保留独立查看与编辑入口。' }}
          </p>
        </header>

        <section v-if="writeup" class="writeup-directory">
          <div class="writeup-directory-head" aria-hidden="true">
            <span>题解标题</span>
            <span>可见性</span>
            <span>推荐状态</span>
            <span>更新时间</span>
            <span class="writeup-directory-head__actions">操作</span>
          </div>

          <article class="writeup-row">
            <div class="writeup-row__title">
              <div class="writeup-row__name">{{ writeup.title }}</div>
            </div>
            <div class="writeup-row__visibility">{{ writeup.visibility }}</div>
            <div class="writeup-row__recommendation">
              {{ writeup.is_recommended ? '推荐题解' : '未推荐' }}
            </div>
            <div class="writeup-row__updated">{{ formatDate(writeup.updated_at) }}</div>
            <div class="writeup-row__actions" role="group" aria-label="题解目录操作">
              <button class="admin-btn admin-btn-outline admin-btn-compact" type="button" @click="openWriteup('view')">
                查看
              </button>
              <div
                class="writeup-actions-menu-shell"
                @mouseenter="openActionMenu"
                @mouseleave="closeActionMenu"
                @focusin="openActionMenu"
                @focusout="handleActionMenuFocusout"
                @keydown.esc="closeActionMenu"
              >
                <button
                  class="admin-btn admin-btn-ghost admin-btn-compact writeup-actions-menu-trigger"
                  data-testid="writeup-more-actions"
                  type="button"
                  aria-label="更多题解操作"
                  :aria-expanded="actionMenuOpen ? 'true' : 'false'"
                  @mouseenter="openActionMenu"
                  @focus="openActionMenu"
                  @click="toggleActionMenu"
                >
                  <MoreHorizontal class="h-4 w-4" />
                </button>

                <div
                  v-if="actionMenuOpen"
                  class="writeup-actions-menu"
                  role="menu"
                  aria-label="更多题解操作"
                >
                  <button
                    class="admin-btn admin-btn-ghost admin-btn-compact writeup-actions-menu__button"
                    role="menuitem"
                    type="button"
                    @click="openWriteup('edit')"
                  >
                    编辑
                  </button>
                  <button
                    :disabled="deleting"
                    class="admin-btn admin-btn-danger admin-btn-compact writeup-actions-menu__button"
                    role="menuitem"
                    type="button"
                    @click="void handleDelete()"
                  >
                    {{ deleting ? '删除中...' : '删除' }}
                  </button>
                </div>
              </div>
            </div>
          </article>
        </section>

        <AppEmpty
          v-else
          icon="BookOpen"
          title="当前还没有题解"
          :description="challengeTitle ? `为《${challengeTitle}》创建第一份官方题解。` : '为当前题目创建第一份官方题解。'"
        >
          <template #actions>
            <button class="admin-btn admin-btn-primary" type="button" @click="openWriteup('edit')">
              编写题解
            </button>
          </template>
        </AppEmpty>
      </section>

      <section class="writeup-manage-section">
        <header class="writeup-manage-section__head">
          <div>
            <div class="journal-note-label">Submission Directory</div>
            <h2 class="writeup-manage-section__title">题解投稿</h2>
          </div>
          <div class="writeup-manage-section__meta">
            <span class="writeup-manage-section__caption">共 {{ submissionTotal }} 篇题解</span>
            <span class="writeup-manage-section__caption">第 {{ submissionPage }} / {{ submissionTotalPages }} 页</span>
          </div>
        </header>
        <p class="writeup-manage-section__copy">
          按题目查看学员投稿，补充展示作者姓名、学号、班级与当前可见状态。
        </p>

        <AppLoading v-if="submissionLoading" class="writeup-manage-loading">正在加载题解投稿...</AppLoading>

        <template v-else>
          <AppEmpty
            v-if="writeupSubmissions.length === 0"
            icon="FileText"
            title="当前还没有学员题解"
            :description="challengeTitle ? `《${challengeTitle}》暂时还没有学员提交题解。` : '当前题目暂时还没有学员提交题解。'"
          />

          <template v-else>
            <section class="submission-directory">
              <div class="submission-directory-head" aria-hidden="true">
                <span>题解标题</span>
                <span>作者</span>
                <span>学号</span>
                <span>班级</span>
                <span>状态</span>
                <span>更新时间</span>
              </div>

              <article
                v-for="item in writeupSubmissions"
                :key="item.id"
                class="submission-row"
              >
                <div class="submission-row__title">
                  <div class="submission-row__name">{{ item.title }}</div>
                  <div class="submission-row__preview">{{ item.content_preview }}</div>
                </div>
                <div class="submission-row__author">
                  <div class="submission-row__author-name">{{ resolveAuthorName(item) }}</div>
                  <div class="submission-row__author-username">@{{ item.student_username }}</div>
                </div>
                <div class="submission-row__student-no">{{ resolveStudentNo(item) }}</div>
                <div class="submission-row__class-name">{{ resolveClassName(item) }}</div>
                <div class="submission-row__status">
                  <div>{{ submissionStatusLabel(item.submission_status) }}</div>
                  <div class="submission-row__status-subtle">{{ visibilityStatusLabel(item.visibility_status) }}</div>
                </div>
                <div class="submission-row__updated">{{ formatDate(item.updated_at) }}</div>
              </article>
            </section>

            <AdminPaginationControls
              :page="submissionPage"
              :total-pages="submissionTotalPages"
              :total="submissionTotal"
              :disabled="submissionLoading"
              :total-label="`共 ${submissionTotal} 篇题解`"
              @change-page="void changeSubmissionPage($event)"
            />
          </template>
        </template>
      </section>
    </template>
  </section>
</template>

<style scoped>
.writeup-manage-panel {
  display: grid;
  gap: var(--space-6);
}

.writeup-manage-section {
  display: grid;
  gap: var(--space-4);
}

.writeup-manage-section + .writeup-manage-section {
  padding-top: var(--space-5);
  border-top: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.writeup-manage-section__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-3);
}

.writeup-manage-section__title {
  margin: 0;
  font-size: var(--font-size-1-08);
  font-weight: 700;
  color: var(--journal-ink);
}

.writeup-manage-section__copy {
  margin: 0;
  max-width: 54rem;
  font-size: var(--font-size-0-88);
  line-height: 1.75;
  color: var(--journal-muted);
}

.writeup-manage-section__meta {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: var(--space-2);
}

.writeup-manage-section__caption {
  display: inline-flex;
  align-items: center;
  min-height: 2rem;
  padding: 0 var(--space-3);
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  background: color-mix(in srgb, var(--journal-surface) 94%, var(--color-bg-base));
  font-size: var(--font-size-0-78);
  color: var(--journal-muted);
}

.writeup-manage-actions {
  display: flex;
  justify-content: flex-end;
}

.writeup-manage-loading {
  padding-block: var(--space-7);
}

.writeup-directory {
  --writeup-directory-columns: minmax(14rem, 1.6fr) minmax(7rem, 0.7fr) minmax(7rem, 0.8fr)
    minmax(11rem, 1fr) minmax(10rem, 10rem);
  display: grid;
  gap: 0;
}

.writeup-directory-head,
.writeup-row {
  display: grid;
  grid-template-columns: var(--writeup-directory-columns);
  gap: var(--space-4);
}

.writeup-directory-head {
  padding: 0 0 var(--space-3);
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  font-size: var(--font-size-0-72);
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.writeup-directory-head__actions {
  text-align: right;
}

.writeup-row {
  align-items: center;
  padding: var(--space-4) 0;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.writeup-row__title,
.writeup-row__visibility,
.writeup-row__recommendation,
.writeup-row__updated,
.writeup-row__actions {
  min-width: 0;
}

.writeup-row__name {
  font-size: var(--font-size-0-92);
  font-weight: 600;
  color: var(--journal-ink);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.writeup-row__visibility,
.writeup-row__recommendation,
.writeup-row__updated {
  font-size: var(--font-size-0-86);
  color: var(--journal-ink);
}

.writeup-row__actions {
  display: flex;
  gap: var(--space-2);
  justify-content: flex-end;
  position: relative;
  justify-self: end;
}

.submission-directory {
  --submission-directory-columns: minmax(14rem, 1.5fr) minmax(12rem, 1.05fr) minmax(9rem, 0.8fr)
    minmax(9rem, 0.8fr) minmax(8rem, 0.75fr) minmax(11rem, 1fr);
  display: grid;
  gap: 0;
}

.submission-directory-head,
.submission-row {
  display: grid;
  grid-template-columns: var(--submission-directory-columns);
  gap: var(--space-4);
}

.submission-directory-head {
  padding: 0 0 var(--space-3);
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  font-size: var(--font-size-0-72);
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.submission-row {
  align-items: center;
  padding: var(--space-4) 0;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.submission-row__title,
.submission-row__author,
.submission-row__student-no,
.submission-row__class-name,
.submission-row__status,
.submission-row__updated {
  min-width: 0;
}

.submission-row__name,
.submission-row__author-name {
  font-size: var(--font-size-0-92);
  font-weight: 600;
  color: var(--journal-ink);
}

.submission-row__preview,
.submission-row__author-username,
.submission-row__status-subtle {
  margin-top: 0.18rem;
  font-size: var(--font-size-0-78);
  color: var(--journal-muted);
}

.submission-row__preview {
  display: -webkit-box;
  overflow: hidden;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.submission-row__student-no,
.submission-row__class-name,
.submission-row__status,
.submission-row__updated {
  font-size: var(--font-size-0-86);
  color: var(--journal-ink);
}

.admin-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  min-height: 2.25rem;
  border-radius: 0.65rem;
  border: 1px solid transparent;
  padding: var(--space-2) var(--space-3-5);
  font-size: var(--font-size-0-84);
  font-weight: 600;
  transition:
    border-color 150ms ease,
    background 150ms ease,
    color 150ms ease,
    box-shadow 150ms ease;
}

.admin-btn:focus-visible {
  outline: none;
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--journal-accent) 18%, transparent);
}

.admin-btn-primary {
  border-color: color-mix(in srgb, var(--journal-accent) 18%, transparent);
  background: var(--journal-accent);
  color: #fff;
}

.admin-btn-outline,
.admin-btn-ghost {
  border-color: color-mix(in srgb, var(--journal-border) 92%, transparent);
  background: color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base));
  color: var(--journal-ink);
}

.admin-btn-danger {
  border-color: color-mix(in srgb, #ef4444 28%, transparent);
  background: color-mix(in srgb, #ef4444 16%, var(--journal-surface));
  color: #b91c1c;
}

.admin-btn-compact {
  min-height: 2.1rem;
}

.admin-btn:hover,
.admin-btn:focus-visible {
  outline: none;
  border-color: color-mix(in srgb, var(--journal-accent) 42%, transparent);
  box-shadow: 0 8px 18px color-mix(in srgb, var(--color-shadow-soft) 72%, transparent);
}

.writeup-actions-menu-shell {
  position: relative;
}

.writeup-actions-menu-trigger {
  min-width: 2.5rem;
  padding-inline: var(--space-2-5);
}

.writeup-actions-menu {
  position: absolute;
  top: calc(100% + var(--space-1-5));
  right: 0;
  z-index: 10;
  display: grid;
  gap: var(--space-2);
  min-width: 9rem;
  padding: var(--space-2-5);
  border: 1px solid color-mix(in srgb, var(--journal-border) 92%, transparent);
  border-radius: 0.9rem;
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--journal-surface) 98%, var(--color-bg-base)),
    color-mix(in srgb, var(--journal-surface-subtle) 96%, var(--color-bg-base))
  );
  box-shadow: 0 16px 32px var(--color-shadow-soft);
}

.writeup-actions-menu__button {
  justify-content: flex-start;
  width: 100%;
}

@media (max-width: 960px) {
  .writeup-manage-section__head {
    flex-direction: column;
  }

  .writeup-manage-section__meta {
    justify-content: flex-start;
  }

  .writeup-directory-head,
  .submission-directory-head {
    display: none;
  }

  .writeup-row,
  .submission-row {
    grid-template-columns: minmax(0, 1fr);
    gap: var(--space-2);
    align-items: start;
  }

  .writeup-row__actions {
    justify-content: flex-start;
    margin-top: var(--space-2);
  }

  .writeup-actions-menu {
    right: auto;
    left: 0;
  }
}
</style>
