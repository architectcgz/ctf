<script setup lang="ts">
import { computed, defineAsyncComponent, ref } from 'vue'
import {
  Plus,
  RefreshCw,
  Save,
  Trash2,
  Layout,
  Server,
  Network,
  ShieldCheck,
} from 'lucide-vue-next'

import {
  useChallengeTopologyStudioPage,
  type TopologyStudioMode,
} from '@/composables/useChallengeTopologyStudioPage'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import SectionCard from '@/components/common/SectionCard.vue'
import TopologyConnectivitySections from './TopologyConnectivitySections.vue'
import TopologyNetworkSection from './TopologyNetworkSection.vue'
import TopologyStatusNotes from './TopologyStatusNotes.vue'
import TopologySummaryGrid from './TopologySummaryGrid.vue'
import TopologyTemplateSidePanel from './TopologyTemplateSidePanel.vue'

import type { CanvasInteractionMode } from './TopologyCanvasBoard.vue'
import type { TopologyLinkDraft, TopologyNetworkDraft, TopologyPolicyDraft } from './topologyDraft'

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

const rootClasses = computed(() => [
  'topology-page',
  isTemplateLibraryMode.value ? 'topology-page--template-library' : 'topology-page--challenge',
  'workspace-shell journal-shell journal-shell-admin journal-notes-card journal-hero flex min-h-full flex-1 flex-col',
])

function updateNetworkDraft(payload: {
  uid: string
  patch: Partial<Pick<TopologyNetworkDraft, 'key' | 'name' | 'cidr' | 'internal'>>
}) {
  const network = draft.value.networks.find((item) => item.uid === payload.uid)
  if (!network) return
  Object.assign(network, payload.patch)
}

function updateLinkDraft(payload: {
  uid: string
  patch: Partial<Pick<TopologyLinkDraft, 'from_node_key' | 'to_node_key'>>
}) {
  const link = draft.value.links.find((item) => item.uid === payload.uid)
  if (!link) return
  Object.assign(link, payload.patch)
}

function removeLinkDraft(uid: string) {
  draft.value.links = draft.value.links.filter((item) => item.uid !== uid)
}

function updatePolicyDraft(payload: {
  uid: string
  patch: Partial<Pick<TopologyPolicyDraft, 'source_node_key' | 'target_node_key' | 'action'>>
}) {
  const policy = draft.value.policies.find((item) => item.uid === payload.uid)
  if (!policy) return
  Object.assign(policy, payload.patch)
}

function removePolicyDraft(uid: string) {
  draft.value.policies = draft.value.policies.filter((item) => item.uid !== uid)
}
</script>

