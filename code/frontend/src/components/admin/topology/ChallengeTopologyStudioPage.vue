<script setup lang="ts">
import { computed, defineAsyncComponent, ref } from 'vue'
import { Blocks, GitBranch, Link2, Plus, RefreshCw, Save, ShieldBan, Trash2, Layout, Server, Network, ShieldCheck } from 'lucide-vue-next'

import {
  useChallengeTopologyStudioPage,
  type TopologyStudioMode,
} from '@/composables/useChallengeTopologyStudioPage'
import AppCard from '@/components/common/AppCard.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import SectionCard from '@/components/common/SectionCard.vue'

import type { CanvasInteractionMode } from './TopologyCanvasBoard.vue'

const TopologyCanvasBoard = defineAsyncComponent(() => import('./TopologyCanvasBoard.vue'))
const TopologyNodeEditor = defineAsyncComponent(() => import('./TopologyNodeEditor.vue'))

const props = withDefaults(
  defineProps<{
    challengeId?: string
    mode?: TopologyStudioMode
  }>(),
  {
    challengeId: '',
    mode: 'challenge',
  }
)

const emit = defineEmits<{
  back: []
}>()

const activeWorkbenchTab = ref<'visual' | 'compute' | 'network' | 'policy'>('visual')

const {
  loading,
  saving,
  templateBusy,
  challenge,
  topology,
  images,
  templates,
  templateKeyword,
  selectedTemplateId,
  templateName,
  templateDescription,
  draft,
  selectedNodeKey,
  selectedEdgeId,
  interactionMode,
  pendingSourceNodeKey,
  nodePositions,
  isTemplateLibraryMode,
  nodeOptions,
  canSaveTemplate,
  selectedTemplate,
  pageHeader,
  loadingText,
  heroEyebrow,
  heroTitle,
  heroDescription,
  statusCard,
  secondaryCard,
  selectedCanvasSummary,
  draftValidationIssues,
  selectedTemplateSummary,
  topologySummary,
  canvasGraph,
  selectedNodeDraft,
  selectedEdgeMeta,
  selectedLinkDraft,
  selectedPolicyDraft,
  selectedEdgeSourceKey,
  selectedEdgeTargetKey,
  selectedEdgeKind,
  canvasModeLabel,
  updateCanvasQuickNumber,
  toggleSelectedNodeNetwork,
  updateSelectedEdgeSourceKey,
  updateSelectedEdgeTargetKey,
  handleSelectedEdgeKindChange,
  reloadAll,
  handleResetTemplateEditor,
  addNetwork,
  removeNetwork,
  addNode,
  removeNode,
  updateNodePosition,
  setInteractionMode,
  handleCanvasSelectNode,
  handleCanvasSelectEdge,
  handleCanvasCreateNode,
  handleCanvasCreateEdge,
  removeSelectedCanvasItem,
  addLink,
  addPolicy,
  loadTemplateIntoDraft,
  handleApplyTemplate,
  handleSaveTopology,
  handleDeleteTopology,
  handleCreateTemplate,
  handleUpdateTemplate,
  handleDeleteTemplate,
  clearTemplateSelection,
  loadTemplates,
  resetTemplateForm,
} = useChallengeTopologyStudioPage({
  challengeId: props.challengeId,
  mode: props.mode,
})
</script>

