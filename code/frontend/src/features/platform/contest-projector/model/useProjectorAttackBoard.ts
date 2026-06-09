import {
  nextTick,
  onMounted,
  onUnmounted,
  ref,
  watch,
  type ComponentPublicInstance,
  type Ref,
} from 'vue'

import type { AWDTeamServiceData } from '@/api/contracts'
import type {
  ContestProjectorAttackEdge,
  ContestProjectorAttackTeamPanel,
} from './projectorTypes'
import { getProjectorAttackServiceKey } from './projectorAttackMapSupport'

interface AttackBeam {
  id: string
  edge: ContestProjectorAttackEdge
  path: string
  markerX: number
  markerY: number
}

interface TeamDragOffset {
  x: number
  y: number
}

interface TeamDragState {
  teamId: string
  pointerId: number
  startX: number
  startY: number
  originX: number
  originY: number
  minX: number
  maxX: number
  minY: number
  maxY: number
}

interface UseProjectorAttackBoardOptions {
  teamPanels: Ref<ContestProjectorAttackTeamPanel[]>
  visibleEdges: Ref<ContestProjectorAttackEdge[]>
  expanded: Ref<boolean>
  boardOnly: Ref<boolean>
  firstBloodTargetKey: Ref<string | null>
}

const boardRingXRadius = 42
const boardRingYRadius = 36