<template>
  <div :class="rootClasses">
    <PageHeader
      v-if="isTemplateLibraryMode"
      class="topology-page-header"
      :eyebrow="pageHeader.eyebrow"
      :title="pageHeader.title"
      :description="pageHeader.description"
    >
      <button
        type="button"
        class="topology-toolbar-btn topology-toolbar-btn--primary"
        @click="handleResetTemplateEditor"
      >
        <Plus class="h-4 w-4" />
        新建空白模板
      </button>
      <button
        type="button"
        class="topology-toolbar-btn topology-toolbar-btn--ghost"
        @click="void reloadAll()"
      >
        <RefreshCw class="h-4 w-4" />
        刷新
      </button>
    </PageHeader>

    <header v-else class="workspace-topbar topology-workspace-topbar">
      <div class="topology-topbar-leading">
        <span class="workspace-overline">Challenge Workspace</span>
        <span class="topology-topbar-chip">{{ pageHeader.eyebrow }}</span>
      </div>
      <div class="topology-topbar-actions">
        <button
          type="button"
          class="ui-btn ui-btn--ghost topology-action-btn"
          @click="emit('back')"
        >
          返回题目详情
        </button>
        <button
          type="button"
          class="ui-btn ui-btn--ghost topology-action-btn"
          @click="void reloadAll()"
        >
          <RefreshCw class="h-4 w-4" />
          刷新
        </button>
        <button
          type="button"
          class="ui-btn ui-btn--primary topology-action-btn"
          :disabled="saving"
          @click="void handleSaveTopology()"
        >
          <Save class="h-4 w-4" />
          {{ saving ? '保存中...' : '保存拓扑' }}
        </button>
      </div>
    </header>

    <div v-if="loading && !isTemplateLibraryMode" class="content-pane topology-loading-pane">
      <AppLoading>{{ loadingText }}</AppLoading>
    </div>

    <section
      v-else-if="loading && isTemplateLibraryMode"
      class="content-pane template-library-main"
    >
      <div class="flex justify-center py-12">
        <AppLoading>{{ loadingText }}</AppLoading>
      </div>
    </section>

    <template v-else>
      <section v-if="isTemplateLibraryMode" class="content-pane template-library-main">
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

            <TopologySummaryGrid :summary="topologySummary" mode="template-library" />
          </div>

          <TopologyStatusNotes
            mode="template-library"
            :status-card="statusCard"
            :secondary-card="secondaryCard"
          />
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
                  @click="activeWorkbenchTab = tab.id as any"
                >
                  <component :is="tab.icon" class="h-4 w-4" />
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

            <div v-if="activeWorkbenchTab === 'visual'" class="space-y-6">
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
                        ? 'topology-mode-btn--allow-active'
                        : 'topology-mode-btn--allow-idle'
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
                      ? 'topology-validation-banner--ok'
                      : 'topology-validation-banner--warn'
                  "
                >
                  <div class="font-medium">
                    {{ draftValidationIssues.length === 0 ? '基础校验已通过' : '基础校验发现问题' }}
                  </div>
                  <ul v-if="draftValidationIssues.length > 0" class="mt-2 space-y-1 text-xs">
                    <li v-for="issue in draftValidationIssues" :key="issue">
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
                  <div
                    class="template-quick-editor rounded-2xl border border-border bg-elevated p-4"
                  >
                    <div class="text-sm font-semibold text-text-primary">画布快速编辑</div>

                    <div
                      v-if="!selectedNodeDraft && !selectedEdgeMeta"
                      class="mt-3 rounded-xl border border-dashed border-border px-4 py-6 text-sm text-text-muted"
                    >
                      请在画布中选择一个节点或连线进行快速配置
                    </div>

                    <div v-else-if="selectedNodeDraft" class="mt-3 space-y-4">
                      <div class="grid gap-3 md:grid-cols-2">
                        <label class="space-y-2">
                          <span class="text-sm text-text-secondary">节点名称</span>
                          <input
                            v-model="selectedNodeDraft.name"
                            type="text"
                            class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
                          />
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
                          />
                        </label>
                      </div>

                      <div class="space-y-2">
                        <div class="text-sm text-text-secondary">所属网络</div>
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
                            />
                            <span>{{ network.name || network.key }}</span>
                          </label>
                        </div>
                      </div>
                    </div>

                    <div v-else-if="selectedEdgeMeta" class="mt-3 space-y-4">
                      <div class="grid gap-3 md:grid-cols-2">
                        <label class="space-y-2">
                          <span class="text-sm text-text-secondary">源节点</span>
                          <select
                            :value="selectedEdgeSourceKey"
                            class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
                            @change="
                              updateSelectedEdgeSourceKey(
                                ($event.target as HTMLSelectElement).value
                              )
                            "
                          >
                            <option v-for="node in nodeOptions" :key="node.key" :value="node.key">
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
                              updateSelectedEdgeTargetKey(
                                ($event.target as HTMLSelectElement).value
                              )
                            "
                          >
                            <option v-for="node in nodeOptions" :key="node.key" :value="node.key">
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

            <div v-else-if="activeWorkbenchTab === 'compute'" class="space-y-6">
              <SectionCard title="入口节点" subtitle="实例访问入口和当前草稿的保存范围。">
                <div class="grid gap-4">
                  <label class="space-y-2">
                    <span class="text-sm text-text-secondary">入口节点</span>
                    <select
                      v-model="draft.entry_node_key"
                      class="w-full rounded-xl border border-border bg-elevated px-3 py-3 text-sm text-text-primary outline-none transition focus:border-primary"
                    >
                      <option v-for="node in nodeOptions" :key="node.key" :value="node.key">
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
                    class="topology-toolbar-btn topology-toolbar-btn--ghost"
                    @click="addNode"
                  >
                    <Plus class="h-4 w-4" />
                    添加节点
                  </button>
                </template>
              </SectionCard>
            </div>

            <div v-else-if="activeWorkbenchTab === 'network'" class="space-y-6">
              <TopologyNetworkSection
                :networks="draft.networks"
                add-button-class="topology-toolbar-btn topology-toolbar-btn--ghost"
                @add-network="addNetwork"
                @remove-network="removeNetwork"
                @update-network="updateNetworkDraft"
              />
            </div>

            <div v-else-if="activeWorkbenchTab === 'policy'" class="space-y-6">
              <TopologyConnectivitySections
                :links="draft.links"
                :policies="draft.policies"
                :node-options="nodeOptions"
                add-button-class="topology-toolbar-btn topology-toolbar-btn--ghost"
                @add-link="addLink"
                @remove-link="removeLinkDraft"
                @update-link="updateLinkDraft"
                @add-policy="addPolicy"
                @remove-policy="removePolicyDraft"
                @update-policy="updatePolicyDraft"
              />
            </div>
          </div>

          <TopologyTemplateSidePanel
            v-model:template-keyword="templateKeyword"
            v-model:template-name="templateName"
            v-model:template-description="templateDescription"
            :is-template-library-mode="isTemplateLibraryMode"
            :selected-template-summary="selectedTemplateSummary"
            :selected-template-id="selectedTemplateId"
            :templates="templates"
            :template-busy="templateBusy"
            @load-template="loadTemplateIntoDraft"
            @clear-template-selection="clearTemplateSelection"
            @search-templates="void loadTemplates()"
            @reset-template-form="resetTemplateForm"
            @apply-template="(template) => void handleApplyTemplate(template)"
            @delete-template="(templateId) => void handleDeleteTemplate(templateId)"
            @reset-template-editor="handleResetTemplateEditor"
            @create-template="void handleCreateTemplate()"
            @update-template="void handleUpdateTemplate()"
          />
        </section>
      </section>

      <template v-else>
        <section class="workspace-tab-heading topology-page-heading">
          <div class="workspace-tab-heading__main">
            <div class="topology-page-kicker">
              {{ heroEyebrow }}
            </div>
            <h1 class="hero-title">
              {{ heroTitle }}
            </h1>
          </div>
          <p class="workspace-page-copy topology-page-copy">
            {{ heroDescription }}
          </p>

          <TopologySummaryGrid :summary="topologySummary" mode="challenge" />
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
                      ? 'topology-mode-btn--allow-active'
                      : 'topology-mode-btn--allow-idle'
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
                    ? 'topology-validation-banner--ok'
                    : 'topology-validation-banner--warn'
                "
              >
                <div class="font-medium">
                  {{ draftValidationIssues.length === 0 ? '基础校验已通过' : '基础校验发现问题' }}
                </div>
                <div
                  v-if="draftValidationIssues.length === 0"
                  class="topology-validation-hint topology-validation-hint--success mt-1 text-xs"
                >
                  当前草稿的入口、节点、网络和链路引用关系正常。
                </div>
                <ul v-else class="mt-2 space-y-1 text-xs">
                  <li v-for="issue in draftValidationIssues" :key="issue">
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
                  <div class="text-sm font-semibold text-text-primary">画布快速编辑</div>

                  <div
                    v-if="!selectedNodeDraft && !selectedEdgeMeta"
                    class="mt-3 rounded-xl border border-dashed border-border px-4 py-6 text-sm text-text-muted"
                  >
                    请选择一个节点或一条边
                  </div>

                  <div v-else-if="selectedNodeDraft" class="mt-3 space-y-4">
                    <div class="grid gap-3 md:grid-cols-2">
                      <label class="space-y-2">
                        <span class="text-sm text-text-secondary">节点名称</span>
                        <input
                          v-model="selectedNodeDraft.name"
                          type="text"
                          class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
                        />
                      </label>
                      <label class="space-y-2">
                        <span class="text-sm text-text-secondary">镜像</span>
                        <select
                          v-model="selectedNodeDraft.image_id"
                          class="w-full rounded-xl border border-border bg-surface px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
                        >
                          <option value="">复用题目主镜像</option>
                          <option v-for="image in images" :key="image.id" :value="image.id">
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
                        />
                      </label>
                    </div>

                    <label
                      class="flex items-center gap-3 rounded-xl border border-border bg-surface px-3 py-3 text-sm text-text-primary"
                    >
                      <input
                        v-model="selectedNodeDraft.inject_flag"
                        type="checkbox"
                        class="h-4 w-4 rounded border-border bg-transparent"
                      />
                      启用 Flag 注入
                    </label>

                    <div class="space-y-2">
                      <div class="text-sm text-text-secondary">所属网络</div>
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
                          />
                          <span>{{ network.name || network.key }}</span>
                        </label>
                      </div>
                    </div>
                  </div>

                  <div v-else-if="selectedEdgeMeta" class="mt-3 space-y-4">
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
                          <option v-for="node in nodeOptions" :key="node.key" :value="node.key">
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
                          <option v-for="node in nodeOptions" :key="node.key" :value="node.key">
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
                  <div class="text-sm font-semibold text-text-primary">网络快速编辑</div>
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
                      />
                      <input
                        v-model="network.name"
                        type="text"
                        class="w-full rounded-xl border border-border bg-elevated px-3 py-2.5 text-sm text-text-primary outline-none transition focus:border-primary"
                        placeholder="网络名称"
                      />
                      <label
                        class="flex items-center gap-2 rounded-xl border border-border bg-elevated px-3 py-2.5 text-sm text-text-primary"
                      >
                        <input
                          v-model="network.internal"
                          type="checkbox"
                          class="h-4 w-4 rounded border-border bg-transparent"
                        />
                        internal
                      </label>
                    </div>
                  </div>
                </div>
              </div>
            </SectionCard>

            <SectionCard title="入口节点" subtitle="实例访问入口和当前草稿的保存范围。">
              <div class="grid gap-4 md:grid-cols-[1fr_auto]">
                <label class="space-y-2">
                  <span class="text-sm text-text-secondary">入口节点</span>
                  <select
                    v-model="draft.entry_node_key"
                    class="w-full rounded-xl border border-border bg-elevated px-3 py-3 text-sm text-text-primary outline-none transition focus:border-primary"
                  >
                    <option v-for="node in nodeOptions" :key="node.key" :value="node.key">
                      {{ node.label }} ({{ node.key }})
                    </option>
                  </select>
                </label>

                <button
                  v-if="!isTemplateLibraryMode"
                  type="button"
                  class="ui-btn ui-btn--danger self-end"
                  :disabled="saving || !topology"
                  @click="void handleDeleteTopology()"
                >
                  <Trash2 class="h-4 w-4" />
                  删除已保存拓扑
                </button>
              </div>
            </SectionCard>

            <TopologyNetworkSection
              :networks="draft.networks"
              add-button-class="ui-btn ui-btn--ghost topology-action-btn"
              @add-network="addNetwork"
              @remove-network="removeNetwork"
              @update-network="updateNetworkDraft"
            />

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
                  class="ui-btn ui-btn--ghost topology-action-btn"
                  @click="addNode"
                >
                  <Plus class="h-4 w-4" />
                  添加节点
                </button>
              </template>
            </SectionCard>

            <TopologyConnectivitySections
              :links="draft.links"
              :policies="draft.policies"
              :node-options="nodeOptions"
              add-button-class="ui-btn ui-btn--ghost topology-action-btn"
              @add-link="addLink"
              @remove-link="removeLinkDraft"
              @update-link="updateLinkDraft"
              @add-policy="addPolicy"
              @remove-policy="removePolicyDraft"
              @update-policy="updatePolicyDraft"
            />
          </div>

          <aside class="context-rail topology-context-rail">
            <div class="topology-context-stack">
              <TopologyStatusNotes
                mode="challenge"
                :status-card="statusCard"
                :secondary-card="secondaryCard"
              />

              <TopologyTemplateSidePanel
                v-model:template-keyword="templateKeyword"
                v-model:template-name="templateName"
                v-model:template-description="templateDescription"
                :is-template-library-mode="isTemplateLibraryMode"
                :selected-template-summary="selectedTemplateSummary"
                :selected-template-id="selectedTemplateId"
                :templates="templates"
                :template-busy="templateBusy"
                @load-template="loadTemplateIntoDraft"
                @clear-template-selection="clearTemplateSelection"
                @search-templates="void loadTemplates()"
                @reset-template-form="resetTemplateForm"
                @apply-template="(template) => void handleApplyTemplate(template)"
                @delete-template="(templateId) => void handleDeleteTemplate(templateId)"
                @reset-template-editor="handleResetTemplateEditor"
                @create-template="void handleCreateTemplate()"
                @update-template="void handleUpdateTemplate()"
              />
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
  --topology-panel-subtle: color-mix(
    in srgb,
    var(--journal-surface-subtle) 96%,
    var(--color-bg-base)
  );
  --topology-divider: color-mix(in srgb, var(--journal-border) 88%, transparent);
  display: flex;
  flex-direction: column;
  min-height: max(100%, calc(100vh - 5rem));
  padding: var(--space-6) var(--space-7);
  background:
    radial-gradient(
      circle at top right,
      color-mix(in srgb, var(--journal-accent) 8%, transparent),
      transparent 22rem
    ),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 98%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle) 96%, var(--color-bg-base))
    );
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
  min-height: max(100%, calc(100vh - 5rem));
  padding: var(--space-6) var(--space-7);
  background: color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base));
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