<template>
  <div
    :class="[
      'topology-page',
      isTemplateLibraryMode
        ? 'topology-page--template-library'
        : 'topology-page--challenge workspace-shell journal-shell journal-shell-admin journal-notes-card journal-hero teacher-management-shell teacher-surface teacher-surface-workspace-bg',
    ]"
  >
    <PageHeader
      v-if="isTemplateLibraryMode"
      class="topology-page-header"
      :eyebrow="pageHeader.eyebrow"
      :title="pageHeader.title"
      :description="pageHeader.description"
    >
      <button
        v-if="!isTemplateLibraryMode"
        type="button"
        :class="
          isTemplateLibraryMode
            ? 'topology-toolbar-btn topology-toolbar-btn--ghost'
            : 'inline-flex items-center gap-2 rounded-xl border border-border px-4 py-2 text-sm font-medium text-text-primary transition hover:border-primary'
        "
        @click="emit('back')"
      >
        返回题目详情
      </button>
      <button
        v-else
        type="button"
        :class="
          isTemplateLibraryMode
            ? 'topology-toolbar-btn topology-toolbar-btn--ghost'
            : 'inline-flex items-center gap-2 rounded-xl border border-border px-4 py-2 text-sm font-medium text-text-primary transition hover:border-primary'
        "
        @click="handleResetTemplateEditor"
      >
        <Plus class="h-4 w-4" />
        新建空白模板
      </button>
      <button
        type="button"
        :class="
          isTemplateLibraryMode
            ? 'topology-toolbar-btn topology-toolbar-btn--ghost'
            : 'inline-flex items-center gap-2 rounded-xl border border-border px-4 py-2 text-sm font-medium text-text-primary transition hover:border-primary'
        "
        @click="void reloadAll()"
      >
        <RefreshCw class="h-4 w-4" />
        刷新
      </button>
      <button
        v-if="!isTemplateLibraryMode"
        type="button"
        :class="
          isTemplateLibraryMode
            ? 'topology-toolbar-btn topology-toolbar-btn--primary'
            : 'inline-flex items-center gap-2 rounded-xl bg-primary px-4 py-2 text-sm font-medium text-white transition hover:opacity-90'
        "
        :disabled="saving"
        @click="void handleSaveTopology()"
      >
        <Save class="h-4 w-4" />
        {{ saving ? '保存中...' : '保存拓扑' }}
      </button>
    </PageHeader>

    <header
      v-else
      class="workspace-topbar topology-workspace-topbar"
    >
      <div class="topology-topbar-leading">
        <span class="workspace-overline">Challenge Workspace</span>
        <span class="topology-topbar-chip">{{ pageHeader.eyebrow }}</span>
      </div>
      <div class="topology-topbar-actions">
        <button
          type="button"
          class="topology-toolbar-btn topology-toolbar-btn--ghost"
          @click="emit('back')"
        >
          返回题目详情
        </button>
        <button
          type="button"
          class="topology-toolbar-btn topology-toolbar-btn--ghost"
          @click="void reloadAll()"
        >
          <RefreshCw class="h-4 w-4" />
          刷新
        </button>
        <button
          type="button"
          class="topology-toolbar-btn topology-toolbar-btn--primary"
          :disabled="saving"
          @click="void handleSaveTopology()"
        >
          <Save class="h-4 w-4" />
          {{ saving ? '保存中...' : '保存拓扑' }}
        </button>
      </div>
    </header>

    <div
      v-if="loading && !isTemplateLibraryMode"
      class="content-pane topology-loading-pane"
    >
      <AppLoading>{{ loadingText }}</AppLoading>
    </div>

    <section
      v-else-if="loading && isTemplateLibraryMode"
      class="template-library-main rounded-[30px] border px-6 py-6 md:px-8"
    >
      <div class="flex justify-center py-12">
        <AppLoading>{{ loadingText }}</AppLoading>
      </div>
    </section>

    <template v-else>
      <section
        v-if="isTemplateLibraryMode"
        class="template-library-main rounded-[30px] border px-6 py-6 md:px-8"
      >
        <section class="topology-hero-grid grid gap-4 xl:grid-cols-[1.04fr_0.96fr]">
          <div class="topology-hero-lead topology-hero-lead--library">
            <div class="topology-hero-kicker">
              <span>{{ heroEyebrow }}</span>
              <span class="topology-hero-badge">真实接口</span>
            </div>
            <h1 class="topology-hero-title">
              {{ heroTitle }}
            </h1>
            <p class="topology-hero-description">
              {{ heroDescription }}
            </p>

            <div class="topology-summary-grid progress-strip metric-panel-grid metric-panel-default-surface">
              <div class="topology-summary-tile progress-card metric-panel-card">
                <div class="topology-summary-label progress-card-label metric-panel-label">网络</div>
                <div class="topology-summary-value progress-card-value metric-panel-value">
                  {{ topologySummary.networks }}
                </div>
                <div class="topology-summary-helper progress-card-hint metric-panel-helper">
                  当前模板草稿中的网络数量
                </div>
              </div>
              <div class="topology-summary-tile progress-card metric-panel-card">
                <div class="topology-summary-label progress-card-label metric-panel-label">节点</div>
                <div class="topology-summary-value progress-card-value metric-panel-value">
                  {{ topologySummary.nodes }}
                </div>
                <div class="topology-summary-helper progress-card-hint metric-panel-helper">
                  当前模板草稿中的节点数量
                </div>
              </div>
              <div class="topology-summary-tile progress-card metric-panel-card">
                <div class="topology-summary-label progress-card-label metric-panel-label">连线</div>
                <div class="topology-summary-value progress-card-value metric-panel-value">
                  {{ topologySummary.links }}
                </div>
                <div class="topology-summary-helper progress-card-hint metric-panel-helper">
                  当前模板草稿中的连线数量
                </div>
              </div>
              <div class="topology-summary-tile progress-card metric-panel-card">
                <div class="topology-summary-label progress-card-label metric-panel-label">策略</div>
                <div class="topology-summary-value progress-card-value metric-panel-value">
                  {{ topologySummary.policies }}
                </div>
                <div class="topology-summary-helper progress-card-hint metric-panel-helper">
                  当前模板草稿中的策略数量
                </div>
              </div>
            </div>
          </div>

          <div class="topology-hero-aside topology-hero-aside--library grid gap-3 md:grid-cols-3 xl:grid-cols-1">
            <section class="template-hero-note template-hero-note--primary">
              <div class="template-metric-icon template-metric-icon--primary">
                <Blocks class="h-5 w-5" />
              </div>
              <div class="template-hero-note__body">
                <div class="template-hero-note__label">{{ statusCard.eyebrow }}</div>
                <div class="template-hero-note__value">{{ statusCard.title }}</div>
                <p class="template-hero-note__copy">{{ statusCard.subtitle }}</p>
              </div>
            </section>

            <section class="template-hero-note template-hero-note--warning">
              <div class="template-metric-icon template-metric-icon--warning">
                <GitBranch class="h-5 w-5" />
              </div>
              <div class="template-hero-note__body">
                <div class="template-hero-note__label">{{ secondaryCard.eyebrow }}</div>
                <div class="template-hero-note__value">{{ secondaryCard.title }}</div>
                <p class="template-hero-note__copy">{{ secondaryCard.subtitle }}</p>
              </div>
            </section>
          </div>
        </section>

        <div class="template-library-divider" />

        <section class="topology-workbench grid gap-6 xl:grid-cols-[1fr_360px]">
          <div class="space-y-6">
            <div class="flex items-center gap-2">
              <div class="template-toolbar-tabs">
                <button
                  v-for="tab in [
                    { id: 'visual', label: '画布', icon: Layout },
                    { id: 'compute', label: '节点', icon: Server },
                    { id: 'network', label: '网络', icon: Network },
                    { id: 'policy', label: '策略', icon: ShieldCheck },
                  ]"
                  :key="tab.id"
                  type="button"
                  class="template-toolbar-tab"
                  :class="
                    activeWorkbenchTab === tab.id
                      ? 'template-toolbar-tab--active'
                      : 'template-toolbar-tab--idle'
                  "
                  @click="activeWorkbenchTab = (tab.id as any)"
                >
                  <component
                    :is="tab.icon"
                    class="h-4 w-4"
                  />
                  <span class="hidden sm:inline">{{ tab.label }}</span>
                </button>
              </div>
              <button
                type="button"
                class="template-toolbar-refresh"
                title="刷新数据"
                @click="void reloadAll()"
              >
                <RefreshCw class="h-4 w-4" />
              </button>
            </div>

            <div
              v-if="activeWorkbenchTab === 'visual'"
              class="space-y-6"
            >
              <SectionCard
                title="图形画布"
                subtitle="拖拽节点调整视图布局，点击节点可快速跳到对应节点编辑卡片。"
              >
                <div class="mb-4 flex flex-wrap items-center gap-2">
                  <button
                    type="button"
                    class="rounded-xl border px-3 py-2 text-sm font-medium transition"
                    :class="
                      interactionMode === 'pan'
                        ? 'border-primary bg-primary/10 text-primary'
                        : 'border-border text-text-primary hover:border-primary'
                    "
                    @click="setInteractionMode('pan')"
                  >
                    浏览
                  </button>
                  <button
                    type="button"
                    class="rounded-xl border px-3 py-2 text-sm font-medium transition"
                    :class="
                      interactionMode === 'add-node'
                        ? 'border-primary bg-primary/10 text-primary'
                        : 'border-border text-text-primary hover:border-primary'
                    "
                    @click="setInteractionMode('add-node')"
                  >
                    新增节点
                  </button>
                  <button
                    type="button"
                    class="rounded-xl border px-3 py-2 text-sm font-medium transition"
                    :class="
                      interactionMode === 'link'
                        ? 'border-primary bg-primary/10 text-primary'
                        : 'border-border text-text-primary hover:border-primary'
                    "
                    @click="setInteractionMode('link')"
                  >
                    连线模式
                  </button>
                  <button
                    type="button"
                    class="rounded-xl border px-3 py-2 text-sm font-medium transition"
                    :class="
                      interactionMode === 'allow'
                        ? 'border-[var(--color-success)] bg-[var(--color-success)]/10 text-[var(--color-success)]'
                        : 'border-border text-text-primary hover:border-[var(--color-success)]/60'
                    "
                    @click="setInteractionMode('allow')"
                  >
                    allow
                  </button>
                  <button
                    type="button"
                    class="rounded-xl border px-3 py-2 text-sm font-medium transition"
                    :class="
                      interactionMode === 'deny'
                        ? 'border-danger bg-danger/10 text-danger'
                        : 'border-border text-text-primary hover:border-danger/60'
                    "
                    @click="setInteractionMode('deny')"
                  >
                    deny
                  </button>
                  <button
                    type="button"
                    class="rounded-xl border border-danger/30 bg-danger/10 px-3 py-2 text-sm font-medium text-danger transition hover:bg-danger/15"
                    @click="removeSelectedCanvasItem"
                  >
                    删除选中
                  </button>
                </div>

                <div
                  class="template-canvas-mode-banner mb-4 rounded-2xl border border-border bg-elevated px-4 py-3 text-sm text-text-secondary"
                >
                  <div class="flex flex-wrap items-center gap-2">
                    <span
                      class="rounded-full border border-primary/20 bg-primary/10 px-2.5 py-1 text-xs text-primary"
                    >
                      当前模式：{{ canvasModeLabel }}
                    </span>
                    <span
                      class="rounded-full border border-border-subtle bg-surface px-2.5 py-1 text-xs text-text-secondary"
                    >
                      {{ selectedCanvasSummary }}
                    </span>
                    <span
                      class="rounded-full border border-border-subtle bg-surface px-2.5 py-1 text-xs text-text-muted"
                    >
                      `Esc` 取消 / `Delete` 删除
                    </span>
                  </div>
                </div>

                <div
                  class="mb-4 rounded-2xl border px-4 py-3 text-sm"
                  :class="
                    draftValidationIssues.length === 0
                      ? 'border-[var(--color-success)]/20 bg-[var(--color-success)]/10 text-[var(--color-success)]'
                      : 'border-[var(--color-warning)]/20 bg-[var(--color-warning)]/10 text-[var(--color-warning)]'
                  "
                >
                  <div class="font-medium">
                    {{ draftValidationIssues.length === 0 ? '基础校验已通过' : '基础校验发现问题' }}
                  </div>
                  <ul
                    v-if="draftValidationIssues.length > 0"
                    class="mt-2 space-y-1 text-xs"
                  >
                    <li
                      v-for="issue in draftValidationIssues"
                      :key="issue"
                    >
                      {{ issue }}
                    </li>
                  </ul>
                </div>

                <TopologyCanvasBoard
                  :graph="canvasGraph"
                  :interaction-mode="interactionMode"
                  :pending-source-node-key="pendingSourceNodeKey"
                  :selected-node-key="selectedNodeKey"
                  :selected-edge-id="selectedEdgeId"
                  @select-node="handleCanvasSelectNode"
                  @select-edge="handleCanvasSelectEdge"
                  @create-node-at="handleCanvasCreateNode"
                  @create-edge="handleCanvasCreateEdge"
                  @clear-pending="pendingSourceNodeKey = null"
                  @update-position="updateNodePosition"
                />

                <div class="mt-4 grid gap-4">
                  <div class="template-quick-editor rounded-2xl border border-border bg-elevated p-4">
                    <div class="text-sm font-semibold text-text-primary">
                      画布快速编辑
                    </div>

                    <div
                      v-if="!selectedNodeDraft && !selectedEdgeMeta"
                      class="mt-3 rounded-xl border border-dashed border-border px-4 py-6 text-sm text-text-muted"
                    >
                      请在画布中选择一个节点或连线进行快速配置
                    </div>

                    <div
                      v-else-if="selectedNodeDraft"
                      class="mt-3 space-y-4"
                    >
                      <div class="grid gap-3 md:grid-cols-2">
                        <label class="space-y-2">
                          <span class="text-sm text-text-secondary">节点名称</span>
                          <input
                            v-model="selectedNodeDraft.name"
                            type="text"
                            class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
                          >
                        </label>
                        <label class="space-y-2">
                          <span class="text-sm text-text-secondary">服务端口</span>
                          <input
                            :value="selectedNodeDraft.service_port ?? ''"
                            type="number"
                            min="1"
                            max="65535"
                            class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
                            @input="
                              updateCanvasQuickNumber(
                                'service_port',
                                ($event.target as HTMLInputElement).value,
                                selectedNodeDraft
                              )
                            "
                          >
                        </label>
                      </div>

                      <div class="space-y-2">
                        <div class="text-sm text-text-secondary">
                          所属网络
                        </div>
                        <div class="flex flex-wrap gap-2">
                          <label
                            v-for="network in draft.networks"
                            :key="network.uid"
                            class="flex items-center gap-2 rounded-xl border border-border bg-surface px-3 py-2 text-sm text-text-primary transition hover:border-primary"
                          >
                            <input
                              :checked="selectedNodeDraft.network_keys.includes(network.key)"
                              type="checkbox"
                              class="h-4 w-4 rounded border-border bg-transparent"
                              @change="
                                toggleSelectedNodeNetwork(
                                  network.key,
                                  ($event.target as HTMLInputElement).checked
                                )
                              "
                            >
                            <span>{{ network.name || network.key }}</span>
                          </label>
                        </div>
                      </div>
                    </div>

                    <div
                      v-else-if="selectedEdgeMeta"
                      class="mt-3 space-y-4"
                    >
                      <div class="grid gap-3 md:grid-cols-2">
                        <label class="space-y-2">
                          <span class="text-sm text-text-secondary">源节点</span>
                          <select
                            :value="selectedEdgeSourceKey"
                            class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
                            @change="
                              updateSelectedEdgeSourceKey(($event.target as HTMLSelectElement).value)
                            "
                          >
                            <option
                              v-for="node in nodeOptions"
                              :key="node.key"
                              :value="node.key"
                            >
                              {{ node.label }}
                            </option>
                          </select>
                        </label>
                        <label class="space-y-2">
                          <span class="text-sm text-text-secondary">目标节点</span>
                          <select
                            :value="selectedEdgeTargetKey"
                            class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
                            @change="
                              updateSelectedEdgeTargetKey(($event.target as HTMLSelectElement).value)
                            "
                          >
                            <option
                              v-for="node in nodeOptions"
                              :key="node.key"
                              :value="node.key"
                            >
                              {{ node.label }}
                            </option>
                          </select>
                        </label>
                      </div>

                      <label class="space-y-2">
                        <span class="text-sm text-text-secondary">边类型</span>
                        <select
                          :value="selectedEdgeKind"
                          class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
                          @change="
                            handleSelectedEdgeKindChange(($event.target as HTMLSelectElement).value)
                          "
                        >
                          <option value="link">logic link</option>
                          <option value="allow">allow</option>
                          <option value="deny">deny</option>
                        </select>
                      </label>
                    </div>
                  </div>
                </div>
              </SectionCard>
            </div>

            <div
              v-else-if="activeWorkbenchTab === 'compute'"
              class="space-y-6"
            >
              <SectionCard
                title="入口节点"
                subtitle="实例访问入口和当前草稿的保存范围。"
              >
                <div class="grid gap-4">
                  <label class="space-y-2">
                    <span class="text-sm text-text-secondary">入口节点</span>
                    <select
                      v-model="draft.entry_node_key"
                      class="w-full rounded-xl border border-border bg-elevated px-3 py-3 text-sm text-text-primary outline-none transition focus:border-primary"
                    >
                      <option
                        v-for="node in nodeOptions"
                        :key="node.key"
                        :value="node.key"
                      >
                        {{ node.label }} ({{ node.key }})
                      </option>
                    </select>
                  </label>
                </div>
              </SectionCard>

              <SectionCard
                title="节点编排"
                subtitle="节点支持单独镜像、资源限制、网络归属和环境变量。"
              >
                <div class="space-y-4">
                  <TopologyNodeEditor
                    v-for="(node, index) in draft.nodes"
                    :key="node.uid"
                    :data-node-editor="node.key"
                    :model-value="node"
                    :index="index"
                    :images="images"
                    :networks="draft.networks"
                    :removable="draft.nodes.length > 1"
                    :selected="selectedNodeKey === node.key"
                    @update:model-value="draft.nodes[index] = $event"
                    @remove="removeNode(node.uid)"
                  />
                </div>

                <template #footer>
                  <button
                    type="button"
                    class="inline-flex items-center gap-2 rounded-xl border border-border px-4 py-2 text-sm font-medium text-text-primary transition hover:border-primary"
                    @click="addNode"
                  >
                    <Plus class="h-4 w-4" />
                    添加节点
                  </button>
                </template>
              </SectionCard>
            </div>

            <div
              v-else-if="activeWorkbenchTab === 'network'"
              class="space-y-6"
            >
              <SectionCard
                title="网络分段"
                subtitle="一个节点可以挂多个网络，运行时会创建多个 Docker Network。"
              >
                <div class="space-y-3">
                  <div
                    v-for="network in draft.networks"
                    :key="network.uid"
                    class="grid gap-3 rounded-2xl border border-border bg-elevated p-4 md:grid-cols-[0.9fr_1fr_0.9fr_auto_auto]"
                  >
                    <input
                      v-model="network.key"
                      type="text"
                      class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
                      placeholder="network key"
                    >
                    <input
                      v-model="network.name"
                      type="text"
                      class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
                      placeholder="网络名称"
                    >
                    <input
                      v-model="network.cidr"
                      type="text"
                      class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
                      placeholder="CIDR（可选）"
                    >
                    <label
                      class="flex items-center gap-3 rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary"
                    >
                      <input
                        v-model="network.internal"
                        type="checkbox"
                        class="h-4 w-4 rounded border-border bg-transparent"
                      >
                      internal
                    </label>
                    <button
                      type="button"
                      class="inline-flex items-center justify-center rounded-xl border border-danger/30 bg-danger/10 px-3 py-2 text-sm text-danger transition hover:bg-danger/15"
                      :disabled="draft.networks.length <= 1"
                      @click="removeNetwork(network.uid)"
                    >
                      <Trash2 class="h-4 w-4" />
                    </button>
                  </div>
                </div>

                <template #footer>
                  <button
                    type="button"
                    class="inline-flex items-center gap-2 rounded-xl border border-border px-4 py-2 text-sm font-medium text-text-primary transition hover:border-primary"
                    @click="addNetwork"
                  >
                    <Plus class="h-4 w-4" />
                    添加网络
                  </button>
                </template>
              </SectionCard>
            </div>

            <div
              v-else-if="activeWorkbenchTab === 'policy'"
              class="space-y-6"
            >
              <SectionCard
                title="拓扑连线"
                subtitle="用于表达逻辑依赖关系，不直接等同于运行时 ACL。"
              >
                <div
                  v-if="draft.links.length === 0"
                  class="rounded-xl border border-dashed border-border px-4 py-6 text-sm text-text-muted"
                >
                  暂无逻辑连线
                </div>
                <div
                  v-else
                  class="space-y-3"
                >
                  <div
                    v-for="link in draft.links"
                    :key="link.uid"
                    class="grid gap-3 rounded-2xl border border-border bg-elevated p-4 md:grid-cols-[1fr_1fr_auto]"
                  >
                    <select
                      v-model="link.from_node_key"
                      class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
                    >
                      <option value="">
                        选择源节点
                      </option>
                      <option
                        v-for="node in nodeOptions"
                        :key="node.key"
                        :value="node.key"
                      >
                        {{ node.label }}
                      </option>
                    </select>
                    <select
                      v-model="link.to_node_key"
                      class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
                    >
                      <option value="">
                        选择目标节点
                      </option>
                      <option
                        v-for="node in nodeOptions"
                        :key="node.key"
                        :value="node.key"
                      >
                        {{ node.label }}
                      </option>
                    </select>
                    <button
                      type="button"
                      class="inline-flex items-center justify-center rounded-xl border border-danger/30 bg-danger/10 px-3 py-2 text-sm text-danger transition hover:bg-danger/15"
                      @click="draft.links = draft.links.filter((item) => item.uid !== link.uid)"
                    >
                      <Trash2 class="h-4 w-4" />
                    </button>
                  </div>
                </div>

                <template #footer>
                  <button
                    type="button"
                    class="inline-flex items-center gap-2 rounded-xl border border-border px-4 py-2 text-sm font-medium text-text-primary transition hover:border-primary"
                    @click="addLink"
                  >
                    <Link2 class="h-4 w-4" />
                    添加连线
                  </button>
                </template>
              </SectionCard>

              <SectionCard
                title="链路策略"
                subtitle="当前前端只开放粗粒度节点 allow/deny，细粒度端口策略尚未支持。"
              >
                <div
                  v-if="draft.policies.length === 0"
                  class="rounded-xl border border-dashed border-border px-4 py-6 text-sm text-text-muted"
                >
                  暂无链路策略
                </div>
                <div
                  v-else
                  class="space-y-3"
                >
                  <div
                    v-for="policy in draft.policies"
                    :key="policy.uid"
                    class="grid gap-3 rounded-2xl border border-border bg-elevated p-4 md:grid-cols-[1fr_1fr_0.7fr_auto]"
                  >
                    <select
                      v-model="policy.source_node_key"
                      class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
                    >
                      <option value="">
                        选择源节点
                      </option>
                      <option
                        v-for="node in nodeOptions"
                        :key="node.key"
                        :value="node.key"
                      >
                        {{ node.label }}
                      </option>
                    </select>
                    <select
                      v-model="policy.target_node_key"
                      class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
                    >
                      <option value="">
                        选择目标节点
                      </option>
                      <option
                        v-for="node in nodeOptions"
                        :key="node.key"
                        :value="node.key"
                      >
                        {{ node.label }}
                      </option>
                    </select>
                    <select
                      v-model="policy.action"
                      class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
                    >
                      <option value="allow">
                        allow
                      </option>
                      <option value="deny">
                        deny
                      </option>
                    </select>
                    <button
                      type="button"
                      class="inline-flex items-center justify-center rounded-xl border border-danger/30 bg-danger/10 px-3 py-2 text-sm text-danger transition hover:bg-danger/15"
                      @click="draft.policies = draft.policies.filter((item) => item.uid !== policy.uid)"
                    >
                      <Trash2 class="h-4 w-4" />
                    </button>
                  </div>
                </div>

                <template #footer>
                  <button
                    type="button"
                    class="inline-flex items-center gap-2 rounded-xl border border-border px-4 py-2 text-sm font-medium text-text-primary transition hover:border-primary"
                    @click="addPolicy"
                  >
                    <ShieldBan class="h-4 w-4" />
                    添加策略
                  </button>
                </template>
              </SectionCard>
            </div>
          </div>

          <div class="topology-side-stack topology-side-stack--library">
            <SectionCard
              title="模板库"
              :subtitle="
                isTemplateLibraryMode
                  ? '从模板库载入后可直接编辑并覆盖模板，或另存为新模板。'
                  : '可按模板快速回填编辑器，或直接应用到题目。'
              "
            >
              <div class="space-y-3">
                <div class="template-focus-card">
                  <div class="text-xs font-semibold uppercase tracking-[0.22em] text-text-muted">
                    当前模板
                  </div>
                  <div class="mt-2 text-sm text-text-primary">
                    {{ selectedTemplateSummary }}
                  </div>
                  <div class="mt-3 flex flex-wrap gap-2">
                    <button
                      v-if="selectedTemplate"
                      type="button"
                      class="template-action-btn"
                      @click="loadTemplateIntoDraft(selectedTemplate)"
                    >
                      重新载入当前模板
                    </button>
                    <button
                      v-if="selectedTemplate"
                      type="button"
                      class="template-action-btn"
                      @click="clearTemplateSelection"
                    >
                      清空模板选择
                    </button>
                  </div>
                </div>

                <div class="template-search-row">
                  <input
                    v-model="templateKeyword"
                    type="text"
                    class="w-full rounded-xl border border-border bg-elevated px-3 py-3 text-sm text-text-primary outline-none transition focus:border-primary"
                    placeholder="按模板名称搜索"
                  >
                  <button
                    type="button"
                    class="template-action-btn"
                    @click="void loadTemplates()"
                  >
                    搜索
                  </button>
                </div>

                <div
                  v-if="templates.length === 0"
                  class="template-empty-state"
                >
                  当前没有模板数据
                </div>

              <div
                v-else
                class="template-library-list"
              >
                <div class="template-directory-head" aria-hidden="true">
                  <span>模板</span>
                  <span>概况</span>
                  <span>操作</span>
                </div>

                <article
                  v-for="template in templates"
                  :key="template.id"
                    :class="[
                      'template-library-item',
                      selectedTemplateId === template.id
                        ? 'template-library-item--active'
                        : 'template-library-item--idle',
                    ]"
                  >
                    <div class="template-library-item__main">
                      <div class="truncate text-base font-bold text-text-primary">
                        {{ template.name }}
                      </div>
                      <div class="mt-1 line-clamp-2 text-xs leading-relaxed text-text-secondary">
                        {{ template.description || '无描述' }}
                      </div>
                      <div class="mt-3 flex flex-wrap gap-x-3 gap-y-1.5 text-[10px] font-bold uppercase tracking-wider text-text-muted">
                        <span class="flex items-center gap-1">
                          <Layout class="h-3 w-3" />
                          {{ template.entry_node_key }}
                        </span>
                        <span class="flex items-center gap-1">
                          <Server class="h-3 w-3" />
                          {{ template.nodes.length }}
                        </span>
                        <span class="flex items-center gap-1">
                          <Network class="h-3 w-3" />
                          {{ template.networks?.length || 0 }}
                        </span>
                        <span class="flex items-center gap-1">
                          <RefreshCw class="h-3 w-3" />
                          {{ template.usage_count }}
                        </span>
                      </div>
                    </div>

                    <div class="template-library-item__meta">
                      <span>入口 {{ template.entry_node_key }}</span>
                      <span>{{ template.nodes.length }} 节点</span>
                      <span>{{ template.networks?.length || 0 }} 网络</span>
                      <span>使用 {{ template.usage_count }}</span>
                    </div>

                    <div class="template-library-item__actions">
                      <button
                        type="button"
                        class="template-action-btn"
                        @click="loadTemplateIntoDraft(template)"
                      >
                        载入编辑
                      </button>
                      <button
                        type="button"
                        class="template-action-btn template-action-btn--danger"
                        :disabled="templateBusy"
                        @click="void handleDeleteTemplate(template.id)"
                      >
                        <Trash2 class="h-3 w-3" />
                      </button>
                    </div>
                  </article>
                </div>
              </div>
            </SectionCard>

            <SectionCard
              title="模板写回"
              subtitle="在独立模板库中可新建空白草稿、载入现有模板后覆盖，或另存为新模板。"
            >
              <div class="template-writeback-form">
                <label class="space-y-2">
                  <span class="text-sm text-text-secondary">模板名称</span>
                  <input
                    v-model="templateName"
                    type="text"
                    class="w-full rounded-xl border border-border bg-elevated px-3 py-3 text-sm text-text-primary outline-none transition focus:border-primary"
                    placeholder="例如 双节点 Web + DB"
                  >
                </label>

                <label class="space-y-2">
                  <span class="text-sm text-text-secondary">模板描述</span>
                  <textarea
                    v-model="templateDescription"
                    rows="4"
                    class="w-full rounded-xl border border-border bg-elevated px-3 py-3 text-sm text-text-primary outline-none transition focus:border-primary"
                    placeholder="说明这个模板的适用场景"
                  />
                </label>

                <div class="flex flex-wrap gap-2">
                  <button
                    type="button"
                    class="template-action-btn flex-1"
                    @click="handleResetTemplateEditor"
                  >
                    新建空白模板
                  </button>
                  <button
                    type="button"
                    class="template-action-btn template-action-btn--primary flex-1"
                    :disabled="templateBusy"
                    @click="void handleCreateTemplate()"
                  >
                    <Plus class="h-4 w-4" />
                    另存为
                  </button>
                  <button
                    type="button"
                    class="template-action-btn flex-1"
                    :disabled="templateBusy || !selectedTemplateId"
                    @click="void handleUpdateTemplate()"
                  >
                    覆盖
                  </button>
                </div>
              </div>
            </SectionCard>

            <SectionCard
              title="当前边界"
              subtitle="避免把未生效能力继续暴露成可用配置。"
            >
              <div class="template-boundary-list">
                <article class="template-boundary-item template-boundary-item--warning">
                  <div class="template-boundary-item__label">已开放</div>
                  <div class="template-boundary-item__copy">
                    多网络、节点、逻辑连线、粗粒度 allow/deny 策略、模板复用。
                  </div>
                </article>
                <article class="template-boundary-item template-boundary-item--danger">
                  <div class="template-boundary-item__label">暂未开放</div>
                  <div class="template-boundary-item__copy">
                    protocol / ports 级细粒度 ACL 前端字段、模板版本化与批量比对能力。
                  </div>
                </article>
                <article class="template-boundary-item template-boundary-item--neutral">
                  <div class="template-boundary-item__label">建议</div>
                  <div class="template-boundary-item__copy">
                    继续开放高级能力前，先补参数校验、可视化提示和误操作保护。
                  </div>
                </article>
              </div>
            </SectionCard>
          </div>
        </section>
      </section>

      <template v-else>
        <section class="workspace-tab-heading topology-page-heading">
          <div class="workspace-tab-heading__main">
            <div class="topology-page-kicker">{{ heroEyebrow }}</div>
            <h1 class="hero-title">{{ heroTitle }}</h1>
          </div>
          <p class="workspace-page-copy topology-page-copy">
            {{ heroDescription }}
          </p>

          <div class="topology-summary-grid topology-summary-grid--challenge metric-panel-grid">
            <article class="topology-summary-card metric-panel-card">
              <div class="topology-summary-label metric-panel-label">网络</div>
              <div class="topology-summary-value metric-panel-value">{{ topologySummary.networks }}</div>
            </article>
            <article class="topology-summary-card metric-panel-card">
              <div class="topology-summary-label metric-panel-label">节点</div>
              <div class="topology-summary-value metric-panel-value">{{ topologySummary.nodes }}</div>
            </article>
            <article class="topology-summary-card metric-panel-card">
              <div class="topology-summary-label metric-panel-label">连线</div>
              <div class="topology-summary-value metric-panel-value">{{ topologySummary.links }}</div>
            </article>
            <article class="topology-summary-card metric-panel-card">
              <div class="topology-summary-label metric-panel-label">策略</div>
              <div class="topology-summary-value metric-panel-value">{{ topologySummary.policies }}</div>
            </article>
          </div>
        </section>

        <div class="journal-divider" />

        <main class="content-pane topology-workspace">
        <div class="topology-primary-column">
          <SectionCard
            title="图形画布"
            subtitle="拖拽节点调整视图布局，点击节点可快速跳到对应节点编辑卡片。"
          >
            <div class="mb-4 flex flex-wrap items-center gap-2">
              <button
                type="button"
                class="rounded-xl border px-3 py-2 text-sm font-medium transition"
                :class="
                  interactionMode === 'pan'
                    ? 'border-primary bg-primary/10 text-primary'
                    : 'border-border text-text-primary hover:border-primary'
                "
                @click="setInteractionMode('pan')"
              >
                浏览
              </button>
              <button
                type="button"
                class="rounded-xl border px-3 py-2 text-sm font-medium transition"
                :class="
                  interactionMode === 'add-node'
                    ? 'border-primary bg-primary/10 text-primary'
                    : 'border-border text-text-primary hover:border-primary'
                "
                @click="setInteractionMode('add-node')"
              >
                画布新增节点
              </button>
              <button
                type="button"
                class="rounded-xl border px-3 py-2 text-sm font-medium transition"
                :class="
                  interactionMode === 'link'
                    ? 'border-primary bg-primary/10 text-primary'
                    : 'border-border text-text-primary hover:border-primary'
                "
                @click="setInteractionMode('link')"
              >
                连线模式
              </button>
              <button
                type="button"
                class="rounded-xl border px-3 py-2 text-sm font-medium transition"
                :class="
                  interactionMode === 'allow'
                    ? 'border-[var(--color-success)] bg-[var(--color-success)]/10 text-[var(--color-success)]'
                    : 'border-border text-text-primary hover:border-[var(--color-success)]/60'
                "
                @click="setInteractionMode('allow')"
              >
                allow 模式
              </button>
              <button
                type="button"
                class="rounded-xl border px-3 py-2 text-sm font-medium transition"
                :class="
                  interactionMode === 'deny'
                    ? 'border-danger bg-danger/10 text-danger'
                    : 'border-border text-text-primary hover:border-danger/60'
                "
                @click="setInteractionMode('deny')"
              >
                deny 模式
              </button>
              <button
                type="button"
                class="rounded-xl border border-danger/30 bg-danger/10 px-3 py-2 text-sm font-medium text-danger transition hover:bg-danger/15"
                @click="removeSelectedCanvasItem"
              >
                删除当前选中
              </button>
            </div>

            <div
              class="mb-4 rounded-2xl border border-border bg-elevated px-4 py-3 text-sm text-text-secondary"
            >
              <div class="flex flex-wrap items-center gap-2">
                <span
                  class="rounded-full border border-primary/20 bg-primary/10 px-2.5 py-1 text-xs text-primary"
                >
                  当前模式：{{ canvasModeLabel }}
                </span>
                <span
                  class="rounded-full border border-border-subtle bg-surface px-2.5 py-1 text-xs text-text-secondary"
                >
                  {{ selectedCanvasSummary }}
                </span>
                <span
                  class="rounded-full border border-border-subtle bg-surface px-2.5 py-1 text-xs text-text-muted"
                >
                  `Esc` 取消连线 / `Delete` 删除选中
                </span>
              </div>
            </div>

            <div
              class="mb-4 rounded-2xl border px-4 py-3 text-sm"
              :class="
                draftValidationIssues.length === 0
                  ? 'border-[var(--color-success)]/20 bg-[var(--color-success)]/10 text-[var(--color-success)]'
                  : 'border-[var(--color-warning)]/20 bg-[var(--color-warning)]/10 text-[var(--color-warning)]'
              "
            >
              <div class="font-medium">
                {{ draftValidationIssues.length === 0 ? '基础校验已通过' : '基础校验发现问题' }}
              </div>
              <div
                v-if="draftValidationIssues.length === 0"
                class="mt-1 text-xs text-[var(--color-success)]/80"
              >
                当前草稿的入口、节点、网络和链路引用关系正常。
              </div>
              <ul
                v-else
                class="mt-2 space-y-1 text-xs"
              >
                <li
                  v-for="issue in draftValidationIssues"
                  :key="issue"
                >
                  {{ issue }}
                </li>
              </ul>
            </div>

            <TopologyCanvasBoard
              :graph="canvasGraph"
              :interaction-mode="interactionMode"
              :pending-source-node-key="pendingSourceNodeKey"
              :selected-node-key="selectedNodeKey"
              :selected-edge-id="selectedEdgeId"
              @select-node="handleCanvasSelectNode"
              @select-edge="handleCanvasSelectEdge"
              @create-node-at="handleCanvasCreateNode"
              @create-edge="handleCanvasCreateEdge"
              @clear-pending="pendingSourceNodeKey = null"
              @update-position="updateNodePosition"
            />

            <div class="mt-4 grid gap-4 xl:grid-cols-[1.08fr_0.92fr]">
              <div class="rounded-2xl border border-border bg-elevated p-4">
                <div class="text-sm font-semibold text-text-primary">
                  画布快速编辑
                </div>

                <div
                  v-if="!selectedNodeDraft && !selectedEdgeMeta"
                  class="mt-3 rounded-xl border border-dashed border-border px-4 py-6 text-sm text-text-muted"
                >
                  请选择一个节点或一条边
                </div>

                <div
                  v-else-if="selectedNodeDraft"
                  class="mt-3 space-y-4"
                >
                  <div class="grid gap-3 md:grid-cols-2">
                    <label class="space-y-2">
                      <span class="text-sm text-text-secondary">节点名称</span>
                      <input
                        v-model="selectedNodeDraft.name"
                        type="text"
                        class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
                      >
                    </label>
                    <label class="space-y-2">
                      <span class="text-sm text-text-secondary">镜像</span>
                      <select
                        v-model="selectedNodeDraft.image_id"
                        class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
                      >
                        <option value="">复用题目主镜像</option>
                        <option
                          v-for="image in images"
                          :key="image.id"
                          :value="image.id"
                        >
                          {{ image.name }}:{{ image.tag }}
                        </option>
                      </select>
                    </label>
                    <label class="space-y-2">
                      <span class="text-sm text-text-secondary">层级</span>
                      <select
                        v-model="selectedNodeDraft.tier"
                        class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
                      >
                        <option value="public">public</option>
                        <option value="service">service</option>
                        <option value="internal">internal</option>
                      </select>
                    </label>
                    <label class="space-y-2">
                      <span class="text-sm text-text-secondary">服务端口</span>
                      <input
                        :value="selectedNodeDraft.service_port ?? ''"
                        type="number"
                        min="1"
                        max="65535"
                        class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
                        @input="
                          updateCanvasQuickNumber(
                            'service_port',
                            ($event.target as HTMLInputElement).value,
                            selectedNodeDraft
                          )
                        "
                      >
                    </label>
                  </div>

                  <label
                    class="flex items-center gap-3 rounded-xl border border-border bg-surface px-3 py-3 text-sm text-text-primary"
                  >
                    <input
                      v-model="selectedNodeDraft.inject_flag"
                      type="checkbox"
                      class="h-4 w-4 rounded border-border bg-transparent"
                    >
                    启用 Flag 注入
                  </label>

                  <div class="space-y-2">
                    <div class="text-sm text-text-secondary">
                      所属网络
                    </div>
                    <div class="grid gap-2 md:grid-cols-2">
                      <label
                        v-for="network in draft.networks"
                        :key="network.uid"
                        class="flex items-center gap-3 rounded-xl border border-border bg-surface px-3 py-3 text-sm text-text-primary"
                      >
                        <input
                          :checked="selectedNodeDraft.network_keys.includes(network.key)"
                          type="checkbox"
                          class="h-4 w-4 rounded border-border bg-transparent"
                          @change="
                            toggleSelectedNodeNetwork(
                              network.key,
                              ($event.target as HTMLInputElement).checked
                            )
                          "
                        >
                        <span>{{ network.name || network.key }}</span>
                      </label>
                    </div>
                  </div>
                </div>

                <div
                  v-else-if="selectedEdgeMeta"
                  class="mt-3 space-y-4"
                >
                  <div class="grid gap-3 md:grid-cols-2">
                    <label class="space-y-2">
                      <span class="text-sm text-text-secondary">源节点</span>
                      <select
                        :value="selectedEdgeSourceKey"
                        class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
                        @change="
                          updateSelectedEdgeSourceKey(($event.target as HTMLSelectElement).value)
                        "
                      >
                        <option
                          v-for="node in nodeOptions"
                          :key="node.key"
                          :value="node.key"
                        >
                          {{ node.label }}
                        </option>
                      </select>
                    </label>
                    <label class="space-y-2">
                      <span class="text-sm text-text-secondary">目标节点</span>
                      <select
                        :value="selectedEdgeTargetKey"
                        class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
                        @change="
                          updateSelectedEdgeTargetKey(($event.target as HTMLSelectElement).value)
                        "
                      >
                        <option
                          v-for="node in nodeOptions"
                          :key="node.key"
                          :value="node.key"
                        >
                          {{ node.label }}
                        </option>
                      </select>
                    </label>
                  </div>

                  <label class="space-y-2">
                    <span class="text-sm text-text-secondary">边类型</span>
                    <select
                      :value="selectedEdgeKind"
                      class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
                      @change="
                        handleSelectedEdgeKindChange(($event.target as HTMLSelectElement).value)
                      "
                    >
                      <option value="link">logic link</option>
                      <option value="allow">allow</option>
                      <option value="deny">deny</option>
                    </select>
                  </label>
                </div>
              </div>

              <div class="rounded-2xl border border-border bg-elevated p-4">
                <div class="text-sm font-semibold text-text-primary">
                  网络快速编辑
                </div>
                <div class="mt-3 space-y-3">
                  <div
                    v-for="network in draft.networks"
                    :key="network.uid"
                    class="grid gap-3 rounded-xl border border-border bg-surface p-3 md:grid-cols-[0.9fr_1fr_auto]"
                  >
                    <input
                      v-model="network.key"
                      type="text"
                      class="w-full rounded-xl border border-border bg-elevated px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
                      placeholder="network key"
                    >
                    <input
                      v-model="network.name"
                      type="text"
                      class="w-full rounded-xl border border-border bg-elevated px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
                      placeholder="网络名称"
                    >
                    <label
                      class="flex items-center gap-2 rounded-xl border border-border bg-elevated px-3 py-2.5 text-sm text-text-primary"
                    >
                      <input
                        v-model="network.internal"
                        type="checkbox"
                        class="h-4 w-4 rounded border-border bg-transparent"
                      >
                      internal
                    </label>
                  </div>
                </div>
              </div>
            </div>
          </SectionCard>

          <SectionCard
            title="入口节点"
            subtitle="实例访问入口和当前草稿的保存范围。"
          >
            <div class="grid gap-4 md:grid-cols-[1fr_auto]">
              <label class="space-y-2">
                <span class="text-sm text-text-secondary">入口节点</span>
                <select
                  v-model="draft.entry_node_key"
                  class="w-full rounded-xl border border-border bg-elevated px-3 py-3 text-sm text-text-primary outline-none transition focus:border-primary"
                >
                  <option
                    v-for="node in nodeOptions"
                    :key="node.key"
                    :value="node.key"
                  >
                    {{ node.label }} ({{ node.key }})
                  </option>
                </select>
              </label>

              <button
                v-if="!isTemplateLibraryMode"
                type="button"
                class="inline-flex items-center gap-2 self-end rounded-xl border border-danger/30 bg-danger/10 px-4 py-3 text-sm font-medium text-danger transition hover:bg-danger/15"
                :disabled="saving || !topology"
                @click="void handleDeleteTopology()"
              >
                <Trash2 class="h-4 w-4" />
                删除已保存拓扑
              </button>
            </div>
          </SectionCard>

          <SectionCard
            title="网络分段"
            subtitle="一个节点可以挂多个网络，运行时会创建多个 Docker Network。"
          >
            <div class="space-y-3">
              <div
                v-for="network in draft.networks"
                :key="network.uid"
                class="grid gap-3 rounded-2xl border border-border bg-elevated p-4 md:grid-cols-[0.9fr_1fr_0.9fr_auto_auto]"
              >
                <input
                  v-model="network.key"
                  type="text"
                  class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
                  placeholder="network key"
                >
                <input
                  v-model="network.name"
                  type="text"
                  class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
                  placeholder="网络名称"
                >
                <input
                  v-model="network.cidr"
                  type="text"
                  class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
                  placeholder="CIDR（可选）"
                >
                <label
                  class="flex items-center gap-3 rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary"
                >
                  <input
                    v-model="network.internal"
                    type="checkbox"
                    class="h-4 w-4 rounded border-border bg-transparent"
                  >
                  internal
                </label>
                <button
                  type="button"
                  class="inline-flex items-center justify-center rounded-xl border border-danger/30 bg-danger/10 px-3 py-2 text-sm text-danger transition hover:bg-danger/15"
                  :disabled="draft.networks.length <= 1"
                  @click="removeNetwork(network.uid)"
                >
                  <Trash2 class="h-4 w-4" />
                </button>
              </div>
            </div>

            <template #footer>
              <button
                type="button"
                class="inline-flex items-center gap-2 rounded-xl border border-border px-4 py-2 text-sm font-medium text-text-primary transition hover:border-primary"
                @click="addNetwork"
              >
                <Plus class="h-4 w-4" />
                添加网络
              </button>
            </template>
          </SectionCard>

          <SectionCard
            title="节点编排"
            subtitle="节点支持单独镜像、资源限制、网络归属和环境变量。"
          >
            <div class="space-y-4">
              <TopologyNodeEditor
                v-for="(node, index) in draft.nodes"
                :key="node.uid"
                :data-node-editor="node.key"
                :model-value="node"
                :index="index"
                :images="images"
                :networks="draft.networks"
                :removable="draft.nodes.length > 1"
                :selected="selectedNodeKey === node.key"
                @update:model-value="draft.nodes[index] = $event"
                @remove="removeNode(node.uid)"
              />
            </div>

            <template #footer>
              <button
                type="button"
                class="inline-flex items-center gap-2 rounded-xl border border-border px-4 py-2 text-sm font-medium text-text-primary transition hover:border-primary"
                @click="addNode"
              >
                <Plus class="h-4 w-4" />
                添加节点
              </button>
            </template>
          </SectionCard>

          <SectionCard
            title="拓扑连线"
            subtitle="用于表达逻辑依赖关系，不直接等同于运行时 ACL。"
          >
            <div
              v-if="draft.links.length === 0"
              class="rounded-xl border border-dashed border-border px-4 py-6 text-sm text-text-muted"
            >
              暂无逻辑连线
            </div>
            <div
              v-else
              class="space-y-3"
            >
              <div
                v-for="link in draft.links"
                :key="link.uid"
                class="grid gap-3 rounded-2xl border border-border bg-elevated p-4 md:grid-cols-[1fr_1fr_auto]"
              >
                <select
                  v-model="link.from_node_key"
                  class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
                >
                  <option value="">
                    选择源节点
                  </option>
                  <option
                    v-for="node in nodeOptions"
                    :key="node.key"
                    :value="node.key"
                  >
                    {{ node.label }}
                  </option>
                </select>
                <select
                  v-model="link.to_node_key"
                  class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
                >
                  <option value="">
                    选择目标节点
                  </option>
                  <option
                    v-for="node in nodeOptions"
                    :key="node.key"
                    :value="node.key"
                  >
                    {{ node.label }}
                  </option>
                </select>
                <button
                  type="button"
                  class="inline-flex items-center justify-center rounded-xl border border-danger/30 bg-danger/10 px-3 py-2 text-sm text-danger transition hover:bg-danger/15"
                  @click="draft.links = draft.links.filter((item) => item.uid !== link.uid)"
                >
                  <Trash2 class="h-4 w-4" />
                </button>
              </div>
            </div>

            <template #footer>
              <button
                type="button"
                class="inline-flex items-center gap-2 rounded-xl border border-border px-4 py-2 text-sm font-medium text-text-primary transition hover:border-primary"
                @click="addLink"
              >
                <Link2 class="h-4 w-4" />
                添加连线
              </button>
            </template>
          </SectionCard>

          <SectionCard
            title="链路策略"
            subtitle="当前前端只开放粗粒度节点 allow/deny，细粒度端口策略尚未支持。"
          >
            <div
              v-if="draft.policies.length === 0"
              class="rounded-xl border border-dashed border-border px-4 py-6 text-sm text-text-muted"
            >
              暂无链路策略
            </div>
            <div
              v-else
              class="space-y-3"
            >
              <div
                v-for="policy in draft.policies"
                :key="policy.uid"
                class="grid gap-3 rounded-2xl border border-border bg-elevated p-4 md:grid-cols-[1fr_1fr_0.7fr_auto]"
              >
                <select
                  v-model="policy.source_node_key"
                  class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
                >
                  <option value="">
                    选择源节点
                  </option>
                  <option
                    v-for="node in nodeOptions"
                    :key="node.key"
                    :value="node.key"
                  >
                    {{ node.label }}
                  </option>
                </select>
                <select
                  v-model="policy.target_node_key"
                  class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
                >
                  <option value="">
                    选择目标节点
                  </option>
                  <option
                    v-for="node in nodeOptions"
                    :key="node.key"
                    :value="node.key"
                  >
                    {{ node.label }}
                  </option>
                </select>
                <select
                  v-model="policy.action"
                  class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
                >
                  <option value="allow">
                    allow
                  </option>
                  <option value="deny">
                    deny
                  </option>
                </select>
                <button
                  type="button"
                  class="inline-flex items-center justify-center rounded-xl border border-danger/30 bg-danger/10 px-3 py-2 text-sm text-danger transition hover:bg-danger/15"
                  @click="draft.policies = draft.policies.filter((item) => item.uid !== policy.uid)"
                >
                  <Trash2 class="h-4 w-4" />
                </button>
              </div>
            </div>

            <template #footer>
              <button
                type="button"
                class="inline-flex items-center gap-2 rounded-xl border border-border px-4 py-2 text-sm font-medium text-text-primary transition hover:border-primary"
                @click="addPolicy"
              >
                <ShieldBan class="h-4 w-4" />
                添加策略
              </button>
            </template>
          </SectionCard>
        </div>

        <aside class="context-rail topology-context-rail">
          <div class="topology-context-stack">
            <section class="topology-status-list">
              <article class="topology-status-note topology-status-note--primary">
                <div class="topology-status-note__icon">
                  <Blocks class="h-5 w-5" />
                </div>
                <div class="topology-status-note__body">
                  <div class="topology-status-note__eyebrow">{{ statusCard.eyebrow }}</div>
                  <div class="topology-status-note__title">{{ statusCard.title }}</div>
                  <p class="topology-status-note__copy">{{ statusCard.subtitle }}</p>
                </div>
              </article>

              <article class="topology-status-note topology-status-note--warning">
                <div class="topology-status-note__icon">
                  <GitBranch class="h-5 w-5" />
                </div>
                <div class="topology-status-note__body">
                  <div class="topology-status-note__eyebrow">{{ secondaryCard.eyebrow }}</div>
                  <div class="topology-status-note__title">{{ secondaryCard.title }}</div>
                  <p class="topology-status-note__copy">{{ secondaryCard.subtitle }}</p>
                </div>
              </article>

              <article class="topology-status-note topology-status-note--danger">
                <div class="topology-status-note__icon">
                  <ShieldBan class="h-5 w-5" />
                </div>
                <div class="topology-status-note__body">
                  <div class="topology-status-note__eyebrow">运行时约束</div>
                  <div class="topology-status-note__title">粗粒度</div>
                  <p class="topology-status-note__copy">
                    当前只支持节点级 allow/deny，不支持端口级 ACL。
                  </p>
                </div>
              </article>
            </section>

            <div class="topology-side-stack space-y-6">
          <SectionCard
            title="模板库"
            :subtitle="
              isTemplateLibraryMode
                ? '从模板库载入后可直接编辑并覆盖模板，或另存为新模板。'
                : '可按模板快速回填编辑器，或直接应用到题目。'
            "
          >
            <div class="space-y-3">
              <div :class="isTemplateLibraryMode ? 'template-focus-card' : 'rounded-2xl border border-border bg-elevated px-4 py-4'">
                <div class="text-xs font-semibold uppercase tracking-[0.22em] text-text-muted">
                  当前模板
                </div>
                <div class="mt-2 text-sm text-text-primary">
                  {{ selectedTemplateSummary }}
                </div>
                <div class="mt-3 flex flex-wrap gap-2">
                  <button
                    v-if="selectedTemplate"
                    type="button"
                    class="rounded-xl border border-border px-3 py-2 text-xs font-medium text-text-primary transition hover:border-primary"
                    @click="loadTemplateIntoDraft(selectedTemplate)"
                  >
                    重新载入当前模板
                  </button>
                  <button
                    v-if="selectedTemplate"
                    type="button"
                    class="rounded-xl border border-border px-3 py-2 text-xs font-medium text-text-primary transition hover:border-primary"
                    @click="clearTemplateSelection"
                  >
                    清空模板选择
                  </button>
                </div>
              </div>

              <div :class="isTemplateLibraryMode ? 'template-search-row' : 'grid gap-3 md:grid-cols-[1fr_auto]'">
                <input
                  v-model="templateKeyword"
                  type="text"
                  class="w-full rounded-xl border border-border bg-elevated px-3 py-3 text-sm text-text-primary outline-none transition focus:border-primary"
                  placeholder="按模板名称搜索"
                >
                <button
                  type="button"
                  class="rounded-xl border border-border px-4 py-3 text-sm font-medium text-text-primary transition hover:border-primary"
                  @click="void loadTemplates()"
                >
                  搜索
                </button>
              </div>

              <div
                v-if="templates.length === 0"
                :class="isTemplateLibraryMode ? 'template-empty-state' : 'rounded-xl border border-dashed border-border px-4 py-6 text-sm text-text-muted'"
              >
                当前没有模板数据
              </div>

              <div
                v-else
                :class="isTemplateLibraryMode ? 'template-library-list' : 'space-y-3'"
              >
                <article
                  v-for="template in templates"
                  :key="template.id"
                  :class="
                    isTemplateLibraryMode
                      ? [
                          'template-library-item',
                          selectedTemplateId === template.id
                            ? 'template-library-item--active'
                            : 'template-library-item--idle',
                        ]
                      : [
                          'rounded-2xl border p-4 transition',
                          selectedTemplateId === template.id
                            ? 'border-primary bg-primary/8'
                            : 'border-border bg-elevated',
                        ]
                  "
                >
                  <div class="min-w-0">
                    <div class="truncate text-base font-semibold text-text-primary">
                      {{ template.name }}
                    </div>
                    <div class="mt-1 text-sm text-text-secondary">
                      {{ template.description || '无描述' }}
                    </div>
                    <div class="mt-2 flex flex-wrap gap-2 text-xs text-text-muted">
                      <span>入口：{{ template.entry_node_key }}</span>
                      <span>节点：{{ template.nodes.length }}</span>
                      <span>网络：{{ template.networks?.length || 0 }}</span>
                      <span>使用：{{ template.usage_count }}</span>
                    </div>
                  </div>

                  <div class="mt-4 flex flex-wrap gap-2">
                    <button
                      type="button"
                      :class="isTemplateLibraryMode ? 'template-action-btn' : 'rounded-xl border border-border px-3 py-2 text-xs font-medium text-text-primary transition hover:border-primary'"
                      @click="loadTemplateIntoDraft(template)"
                    >
                      {{ isTemplateLibraryMode ? '载入编辑' : '载入草稿' }}
                    </button>
                    <button
                      v-if="!isTemplateLibraryMode"
                      type="button"
                      :class="isTemplateLibraryMode ? 'template-action-btn' : 'rounded-xl border border-border px-3 py-2 text-xs font-medium text-text-primary transition hover:border-primary'"
                      @click="resetTemplateForm(template)"
                    >
                      选中
                    </button>
                    <button
                      v-if="!isTemplateLibraryMode"
                      type="button"
                      :class="isTemplateLibraryMode ? 'template-action-btn template-action-btn--primary' : 'rounded-xl bg-primary px-3 py-2 text-xs font-medium text-white transition hover:opacity-90'"
                      :disabled="templateBusy"
                      @click="void handleApplyTemplate(template)"
                    >
                      应用到题目
                    </button>
                    <button
                      type="button"
                      :class="isTemplateLibraryMode ? 'template-action-btn template-action-btn--danger' : 'rounded-xl border border-danger/30 bg-danger/10 px-3 py-2 text-xs font-medium text-danger transition hover:bg-danger/15'"
                      :disabled="templateBusy"
                      @click="void handleDeleteTemplate(template.id)"
                    >
                      删除模板
                    </button>
                  </div>
                </article>
              </div>
            </div>
          </SectionCard>

          <SectionCard
            title="模板写回"
            :subtitle="
              isTemplateLibraryMode
                ? '在独立模板库中可新建空白草稿、载入现有模板后覆盖，或另存为新模板。'
                : '把当前编辑器草稿保存为新模板，或覆盖已选中的模板。'
            "
          >
            <div :class="isTemplateLibraryMode ? 'template-writeback-form' : 'space-y-4'">
              <label class="space-y-2">
                <span class="text-sm text-text-secondary">模板名称</span>
                <input
                  v-model="templateName"
                  type="text"
                  class="w-full rounded-xl border border-border bg-elevated px-3 py-3 text-sm text-text-primary outline-none transition focus:border-primary"
                  placeholder="例如 双节点 Web + DB"
                >
              </label>

              <label class="space-y-2">
                <span class="text-sm text-text-secondary">模板描述</span>
                <textarea
                  v-model="templateDescription"
                  rows="4"
                  class="w-full rounded-xl border border-border bg-elevated px-3 py-3 text-sm text-text-primary outline-none transition focus:border-primary"
                  placeholder="说明这个模板的适用场景"
                />
              </label>

              <div class="flex flex-wrap gap-2">
                <button
                  v-if="isTemplateLibraryMode"
                  type="button"
                  :class="isTemplateLibraryMode ? 'template-action-btn' : 'inline-flex items-center gap-2 rounded-xl border border-border px-4 py-2 text-sm font-medium text-text-primary transition hover:border-primary'"
                  @click="handleResetTemplateEditor"
                >
                  新建空白草稿
                </button>
                <button
                  type="button"
                  :class="isTemplateLibraryMode ? 'template-action-btn template-action-btn--primary' : 'inline-flex items-center gap-2 rounded-xl bg-primary px-4 py-2 text-sm font-medium text-white transition hover:opacity-90'"
                  :disabled="templateBusy"
                  @click="void handleCreateTemplate()"
                >
                  <Plus class="h-4 w-4" />
                  保存为新模板
                </button>
                <button
                  type="button"
                  :class="isTemplateLibraryMode ? 'template-action-btn' : 'inline-flex items-center gap-2 rounded-xl border border-border px-4 py-2 text-sm font-medium text-text-primary transition hover:border-primary'"
                  :disabled="templateBusy || !selectedTemplateId"
                  @click="void handleUpdateTemplate()"
                >
                  覆盖已选模板
                </button>
              </div>
            </div>
          </SectionCard>

          <SectionCard
            title="当前边界"
            subtitle="避免把未生效能力继续暴露成可用配置。"
          >
            <div :class="isTemplateLibraryMode ? 'template-boundary-list' : 'space-y-4'">
              <AppCard
                variant="action"
                accent="warning"
                eyebrow="已开放"
                subtitle="多网络、节点、逻辑连线、粗粒度 allow/deny 策略、模板复用。"
              >
                <template #default />
              </AppCard>
              <AppCard
                variant="action"
                accent="danger"
                eyebrow="暂未开放"
                subtitle="protocol / ports 级细粒度 ACL 前端字段、模板版本化与批量比对能力。"
              >
                <template #default />
              </AppCard>
              <AppCard
                variant="action"
                accent="neutral"
                eyebrow="建议"
                subtitle="继续开放高级能力前，先补参数校验、可视化提示和误操作保护。"
              >
                <template #default />
              </AppCard>
            </div>
          </SectionCard>
            </div>
          </div>
        </aside>
      </main>
      </template>

      <AppEmpty
        v-if="!challenge && !isTemplateLibraryMode"
        title="题目不存在"
        description="无法读取当前题目的基础信息，请返回题目列表后重试。"
        icon="Blocks"
      />
    </template>
  </div>