export function useProjectorAttackBoard({
  teamPanels,
  visibleEdges,
  expanded,
  boardOnly,
  firstBloodTargetKey,
}: UseProjectorAttackBoardOptions) {
  const boardRef = ref<HTMLElement | null>(null)
  const teamDragOffsets = ref<Record<string, TeamDragOffset>>({})
  const teamDragState = ref<TeamDragState | null>(null)
  const beams = ref<AttackBeam[]>([])
  const teamRefs = new Map<string, HTMLElement>()
  const serviceRefs = new Map<string, HTMLElement>()
  let resizeObserver: ResizeObserver | null = null

  function clampValue(value: number, min: number, max: number): number {
    return Math.min(Math.max(value, min), max)
  }

  function getBoardNodePosition(index: number, total: number): { x: number; y: number } {
    if (total <= 1) {
      return { x: 50, y: 50 }
    }

    const angle = -Math.PI / 2 + (Math.PI * 2 * index) / total
    return {
      x: 50 + Math.cos(angle) * boardRingXRadius,
      y: 50 + Math.sin(angle) * boardRingYRadius,
    }
  }

  const dragStorageKey = ref('')

  watch(
    teamPanels,
    (panels) => {
      dragStorageKey.value = `contest-projector:attack-board:drag-offsets:${panels
        .map((panel) => panel.row.team_id)
        .sort()
        .join('|')}`
    },
    { immediate: true }
  )

  function loadTeamDragOffsets(): void {
    if (typeof window === 'undefined') {
      return
    }
    try {
      const rawValue = window.localStorage.getItem(dragStorageKey.value)
      if (!rawValue) {
        teamDragOffsets.value = {}
        return
      }
      const parsed = JSON.parse(rawValue) as Record<string, TeamDragOffset>
      teamDragOffsets.value = Object.fromEntries(
        Object.entries(parsed).filter(([teamId]) =>
          teamPanels.value.some((panel: ContestProjectorAttackTeamPanel) => panel.row.team_id === teamId)
        )
      )
    } catch {
      teamDragOffsets.value = {}
    }
  }

  function saveTeamDragOffsets(): void {
    if (typeof window === 'undefined') {
      return
    }
    try {
      window.localStorage.setItem(dragStorageKey.value, JSON.stringify(teamDragOffsets.value))
    } catch {
      // Ignore storage failures; dragging should still work for the current session.
    }
  }

  function setTeamRef(teamId: string, element: Element | ComponentPublicInstance | null): void {
    if (element instanceof HTMLElement) {
      teamRefs.set(teamId, element)
      resizeObserver?.observe(element)
      return
    }
    teamRefs.delete(teamId)
  }

  function setServiceRef(key: string, element: Element | ComponentPublicInstance | null): void {
    if (element instanceof HTMLElement) {
      serviceRefs.set(key, element)
      resizeObserver?.observe(element)
      return
    }
    serviceRefs.delete(key)
  }

  function updateBeams(): void {
    const board = boardRef.value
    if (!board) {
      beams.value = []
      return
    }

    const boardRect = board.getBoundingClientRect()
    beams.value = visibleEdges.value
      .map((edge: ContestProjectorAttackEdge) => {
        const source = teamRefs.get(edge.attacker_team_id)
        const target = serviceRefs.get(edge.latest_target_key)
        if (!source || !target) {
          return null
        }

        const sourceRect = source.getBoundingClientRect()
        const targetRect = target.getBoundingClientRect()
        const sourceCenterX = sourceRect.left + sourceRect.width / 2 - boardRect.left
        const targetCenterX = targetRect.left + targetRect.width / 2 - boardRect.left
        const sourceX =
          sourceCenterX <= targetCenterX
            ? sourceRect.right - boardRect.left
            : sourceRect.left - boardRect.left
        const sourceY = sourceRect.top + sourceRect.height / 2 - boardRect.top
        const targetX =
          sourceCenterX <= targetCenterX
            ? targetRect.left - boardRect.left
            : targetRect.right - boardRect.left
        const targetY = targetRect.top + targetRect.height / 2 - boardRect.top
        const distanceX = Math.abs(targetX - sourceX)
        const curve = Math.max(72, distanceX * 0.42)
        const controlAX = sourceX + (targetX >= sourceX ? curve : -curve)
        const controlBX = targetX - (targetX >= sourceX ? curve : -curve)

        return {
          id: edge.id,
          edge,
          path: `M ${sourceX} ${sourceY} C ${controlAX} ${sourceY}, ${controlBX} ${targetY}, ${targetX} ${targetY}`,
          markerX: targetX,
          markerY: targetY,
        }
      })
      .filter((item: AttackBeam | null): item is AttackBeam => item !== null)
  }

  async function scheduleBeamUpdate(): Promise<void> {
    await nextTick()
    updateBeams()
  }

  function getTeamNodeStyle(teamId: string, index: number): Record<string, string> | undefined {
    if (!expanded.value) {
      return undefined
    }

    const offset = teamDragOffsets.value[teamId]
    if (boardOnly.value) {
      const position = getBoardNodePosition(index, teamPanels.value.length)
      const dragTransform = offset ? ` translate3d(${offset.x}px, ${offset.y}px, 0)` : ''
      return {
        left: `${position.x}%`,
        top: `${position.y}%`,
        transform: `translate(-50%, -50%)${dragTransform}`,
      }
    }

    if (!offset) {
      return undefined
    }
    return {
      transform: `translate3d(${offset.x}px, ${offset.y}px, 0)`,
    }
  }

  function isDraggingTeam(teamId: string): boolean {
    return teamDragState.value?.teamId === teamId
  }

  function startTeamDrag(event: PointerEvent, teamId: string): void {
    if (!expanded.value) {
      return
    }

    const board = boardRef.value
    const target = event.currentTarget
    if (!board || !(target instanceof HTMLElement)) {
      return
    }

    event.preventDefault()
    event.stopPropagation()

    const boardRect = board.getBoundingClientRect()
    const targetRect = target.getBoundingClientRect()
    const currentOffset = teamDragOffsets.value[teamId] ?? { x: 0, y: 0 }
    const safeInset = 8

    teamDragState.value = {
      teamId,
      pointerId: event.pointerId,
      startX: event.clientX,
      startY: event.clientY,
      originX: currentOffset.x,
      originY: currentOffset.y,
      minX: safeInset - (targetRect.left - currentOffset.x - boardRect.left),
      maxX:
        boardRect.width -
        safeInset -
        targetRect.width -
        (targetRect.left - currentOffset.x - boardRect.left),
      minY: safeInset - (targetRect.top - currentOffset.y - boardRect.top),
      maxY:
        boardRect.height -
        safeInset -
        targetRect.height -
        (targetRect.top - currentOffset.y - boardRect.top),
    }

    target.setPointerCapture(event.pointerId)
  }

  function moveTeamDrag(event: PointerEvent): void {
    const state = teamDragState.value
    if (!state || state.pointerId !== event.pointerId) {
      return
    }

    event.preventDefault()
    const nextOffset = {
      x: clampValue(state.originX + event.clientX - state.startX, state.minX, state.maxX),
      y: clampValue(state.originY + event.clientY - state.startY, state.minY, state.maxY),
    }
    teamDragOffsets.value = {
      ...teamDragOffsets.value,
      [state.teamId]: nextOffset,
    }
    requestAnimationFrame(updateBeams)
  }

  function endTeamDrag(event: PointerEvent): void {
    const state = teamDragState.value
    if (!state || state.pointerId !== event.pointerId) {
      return
    }

    const target = event.currentTarget
    if (target instanceof HTMLElement && target.hasPointerCapture(event.pointerId)) {
      target.releasePointerCapture(event.pointerId)
    }
    teamDragState.value = null
    saveTeamDragOffsets()
    requestAnimationFrame(updateBeams)
  }

  function resetTeamDrag(teamId: string): void {
    if (!expanded.value) {
      return
    }
    const { [teamId]: _removed, ...restOffsets } = teamDragOffsets.value
    teamDragOffsets.value = restOffsets
    saveTeamDragOffsets()
    void scheduleBeamUpdate()
  }

  function getServiceAttackCount(teamId: string, service: AWDTeamServiceData): number {
    const serviceKey = getProjectorAttackServiceKey(teamId, service)
    return visibleEdges.value
      .filter((edge: ContestProjectorAttackEdge) => edge.latest_target_key === serviceKey && edge.success > 0)
      .reduce((total: number, edge: ContestProjectorAttackEdge) => total + edge.success, 0)
  }

  function isFirstBloodTarget(teamId: string, service: AWDTeamServiceData): boolean {
    return firstBloodTargetKey.value === getProjectorAttackServiceKey(teamId, service)
  }

  watch(
    () => [teamPanels.value, visibleEdges.value, expanded.value],
    () => {
      void scheduleBeamUpdate()
    },
    { deep: true }
  )

  watch(
    () => [dragStorageKey.value, expanded.value],
    () => {
      if (expanded.value) {
        loadTeamDragOffsets()
        void scheduleBeamUpdate()
      }
    }
  )

  onMounted(() => {
    if (typeof ResizeObserver !== 'undefined') {
      resizeObserver = new ResizeObserver(() => updateBeams())
      if (boardRef.value) {
        resizeObserver.observe(boardRef.value)
      }
    }
    if (expanded.value) {
      loadTeamDragOffsets()
    }
    void scheduleBeamUpdate()
  })

  onUnmounted(() => {
    resizeObserver?.disconnect()
    resizeObserver = null
    teamRefs.clear()
    serviceRefs.clear()
  })

  return {
    boardRef,
    beams,
    setTeamRef,
    setServiceRef,
    getTeamNodeStyle,
    isDraggingTeam,
    startTeamDrag,
    moveTeamDrag,
    endTeamDrag,
    resetTeamDrag,
    getServiceAttackCount,
    isFirstBloodTarget,
  }
}
