import { onBeforeUnmount, onMounted, watch, type Ref } from 'vue'

import {
  normalizeCanvasPositions,
  type CanvasNodePosition,
} from './topologyLayout'
import type { TopologyEditorDraft } from './topologyDraft'
import type { CanvasInteractionMode } from './topologyTypes'

interface UseTopologyInteractionBindingsOptions {
  draft: Ref<TopologyEditorDraft>
  selectedNodeKey: Ref<string | null>
  selectedEdgeId: Ref<string | null>
  interactionMode: Ref<CanvasInteractionMode>
  pendingSourceNodeKey: Ref<string | null>
  nodePositions: Ref<Record<string, CanvasNodePosition>>
  removeSelectedCanvasItem: () => void | Promise<void>
  reloadAll: () => Promise<void>
}

export function useTopologyInteractionBindings(options: UseTopologyInteractionBindingsOptions) {
  const {
    draft,
    selectedNodeKey,
    selectedEdgeId,
    interactionMode,
    pendingSourceNodeKey,
    nodePositions,
    removeSelectedCanvasItem,
    reloadAll,
  } = options

  function isEditingTarget(target: EventTarget | null): boolean {
    if (!(target instanceof HTMLElement)) {
      return false
    }
    const tag = target.tagName
    return tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT' || target.isContentEditable
  }

  function handleGlobalKeydown(event: KeyboardEvent) {
    if (isEditingTarget(event.target)) {
      return
    }

    if (event.key === 'Escape') {
      pendingSourceNodeKey.value = null
      selectedEdgeId.value = null
      interactionMode.value = 'pan'
      return
    }

    if (
      (event.key === 'Delete' || event.key === 'Backspace') &&
      (selectedNodeKey.value || selectedEdgeId.value)
    ) {
      event.preventDefault()
      void removeSelectedCanvasItem()
    }
  }

  onMounted(async () => {
    window.addEventListener('keydown', handleGlobalKeydown)
    await reloadAll()
  })

  onBeforeUnmount(() => {
    window.removeEventListener('keydown', handleGlobalKeydown)
  })

  watch(
    () => draft.value.nodes.length,
    () => {
      nodePositions.value = normalizeCanvasPositions(draft.value, nodePositions.value)
      if (
        selectedNodeKey.value &&
        !draft.value.nodes.some((node) => node.key === selectedNodeKey.value)
      ) {
        selectedNodeKey.value = draft.value.nodes[0]?.key || null
      }
    }
  )

  watch(interactionMode, (value) => {
    if (value === 'pan' || value === 'add-node') {
      pendingSourceNodeKey.value = null
    }
  })
}