</template>

<style scoped>
.topology-page--challenge {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-border: color-mix(in srgb, var(--color-border-default) 84%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 80%, var(--color-bg-base));
  --journal-accent: color-mix(in srgb, var(--color-primary) 88%, var(--journal-ink));
  --topology-panel: color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base));
  --topology-panel-subtle: color-mix(in srgb, var(--journal-surface-subtle) 96%, var(--color-bg-base));
  --topology-divider: color-mix(in srgb, var(--journal-border) 88%, transparent);
  display: flex;
  flex-direction: column;
  min-height: max(100%, calc(100vh - 5rem));
  border: 1px solid var(--journal-border);
  border-radius: 30px;
  padding: var(--space-6) var(--space-7);
  background:
    radial-gradient(circle at top right, color-mix(in srgb, var(--journal-accent) 8%, transparent), transparent 22rem),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 98%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle) 96%, var(--color-bg-base))
    );
  box-shadow: 0 24px 56px var(--color-shadow-soft);
}

.topology-page--challenge .workspace-topbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
  padding-bottom: var(--space-6);
}

.topology-page--challenge .topology-topbar-leading {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-2-5);
}

.topology-page--challenge .workspace-overline,
.topology-page--challenge .topology-page-kicker {
  display: inline-flex;
  align-items: center;
  font-size: var(--font-size-0-72);
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.topology-page--challenge .topology-topbar-chip {
  display: inline-flex;
  align-items: center;
  min-height: 1.8rem;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--journal-accent) 18%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 8%, transparent);
  padding: 0 var(--space-3);
  font-size: var(--font-size-0-76);
  font-weight: 600;
  color: var(--journal-accent);
}