.topology-page--challenge .topology-action-btn,
.topology-page--template-library .topology-action-btn {
  --ui-btn-height: 2.45rem;
  --ui-btn-padding: var(--space-2) var(--space-4);
  --ui-btn-radius: 0.75rem;
  --ui-btn-font-size: var(--font-size-0-84);
}

.topology-page--challenge .topology-action-btn {
  --ui-btn-secondary-border: var(--journal-border);
  --ui-btn-secondary-background: color-mix(
    in srgb,
    var(--journal-surface) 94%,
    var(--color-bg-base)
  );
  --ui-btn-secondary-color: var(--journal-ink);
  --ui-btn-secondary-hover-border: color-mix(in srgb, var(--journal-accent) 28%, transparent);
  --ui-btn-secondary-hover-background: color-mix(
    in srgb,
    var(--journal-accent) 4%,
    var(--journal-surface)
  );
  --ui-btn-secondary-hover-color: var(--journal-accent);
  --ui-btn-ghost-color: var(--journal-ink);
  --ui-btn-ghost-hover-color: var(--journal-accent);
  --ui-btn-ghost-hover-background: color-mix(
    in srgb,
    var(--journal-accent) 4%,
    var(--journal-surface)
  );
  --ui-btn-primary-border: transparent;
  --ui-btn-primary-background: var(--journal-accent);
  --ui-btn-primary-color: var(--color-bg-base);
  --ui-btn-primary-hover-background: color-mix(
    in srgb,
    var(--journal-accent) 88%,
    var(--color-bg-base)
  );
  --ui-btn-primary-hover-shadow: 0 12px 28px
    color-mix(in srgb, var(--journal-accent) 16%, transparent);
  --ui-btn-danger-border: color-mix(in srgb, var(--color-danger) 28%, transparent);
  --ui-btn-danger-background: color-mix(in srgb, var(--color-danger) 10%, var(--journal-surface));
  --ui-btn-danger-color: color-mix(in srgb, var(--color-danger) 88%, var(--journal-ink));
  --ui-btn-danger-hover-border: color-mix(in srgb, var(--color-danger) 34%, transparent);
  --ui-btn-danger-hover-background: color-mix(
    in srgb,
    var(--color-danger) 14%,
    var(--journal-surface)
  );
  --ui-btn-focus-ring: color-mix(in srgb, var(--journal-accent) 18%, transparent);
}

