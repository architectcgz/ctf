<script setup lang="ts">
import { computed } from 'vue'
import { Blocks, GitBranch, Link2, Plus, RefreshCw, Save, ShieldBan, Trash2 } from 'lucide-vue-next'

import {
  useChallengeTopologyStudioPage,
  type TopologyStudioMode,
} from '@/composables/useChallengeTopologyStudioPage'
import AppCard from '@/components/common/AppCard.vue'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppLoading from '@/components/common/AppLoading.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import SectionCard from '@/components/common/SectionCard.vue'

import TopologyCanvasBoard from './TopologyCanvasBoard.vue'
import type { CanvasInteractionMode } from './TopologyCanvasBoard.vue'
import TopologyNodeEditor from './TopologyNodeEditor.vue'

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
  <div class="space-y-6">
    <PageHeader
      :eyebrow="pageHeader.eyebrow"
      :title="pageHeader.title"
      :description="pageHeader.description"
    >
      <button
        v-if="!isTemplateLibraryMode"
        type="button"
        class="inline-flex items-center gap-2 rounded-xl border border-border px-4 py-2 text-sm font-medium text-text-primary transition hover:border-primary"
        @click="emit('back')"
      >
        返回挑战详情
      </button>
      <button
        v-else
        type="button"
        class="inline-flex items-center gap-2 rounded-xl border border-border px-4 py-2 text-sm font-medium text-text-primary transition hover:border-primary"
        @click="handleResetTemplateEditor"
      >
        <Plus class="h-4 w-4" />
        新建空白模板
      </button>
      <button
        type="button"
        class="inline-flex items-center gap-2 rounded-xl border border-border px-4 py-2 text-sm font-medium text-text-primary transition hover:border-primary"
        @click="void reloadAll()"
      >
        <RefreshCw class="h-4 w-4" />
        刷新
      </button>
      <button
        v-if="!isTemplateLibraryMode"
        type="button"
        class="inline-flex items-center gap-2 rounded-xl bg-primary px-4 py-2 text-sm font-medium text-white transition hover:opacity-90"
        :disabled="saving"
        @click="void handleSaveTopology()"
      >
        <Save class="h-4 w-4" />
        {{ saving ? '保存中...' : '保存拓扑' }}
      </button>
    </PageHeader>

    <div
      v-if="loading"
      class="flex justify-center py-16"
    >
      <AppLoading>{{ loadingText }}</AppLoading>
    </div>

    <template v-else>
      <section class="grid gap-4 xl:grid-cols-[1.04fr_0.96fr]">
        <div
          class="rounded-[30px] border border-primary/20 bg-[linear-gradient(145deg,rgba(8,145,178,0.22),rgba(15,23,42,0.94))] p-6 shadow-[0_24px_70px_var(--color-shadow-soft)]"
        >
          <div
            class="flex flex-wrap items-center gap-2 text-[11px] font-semibold uppercase tracking-[0.22em] text-cyan-100/75"
          >
            <span>{{ heroEyebrow }}</span>
            <span class="rounded-full border border-white/10 bg-white/5 px-2 py-1">真实接口</span>
          </div>
          <h2 class="mt-3 text-3xl font-semibold tracking-tight text-white">
            {{ heroTitle }}
          </h2>
          <p class="mt-3 max-w-3xl text-sm leading-7 text-cyan-50/80">
            {{ heroDescription }}
          </p>

          <div class="mt-6 grid gap-3 md:grid-cols-4">
            <div class="rounded-[24px] border border-white/10 bg-white/6 px-4 py-4">
              <div class="text-[11px] uppercase tracking-[0.18em] text-cyan-100/60">
                网络
              </div>
              <div class="mt-2 text-2xl font-semibold text-white">
                {{ topologySummary.networks }}
              </div>
            </div>
            <div class="rounded-[24px] border border-white/10 bg-white/6 px-4 py-4">
              <div class="text-[11px] uppercase tracking-[0.18em] text-cyan-100/60">
                节点
              </div>
              <div class="mt-2 text-2xl font-semibold text-white">
                {{ topologySummary.nodes }}
              </div>
            </div>
            <div class="rounded-[24px] border border-white/10 bg-white/6 px-4 py-4">
              <div class="text-[11px] uppercase tracking-[0.18em] text-cyan-100/60">
                连线
              </div>
              <div class="mt-2 text-2xl font-semibold text-white">
                {{ topologySummary.links }}
              </div>
            </div>
            <div class="rounded-[24px] border border-white/10 bg-white/6 px-4 py-4">
              <div class="text-[11px] uppercase tracking-[0.18em] text-cyan-100/60">
                策略
              </div>
              <div class="mt-2 text-2xl font-semibold text-white">
                {{ topologySummary.policies }}
              </div>
            </div>
          </div>
        </div>

        <div class="grid gap-3 md:grid-cols-3 xl:grid-cols-1">
          <AppCard
            variant="metric"
            accent="primary"
            :eyebrow="statusCard.eyebrow"
            :title="statusCard.title"
            :subtitle="statusCard.subtitle"
          >
            <template #header>
              <div
                class="flex h-11 w-11 items-center justify-center rounded-2xl border border-primary/20 bg-primary/12 text-primary"
              >
                <Blocks class="h-5 w-5" />
              </div>
            </template>
          </AppCard>

          <AppCard
            variant="metric"
            accent="warning"
            :eyebrow="secondaryCard.eyebrow"
            :title="secondaryCard.title"
            :subtitle="secondaryCard.subtitle"
          >
            <template #header>
              <div
                class="flex h-11 w-11 items-center justify-center rounded-2xl border border-amber-500/20 bg-amber-500/12 text-amber-300"
              >
                <GitBranch class="h-5 w-5" />
              </div>
            </template>
          </AppCard>

          <AppCard
            variant="metric"
            accent="danger"
            eyebrow="运行时约束"
            title="粗粒度"
            subtitle="当前只支持节点级 allow/deny，不支持端口级 ACL。"
          >
            <template #header>
              <div
                class="flex h-11 w-11 items-center justify-center rounded-2xl border border-danger/20 bg-danger/12 text-danger"
              >
                <ShieldBan class="h-5 w-5" />
              </div>
            </template>
          </AppCard>
        </div>
      </section>

      <section class="grid gap-6 xl:grid-cols-[1.18fr_0.82fr]">
        <div class="space-y-6">
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
                    ? 'border-emerald-500 bg-emerald-500/10 text-emerald-300'
                    : 'border-border text-text-primary hover:border-emerald-500/60'
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
                  ? 'border-emerald-500/20 bg-emerald-500/10 text-emerald-200'
                  : 'border-amber-500/20 bg-amber-500/10 text-amber-100'
              "
            >
              <div class="font-medium">
                {{ draftValidationIssues.length === 0 ? '基础校验已通过' : '基础校验发现问题' }}
              </div>
              <div
                v-if="draftValidationIssues.length === 0"
                class="mt-1 text-xs text-emerald-100/80"
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
                        <option value="">复用挑战主镜像</option>
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

        <div class="space-y-6">
          <SectionCard
            title="模板库"
            :subtitle="
              isTemplateLibraryMode
                ? '从模板库载入后可直接编辑并覆盖模板，或另存为新模板。'
                : '可按模板快速回填编辑器，或直接应用到挑战。'
            "
          >
            <div class="space-y-3">
              <div class="rounded-2xl border border-border bg-elevated px-4 py-4">
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

              <div class="grid gap-3 md:grid-cols-[1fr_auto]">
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
                class="rounded-xl border border-dashed border-border px-4 py-6 text-sm text-text-muted"
              >
                当前没有模板数据
              </div>

              <div
                v-else
                class="space-y-3"
              >
                <article
                  v-for="template in templates"
                  :key="template.id"
                  class="rounded-2xl border p-4 transition"
                  :class="
                    selectedTemplateId === template.id
                      ? 'border-primary bg-primary/8'
                      : 'border-border bg-elevated'
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
                      class="rounded-xl border border-border px-3 py-2 text-xs font-medium text-text-primary transition hover:border-primary"
                      @click="loadTemplateIntoDraft(template)"
                    >
                      {{ isTemplateLibraryMode ? '载入编辑' : '载入草稿' }}
                    </button>
                    <button
                      v-if="!isTemplateLibraryMode"
                      type="button"
                      class="rounded-xl border border-border px-3 py-2 text-xs font-medium text-text-primary transition hover:border-primary"
                      @click="resetTemplateForm(template)"
                    >
                      选中
                    </button>
                    <button
                      v-if="!isTemplateLibraryMode"
                      type="button"
                      class="rounded-xl bg-primary px-3 py-2 text-xs font-medium text-white transition hover:opacity-90"
                      :disabled="templateBusy"
                      @click="void handleApplyTemplate(template)"
                    >
                      应用到挑战
                    </button>
                    <button
                      type="button"
                      class="rounded-xl border border-danger/30 bg-danger/10 px-3 py-2 text-xs font-medium text-danger transition hover:bg-danger/15"
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
            <div class="space-y-4">
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
                  class="inline-flex items-center gap-2 rounded-xl border border-border px-4 py-2 text-sm font-medium text-text-primary transition hover:border-primary"
                  @click="handleResetTemplateEditor"
                >
                  新建空白草稿
                </button>
                <button
                  type="button"
                  class="inline-flex items-center gap-2 rounded-xl bg-primary px-4 py-2 text-sm font-medium text-white transition hover:opacity-90"
                  :disabled="templateBusy"
                  @click="void handleCreateTemplate()"
                >
                  <Plus class="h-4 w-4" />
                  保存为新模板
                </button>
                <button
                  type="button"
                  class="inline-flex items-center gap-2 rounded-xl border border-border px-4 py-2 text-sm font-medium text-text-primary transition hover:border-primary"
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
          </SectionCard>
        </div>
      </section>

      <AppEmpty
        v-if="!challenge"
        title="挑战不存在"
        description="无法读取当前挑战的基础信息，请返回挑战列表后重试。"
        icon="Blocks"
      />
    </template>
  </div>
</template>