.topology-page--challenge .topology-topbar-actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.topology-page--challenge .topology-toolbar-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  min-height: 2.45rem;
  border-radius: 0.75rem;
  border: 1px solid var(--journal-border);
  padding: var(--space-2) var(--space-4);
  font-size: var(--font-size-0-84);
  font-weight: 600;
  transition:
    border-color 150ms ease,
    background 150ms ease,
    color 150ms ease,
    box-shadow 150ms ease;
}

.topology-page--challenge .topology-toolbar-btn:focus-visible {
  outline: none;
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--journal-accent) 18%, transparent);
}

.topology-page--challenge .topology-toolbar-btn--ghost {
  background: color-mix(in srgb, var(--journal-surface) 94%, var(--color-bg-base));
  color: var(--journal-ink);
}

.topology-page--challenge .topology-toolbar-btn--ghost:hover {
  border-color: color-mix(in srgb, var(--journal-accent) 28%, transparent);
  color: var(--journal-accent);
}

.topology-page--challenge .topology-toolbar-btn--primary {
  border-color: transparent;
  background: var(--journal-accent);
  color: var(--color-bg-base);
  box-shadow: 0 12px 28px color-mix(in srgb, var(--journal-accent) 16%, transparent);
}

.topology-page--challenge .topology-toolbar-btn--primary:hover {
  background: color-mix(in srgb, var(--journal-accent) 88%, black);
}