.topology-page--challenge .topology-action-btn:disabled {
  opacity: 0.65;
  cursor: not-allowed;
  box-shadow: none;
}

.topology-action-btn--icon {
  min-width: 2.75rem;
  padding-inline: var(--space-3);
}

.topology-toolbar-btn {
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
    background-color 150ms ease,
    color 150ms ease;
}

.topology-toolbar-btn--ghost {
  background: color-mix(in srgb, var(--journal-surface) 94%, var(--color-bg-base));
  color: var(--journal-ink);
}

.topology-toolbar-btn--ghost:hover {
  border-color: color-mix(in srgb, var(--journal-accent) 28%, transparent);
  background: color-mix(in srgb, var(--journal-accent) 4%, var(--journal-surface));
  color: var(--journal-accent);
}

.topology-toolbar-btn--primary {
  border-color: transparent;
  background: var(--journal-accent);
  color: var(--color-bg-base);
}

.topology-toolbar-btn--primary:hover {
  background: color-mix(in srgb, var(--journal-accent) 88%, var(--color-bg-base));
}

.topology-mode-btn--allow-active {
  border-color: var(--color-success);
  background: color-mix(in srgb, var(--color-success) 10%, transparent);
  color: var(--color-success);
}

.topology-mode-btn--allow-idle {
  border-color: var(--color-border-default);
  color: var(--color-text-primary);
}