.topology-page--challenge .topology-toolbar-btn:disabled {
  opacity: 0.65;
  cursor: not-allowed;
  box-shadow: none;
}

.topology-page--challenge .topology-page-heading {
  display: grid;
  gap: var(--space-4);
  padding-bottom: var(--space-6);
}

.topology-page--challenge .topology-page-copy {
  max-width: 48rem;
}

.topology-page--challenge .topology-summary-grid--challenge {
  margin-top: var(--space-2);
  --metric-panel-grid-gap: var(--space-3);
  --metric-panel-columns: repeat(4, minmax(0, 1fr));
}

.topology-page--challenge .topology-summary-card {
  border: 1px solid var(--journal-border);
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--topology-panel) 98%, var(--color-bg-base)),
    color-mix(in srgb, var(--topology-panel-subtle) 96%, var(--color-bg-base))
  );
  box-shadow: 0 10px 24px var(--color-shadow-soft);
}

.topology-page--challenge .content-pane.topology-workspace {
  display: grid;
  gap: var(--space-7);
  grid-template-columns: minmax(0, 1fr) minmax(19rem, 22rem);
  align-items: start;
  min-width: 0;
  padding: 0;
}

.topology-page--challenge .topology-primary-column,
.topology-page--challenge .topology-context-stack,
.topology-page--challenge .topology-side-stack {
  display: grid;
  gap: var(--space-6);
}

.topology-page--challenge .topology-context-rail {
  min-width: 0;
  padding-left: var(--space-6);
  border-left: 1px solid var(--topology-divider);
}

.topology-page--challenge .topology-context-stack {
  position: sticky;
  top: var(--space-6);
}

.topology-page--challenge .topology-status-list {
  display: grid;
  gap: var(--space-3);
}

.topology-page--challenge .topology-status-note {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  gap: var(--space-3);
  padding: var(--space-4);
  border: 1px solid var(--journal-border);
  border-radius: 18px;
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--topology-panel) 98%, var(--color-bg-base)),
    color-mix(in srgb, var(--topology-panel-subtle) 96%, var(--color-bg-base))
  );
  box-shadow: 0 12px 28px var(--color-shadow-soft);
}

.topology-page--challenge .topology-status-note__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2.75rem;
  height: 2.75rem;
  border-radius: 0.9rem;
}

.topology-page--challenge .topology-status-note--primary .topology-status-note__icon {
  border: 1px solid color-mix(in srgb, var(--journal-accent) 22%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  color: var(--journal-accent);
}

.topology-page--challenge .topology-status-note--warning .topology-status-note__icon {
  border: 1px solid color-mix(in srgb, var(--color-warning) 24%, transparent);
  background: color-mix(in srgb, var(--color-warning) 10%, transparent);
  color: var(--color-warning);
}

.topology-page--challenge .topology-status-note--danger .topology-status-note__icon {
  border: 1px solid color-mix(in srgb, var(--color-danger) 24%, transparent);
  background: color-mix(in srgb, var(--color-danger) 10%, transparent);
  color: color-mix(in srgb, var(--color-danger) 88%, var(--journal-ink));
}

.topology-page--challenge .topology-status-note__eyebrow {
  font-size: var(--font-size-0-72);
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.topology-page--challenge .topology-status-note__title {
  margin-top: var(--space-1);
  font-size: var(--font-size-1-10);
  font-weight: 700;
  color: var(--journal-ink);
}

.topology-page--challenge .topology-status-note__copy {
  margin-top: var(--space-1-5);
  font-size: var(--font-size-0-86);
  line-height: 1.65;
  color: var(--journal-muted);
}

.topology-page--challenge :deep(.section-card) {
  padding: var(--space-5) 0 0;
  border-top: 1px solid var(--topology-divider);
}

.topology-page--challenge .topology-primary-column :deep(.section-card:first-child),
.topology-page--challenge .topology-side-stack :deep(.section-card:first-child) {
  padding-top: 0;
  border-top: 0;
}

.topology-page--challenge :deep(.section-card__header) {
  margin-bottom: var(--space-4);
  padding-bottom: var(--space-3);
  border-bottom-color: var(--topology-divider);
}

.topology-page--challenge :deep(.section-card__header h2) {
  color: var(--journal-ink);
  font-size: var(--font-size-1-08);
}

.topology-page--challenge :deep(.section-card__header p) {
  color: var(--journal-muted);
}

.topology-page--challenge :deep(.section-card__body) {
  padding-left: 0;
}

.topology-page--challenge :deep(.section-card__body > .rounded-2xl),
.topology-page--challenge :deep(.section-card__body > .rounded-xl),
.topology-page--challenge :deep([data-node-editor]),
.topology-page--challenge :deep(.topology-canvas-board__root) {
  border-color: var(--journal-border);
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--topology-panel) 98%, var(--color-bg-base)),
    color-mix(in srgb, var(--topology-panel-subtle) 96%, var(--color-bg-base))
  );
  box-shadow: 0 14px 30px var(--color-shadow-soft);
}