.topology-mode-btn--allow-idle:hover {
  border-color: color-mix(in srgb, var(--color-success) 60%, var(--color-border-default));
}

.topology-validation-banner--ok {
  border-color: color-mix(in srgb, var(--color-success) 20%, var(--color-border-default));
  background: color-mix(in srgb, var(--color-success) 10%, transparent);
  color: var(--color-success);
}

.topology-validation-banner--warn {
  border-color: color-mix(in srgb, var(--color-warning) 20%, var(--color-border-default));
  background: color-mix(in srgb, var(--color-warning) 10%, transparent);
  color: var(--color-warning);
}

.topology-validation-hint--success {
  color: color-mix(in srgb, var(--color-success) 80%, transparent);
}

.topology-page--challenge .topology-page-heading {
  display: grid;
  gap: var(--space-4);
  padding-bottom: var(--space-6);
}

.topology-page--challenge .topology-page-copy {
  max-width: 48rem;
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

.topology-page--template-library .template-library-main,
.topology-page--template-library :deep(.page-header) {
  border-color: var(--journal-border);
  border-radius: 0;
  background:
    radial-gradient(
      circle at top right,
      color-mix(in srgb, var(--journal-accent) 7%, transparent),
      transparent 22rem
    ),
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

.topology-page--template-library .topology-action-btn {
  --ui-btn-font-size: var(--font-size-0-82);
  --ui-btn-secondary-border: var(--journal-border);
  --ui-btn-secondary-background: color-mix(
    in srgb,
    var(--journal-surface) 92%,
    var(--color-bg-base)
  );
  --ui-btn-secondary-color: var(--journal-ink);
  --ui-btn-secondary-hover-border: color-mix(in srgb, var(--journal-accent) 28%, transparent);
  --ui-btn-secondary-hover-background: color-mix(
    in srgb,
    var(--journal-accent) 4%,
    var(--journal-surface)
  );
  --ui-btn-secondary-hover-color: var(--journal-accent);
  --ui-btn-ghost-color: var(--journal-ink);
  --ui-btn-ghost-hover-color: var(--journal-accent);
  --ui-btn-ghost-hover-background: color-mix(
    in srgb,
    var(--journal-accent) 4%,
    var(--journal-surface)
  );
  --ui-btn-primary-border: transparent;
  --ui-btn-primary-background: var(--journal-accent);
  --ui-btn-primary-hover-background: color-mix(
    in srgb,
    var(--journal-accent) 88%,
    var(--color-bg-base)
  );
  --ui-btn-primary-hover-color: var(--color-bg-base);
  --ui-btn-primary-color: var(--color-bg-base);
  --ui-btn-danger-border: color-mix(in srgb, var(--color-danger) 28%, transparent);
  --ui-btn-danger-background: color-mix(in srgb, var(--color-danger) 10%, var(--journal-surface));
  --ui-btn-danger-color: color-mix(in srgb, var(--color-danger) 88%, var(--journal-ink));
  --ui-btn-danger-hover-border: color-mix(in srgb, var(--color-danger) 34%, transparent);
  --ui-btn-danger-hover-background: color-mix(
    in srgb,
    var(--color-danger) 14%,
    var(--journal-surface)
  );
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

.topology-page--template-library .template-toolbar-tabs,
.topology-page--template-library .template-canvas-mode-banner,
.topology-page--template-library .template-quick-editor {
  border: 0;
  box-shadow: none;
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
    radial-gradient(
      circle at top right,
      color-mix(in srgb, var(--journal-accent) 10%, transparent),
      transparent 18rem
    ),
    linear-gradient(
      180deg,
      color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base)),
      color-mix(in srgb, var(--journal-surface-subtle) 94%, var(--color-bg-base))
    );
}

:global([data-theme='dark']) .topology-page--template-library .topology-action-btn,
:global([data-theme='dark']) .topology-page--template-library :deep([data-node-editor]),
:global([data-theme='dark']) .topology-page--template-library :deep(input),
:global([data-theme='dark']) .topology-page--template-library :deep(select),
:global([data-theme='dark']) .topology-page--template-library :deep(textarea) {
  background: color-mix(in srgb, var(--journal-surface) 94%, transparent);
}

@media (max-width: 1023px) {
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
}

@media (max-width: 767px) {
  .topology-page--challenge {
    padding: var(--space-5);
  }

  .topology-page--challenge .workspace-topbar {
    align-items: flex-start;
    padding-bottom: var(--space-5);
  }

  .topology-page--template-library .template-toolbar-tabs {
    gap: var(--space-3);
    overflow-x: auto;
  }

  .topology-page--template-library .topology-hero-lead--library {
    padding: 0;
  }
}
</style>