.topology-page--challenge :deep(input),
.topology-page--challenge :deep(select),
.topology-page--challenge :deep(textarea) {
  border-color: var(--journal-border);
  background: color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base));
  color: var(--journal-ink);
}

.topology-page--challenge :deep(input::placeholder),
.topology-page--challenge :deep(textarea::placeholder) {
  color: color-mix(in srgb, var(--journal-muted) 78%, transparent);
}

.topology-page--challenge :deep(option) {
  background: var(--journal-surface);
  color: var(--journal-ink);
}

.topology-page--challenge :deep(input:focus),
.topology-page--challenge :deep(select:focus),
.topology-page--challenge :deep(textarea:focus) {
  border-color: color-mix(in srgb, var(--journal-accent) 48%, transparent);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--journal-accent) 14%, transparent);
  outline: none;
}

.topology-page--challenge :deep(.topology-canvas-board__surface) {
  border-color: color-mix(in srgb, var(--journal-border) 70%, transparent);
}

.topology-loading-pane {
  display: flex;
  justify-content: center;
  padding-block: var(--space-10);
}

.topology-page--template-library {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-accent: var(--color-primary);
  --journal-border: color-mix(in srgb, var(--color-border-default) 84%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 78%, var(--color-bg-base));
  display: grid;
  gap: var(--space-5);
}

.topology-page--template-library .template-library-main,
.topology-page--template-library :deep(.page-header) {
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

.topology-page--template-library .template-library-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-3);
}

.topology-page--template-library .template-library-divider {
  margin: var(--space-6) 0;
  border-top: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.topology-page--template-library .topology-hero-lead--library {
  padding: 0;
}

.topology-page--template-library .topology-hero-kicker {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--font-size-0-72);
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.topology-page--template-library .topology-hero-badge {
  border: 1px solid color-mix(in srgb, var(--journal-accent) 18%, transparent);
  border-radius: 0.6rem;
  background: color-mix(in srgb, var(--journal-accent) 7%, transparent);
  padding: var(--space-1) var(--space-2);
  color: var(--journal-accent);
}

.topology-page--template-library .topology-hero-description {
  max-width: 46rem;
}

.topology-page--template-library .topology-summary-grid {
  margin-top: var(--space-6);
  --metric-panel-grid-gap: var(--space-3);
  --metric-panel-columns: repeat(4, minmax(0, 1fr));
}

.topology-page--template-library .template-focus-card,
.topology-page--template-library .template-empty-state {
  padding: 0 0 0 var(--space-4);
  border: 0;
  border-left: 2px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  border-radius: 0;
  background: transparent;
  box-shadow: none;
}

.topology-page--template-library .topology-hero-aside--library {
  align-self: start;
  border-left: 0;
  padding-left: 0;
}

.topology-page--template-library .template-hero-note {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  gap: var(--space-3);
  padding: 0 0 0 var(--space-4);
  border-left: 2px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.topology-page--template-library .template-hero-note__body {
  min-width: 0;
}

.topology-page--template-library .template-hero-note__label {
  font-size: var(--font-size-0-72);
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.topology-page--template-library .template-hero-note__value {
  margin-top: var(--space-1-5);
  font-size: var(--font-size-1-10);
  font-weight: 700;
  color: var(--journal-ink);
}

.topology-page--template-library .template-hero-note__copy {
  margin-top: var(--space-1-5);
  font-size: var(--font-size-0-86);
  line-height: 1.6;
  color: var(--journal-muted);
}

.topology-page--template-library .template-metric-icon {
  display: flex;
  height: 2.75rem;
  width: 2.75rem;
  align-items: center;
  justify-content: center;
  border-radius: 0.9rem;
}

.topology-page--template-library .template-metric-icon--primary {
  border: 1px solid color-mix(in srgb, var(--journal-accent) 22%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 10%, transparent);
  color: var(--journal-accent);
}

.topology-page--template-library .template-metric-icon--warning {
  border: 1px solid color-mix(in srgb, var(--color-warning) 24%, transparent);
  background: color-mix(in srgb, var(--color-warning) 10%, transparent);
  color: var(--color-warning);
}

.topology-page--template-library .template-metric-icon--danger {
  border: 1px solid color-mix(in srgb, var(--color-danger) 24%, transparent);
  background: color-mix(in srgb, var(--color-danger) 10%, transparent);
  color: var(--color-danger);
}

.topology-page--template-library .topology-toolbar-btn,
.topology-page--template-library .template-action-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  min-height: 2.45rem;
  border-radius: 0.75rem;
  border: 1px solid var(--journal-border);
  padding: var(--space-2) var(--space-4);
  font-size: var(--font-size-0-82);
  font-weight: 600;
  transition:
    border-color 150ms ease,
    background 150ms ease,
    color 150ms ease;
}

.topology-page--template-library .template-toolbar-tabs {
  display: flex;
  flex: 1 1 auto;
  align-items: center;
  gap: var(--space-4);
  min-height: 2.8rem;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.topology-page--template-library .template-toolbar-tab {
  display: inline-flex;
  min-height: 2.8rem;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  padding: 0 var(--space-0-5);
  border-bottom: 2px solid transparent;
  font-size: var(--font-size-0-88);
  font-weight: 700;
  transition:
    border-color 150ms ease,
    color 150ms ease,
    background 150ms ease;
}

.topology-page--template-library .template-toolbar-tab--idle {
  color: var(--journal-muted);
}

.topology-page--template-library .template-toolbar-tab--idle:hover {
  color: var(--journal-ink);
}

.topology-page--template-library .template-toolbar-tab--active {
  border-bottom-color: color-mix(in srgb, var(--journal-accent) 58%, transparent);
  color: var(--journal-accent);
}

.topology-page--template-library .template-toolbar-refresh {
  display: inline-flex;
  height: 2.5rem;
  width: 2.5rem;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--journal-border);
  border-radius: 0.75rem;
  background: color-mix(in srgb, var(--journal-surface) 92%, var(--color-bg-base));
  color: var(--journal-muted);
  transition:
    border-color 150ms ease,
    background 150ms ease,
    color 150ms ease;
}

.topology-page--template-library .template-toolbar-refresh:hover {
  border-color: color-mix(in srgb, var(--journal-accent) 28%, transparent);
  color: var(--journal-accent);
}

.topology-page--template-library .topology-toolbar-btn--ghost,
.topology-page--template-library .template-action-btn {
  background: color-mix(in srgb, var(--journal-surface) 92%, var(--color-bg-base));
  color: var(--journal-ink);
}

.topology-page--template-library .topology-toolbar-btn--ghost:hover,
.topology-page--template-library .template-action-btn:hover {
  border-color: color-mix(in srgb, var(--journal-accent) 28%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 4%, var(--journal-surface));
  color: var(--journal-accent);
}

.topology-page--template-library .topology-toolbar-btn--primary,
.topology-page--template-library .template-action-btn--primary {
  border-color: transparent;
  background: var(--journal-accent);
  color: #fff;
}

.topology-page--template-library .topology-toolbar-btn--primary:hover,
.topology-page--template-library .template-action-btn--primary:hover {
  background: color-mix(in srgb, var(--journal-accent) 88%, black);
  color: #fff;
}

.topology-page--template-library .template-action-btn--danger {
  border-color: color-mix(in srgb, var(--color-danger) 28%, transparent);
  background: color-mix(in srgb, var(--color-danger) 10%, var(--journal-surface));
  color: color-mix(in srgb, var(--color-danger) 88%, var(--journal-ink));
}

.topology-page--template-library :deep(.page-header__eyebrow) {
  border-left-width: 0;
  padding-left: 0;
  color: var(--journal-accent);
}

.topology-page--template-library :deep(.section-card) {
  padding: var(--space-4-5) 0 var(--space-1);
  border-top: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.topology-page--template-library :deep(.section-card__header) {
  margin-bottom: var(--space-4);
  align-items: center;
  border: 0;
  border-radius: 0;
  background: transparent;
  padding: 0 0 0 var(--space-4);
  border-left: 2px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.topology-page--template-library :deep(.section-card__header h2) {
  font-size: var(--font-size-1-10);
  color: var(--journal-ink);
}

.topology-page--template-library :deep(.section-card__header p) {
  color: var(--journal-muted);
}

.topology-page--template-library :deep(.section-card__body) {
  padding-left: 0;
}

.topology-page--template-library :deep(.topology-hero-aside--library > section) {
  padding-left: 0;
  background: transparent;
}

.topology-page--template-library :deep(.topology-hero-aside--library > section h2) {
  font-size: var(--font-size-1-45);
}

.topology-page--template-library :deep(.topology-hero-aside--library > section p) {
  color: var(--journal-muted);
}

.topology-page--template-library :deep(.section-card__body > .rounded-2xl),
.topology-page--template-library :deep(.section-card__body > .rounded-xl),
.topology-page--template-library :deep([data-node-editor]),
.topology-page--template-library :deep(.topology-canvas-board__root) {
  box-shadow: none;
}

.topology-page--template-library :deep(.topology-canvas-board__root) {
  border: 0;
}

.topology-page--template-library :deep([data-node-editor]) {
  border-color: var(--journal-border);
  border-radius: 16px;
  background: color-mix(in srgb, var(--journal-surface) 94%, var(--color-bg-base));
}

.topology-page--template-library :deep([data-node-editor].border-primary) {
  border-color: color-mix(in srgb, var(--journal-accent) 42%, transparent);
}

.topology-page--template-library :deep(input),
.topology-page--template-library :deep(select),
.topology-page--template-library :deep(textarea) {
  border-color: var(--journal-border);
  background: var(--journal-surface);
  color: var(--journal-ink);
}

.topology-page--template-library :deep(input:focus),
.topology-page--template-library :deep(select:focus),
.topology-page--template-library :deep(textarea:focus) {
  border-color: color-mix(in srgb, var(--journal-accent) 48%, transparent);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--journal-accent) 14%, transparent);
  outline: none;
}

.topology-page--template-library :deep(.topology-canvas-board__surface) {
  border-color: color-mix(in srgb, var(--journal-border) 70%, transparent);
}

.topology-page--template-library .topology-workbench {
  gap: var(--space-5);
}

.topology-page--template-library .topology-side-stack--library {
  border: 0;
  background: transparent;
  box-shadow: none;
  padding: 0;
}

.topology-page--template-library .topology-side-stack--library :deep(.section-card:first-child) {
  padding-top: 0;
  border-top: 0;
}

.topology-page--template-library .template-search-row {
  display: grid;
  gap: var(--space-3);
  grid-template-columns: minmax(0, 1fr) auto;
}

.topology-page--template-library .template-library-list,
.topology-page--template-library .template-writeback-form,
.topology-page--template-library .template-boundary-list {
  display: grid;
  gap: 0;
}

.topology-page--template-library .template-directory-head {
  display: grid;
  grid-template-columns: minmax(0, 1.4fr) minmax(0, 0.95fr) auto;
  gap: var(--space-4);
  padding: 0 0 var(--space-3);
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  font-size: var(--font-size-0-72);
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--journal-muted);
}

.topology-page--template-library .template-toolbar-tabs,
.topology-page--template-library .template-canvas-mode-banner,
.topology-page--template-library .template-quick-editor {
  border: 0;
  box-shadow: none;
}

.topology-page--template-library .template-library-item {
  display: grid;
  grid-template-columns: minmax(0, 1.4fr) minmax(0, 0.95fr) auto;
  align-items: start;
  gap: var(--space-4);
  border-radius: 0;
  padding: var(--space-4) 0;
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  box-shadow: none;
  transition:
    border-color 150ms ease,
    background 150ms ease,
    color 150ms ease;
}

.topology-page--template-library .template-library-item--idle {
  border-left: 0;
  background: transparent;
}

.topology-page--template-library .template-library-item--idle:hover {
  background: color-mix(in srgb, var(--journal-accent) 4%, transparent);
}

.topology-page--template-library .template-library-item--active {
  border-left: 2px solid color-mix(in srgb, var(--journal-accent) 58%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 6%, transparent);
}

.topology-page--template-library .template-library-item__main {
  min-width: 0;
}

.topology-page--template-library .template-library-item__meta {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2) var(--space-3);
  align-content: start;
  padding-top: var(--space-0-5);
  font-size: var(--font-size-0-76);
  line-height: 1.6;
  color: var(--journal-muted);
}

.topology-page--template-library .template-library-item__actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: var(--space-2);
}

.topology-page--template-library .template-boundary-item {
  display: grid;
  gap: var(--space-1-5);
  padding: var(--space-4) 0 var(--space-4) var(--space-4);
  border-bottom: 1px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
  border-left: 2px solid color-mix(in srgb, var(--journal-border) 88%, transparent);
}

.topology-page--template-library .template-boundary-item__label {
  font-size: var(--font-size-0-72);
  font-weight: 700;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

.topology-page--template-library .template-boundary-item__copy {
  font-size: var(--font-size-0-88);
  line-height: 1.65;
  color: var(--journal-muted);
}

.topology-page--template-library .template-boundary-item--warning .template-boundary-item__label {
  color: var(--color-warning);
}

.topology-page--template-library .template-boundary-item--danger .template-boundary-item__label {
  color: var(--color-danger);
}

.topology-page--template-library .template-boundary-item--neutral .template-boundary-item__label {
  color: var(--journal-accent);
}

:global([data-theme='dark']) .topology-page--template-library {
  --journal-ink: var(--color-text-primary);
  --journal-muted: var(--color-text-secondary);
  --journal-accent: var(--color-primary-hover);
  --journal-border: color-mix(in srgb, var(--color-border-default) 84%, transparent);
  --journal-surface: color-mix(in srgb, var(--color-bg-surface) 92%, var(--color-bg-base));
  --journal-surface-subtle: color-mix(in srgb, var(--color-bg-surface) 78%, var(--color-bg-base));
}

:global([data-theme='dark']) .topology-page--template-library .template-library-main,
:global([data-theme='dark']) .topology-page--template-library :deep(.page-header) {
  background:
    radial-gradient(circle at top right, color-mix(in srgb, var(--journal-accent) 10%, transparent), transparent 18rem),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle) 94%, var(--color-bg-base))
    );
}

:global([data-theme='dark']) .topology-page--template-library .topology-toolbar-btn--ghost,
:global([data-theme='dark']) .topology-page--template-library .template-action-btn,
:global([data-theme='dark']) .topology-page--template-library :deep([data-node-editor]),
:global([data-theme='dark']) .topology-page--template-library :deep(input),
:global([data-theme='dark']) .topology-page--template-library :deep(select),
:global([data-theme='dark']) .topology-page--template-library :deep(textarea) {
  background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
}

@media (max-width: 1023px) {
  .topology-page--challenge .topology-summary-grid--challenge {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .topology-page--challenge .content-pane.topology-workspace {
    grid-template-columns: minmax(0, 1fr);
  }

  .topology-page--challenge .topology-context-rail {
    padding-left: 0;
    padding-top: var(--space-6);
    border-top: 1px solid var(--topology-divider);
    border-left: 0;
  }

  .topology-page--challenge .topology-context-stack {
    position: static;
  }

  .topology-page--template-library .topology-summary-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .topology-page--template-library .template-directory-head {
    display: none;
  }

  .topology-page--template-library .template-library-item {
    grid-template-columns: minmax(0, 1fr);
  }

  .topology-page--template-library .template-library-item__actions {
    justify-content: flex-start;
  }

  .topology-page--template-library .topology-hero-aside--library {
    border-left: 0;
    padding-left: 0;
  }

  .topology-page--template-library :deep(.topology-hero-aside--library > section) {
    padding-left: 0;
    background: transparent;
  }
}

@media (max-width: 767px) {
  .topology-page--challenge {
    padding: var(--space-5);
    border-radius: 24px;
  }

  .topology-page--challenge .workspace-topbar {
    align-items: flex-start;
    padding-bottom: var(--space-5);
  }

  .topology-page--challenge .topology-summary-grid--challenge {
    grid-template-columns: 1fr;
  }

  .topology-page--template-library .topology-summary-grid {
    grid-template-columns: 1fr;
  }

  .topology-page--template-library .template-toolbar-tabs {
    gap: var(--space-3);
    overflow-x: auto;
  }

  .topology-page--template-library .template-search-row {
    grid-template-columns: 1fr;
  }

  .topology-page--template-library .topology-hero-lead--library {
    padding: 0;
  }
}
</style>
