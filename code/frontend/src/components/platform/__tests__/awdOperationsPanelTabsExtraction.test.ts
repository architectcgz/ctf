import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import awdOperationsPanelSource from '@/features/contest-awd-admin/ui/AWDOperationsPanel.vue?raw'
import awdOperationsDialogHubSource from '@/features/contest-awd-admin/ui/AWDOperationsDialogHub.vue?raw'
import awdOperationsPreRuntimeStageSource from '@/features/contest-awd-admin/ui/AWDOperationsPreRuntimeStage.vue?raw'
import awdOperationsRuntimeStageSource from '@/features/contest-awd-admin/ui/AWDOperationsRuntimeStage.vue?raw'
import awdOperationsTabsSource from '@/features/contest-awd-admin/ui/AWDOperationsTabs.vue?raw'
import awdOperationsDialogAvailabilitySource from '@/features/contest-awd-admin/ui/useAwdOperationsDialogAvailability.ts?raw'
import awdOperationsDialogContractsSource from '@/features/contest-awd-admin/ui/awdOperationsDialogContracts.ts?raw'
import awdOperationsDialogStateSource from '@/features/contest-awd-admin/ui/useAwdOperationsDialogState.ts?raw'
import awdOperationsViewStateSource from '@/features/contest-awd-admin/ui/useAwdOperationsPanelViewState.ts?raw'

const awdOperationsAggregateSource = [
  awdOperationsPanelSource,
  awdOperationsViewStateSource,
  awdOperationsDialogStateSource,
  readFileSync(
    resolve(process.cwd(), 'src/features/contest-awd-admin/ui/awdOperationsPanel.css'),
    'utf8'
  ),
].join('\n')

describe('awd operations panel tabs extraction', () => {
  it('AWDOperationsPanel 应复用 useTabKeyboardNavigation，而不是继续本地维护按钮 ref 与键盘导航状态机', () => {
    expect(awdOperationsViewStateSource).toContain(
      "import { useTabKeyboardNavigation } from '@/composables/useTabKeyboardNavigation'"
    )
    expect(awdOperationsViewStateSource).toContain('useTabKeyboardNavigation<AWDOperationsPanelKey>({')
    expect(awdOperationsPanelSource).toContain("import { useAwdOperationsPanelViewState } from './useAwdOperationsPanelViewState'")
    expect(awdOperationsPanelSource).not.toContain(
      'const tabButtonRefs = ref<Array<HTMLButtonElement | null>>([])'
    )
    expect(awdOperationsPanelSource).not.toContain(
      'function setTabButtonRef(index: number, element: HTMLButtonElement | null) {'
    )
    expect(awdOperationsPanelSource).not.toContain('function focusTabByIndex(index: number) {')
    expect(awdOperationsPanelSource).not.toContain(
      'function handlePanelKeydown(event: KeyboardEvent, index: number) {'
    )
    expect(awdOperationsPanelSource).not.toContain('const activePanel = ref<AWDOperationsPanelKey>(')
    expect(awdOperationsPreRuntimeStageSource).toContain('<AWDOperationsTabs')
    expect(awdOperationsRuntimeStageSource).toContain('<AWDOperationsTabs')
    expect(awdOperationsTabsSource).toContain('class="studio-ops-tabs"')
  })

  it('AWDOperationsPanel 应将赛事选择器与未开赛运行壳层下沉到独立子组件，而不是继续在父组件内联整段结构', () => {
    expect(awdOperationsPanelSource).toContain('<AWDContestSelectorField')
    expect(awdOperationsPanelSource).toContain('<AWDOperationsPreRuntimeStage')
    expect(awdOperationsPanelSource).toContain('<AWDOperationsRuntimeStage')
    expect(awdOperationsPanelSource).toContain('<AWDOperationsDialogHub')
    expect(awdOperationsPanelSource).toContain(
      "import { useAwdOperationsDialogAvailability } from './useAwdOperationsDialogAvailability'"
    )
    expect(awdOperationsPanelSource).toContain("import { useAwdOperationsDialogState } from './useAwdOperationsDialogState'")
    expect(awdOperationsPanelSource).toContain('} = useAwdOperationsDialogAvailability({')
    expect(awdOperationsDialogStateSource).toContain('const roundDialogOpen = ref(false)')
    expect(awdOperationsDialogStateSource).toContain('const serviceCheckDialogOpen = ref(false)')
    expect(awdOperationsDialogStateSource).toContain('const attackLogDialogOpen = ref(false)')
    expect(awdOperationsDialogStateSource).toContain('runDialogMutationAndClose(')
    expect(awdOperationsDialogStateSource).not.toContain('const canRecordServiceChecks = computed(')
    expect(awdOperationsDialogStateSource).not.toContain('const serviceCheckHint = computed(() => {')
    expect(awdOperationsDialogAvailabilitySource).toContain('const canRecordServiceChecks = computed(')
    expect(awdOperationsDialogAvailabilitySource).toContain('const attackLogHint = computed(() => {')
    expect(awdOperationsDialogStateSource).toContain("from './awdOperationsDialogContracts'")
    expect(awdOperationsDialogHubSource).toContain("from './awdOperationsDialogContracts'")
    expect(awdOperationsDialogContractsSource).toContain('export interface AwdCreateRoundPayload')
    expect(awdOperationsDialogContractsSource).toContain('export interface AwdCreateServiceCheckPayload')
    expect(awdOperationsDialogContractsSource).toContain('export interface AwdCreateAttackLogPayload')
    expect(awdOperationsAggregateSource).toContain('.studio-ops-shell')
    expect(awdOperationsPanelSource).not.toContain('<style scoped>')
    expect(awdOperationsPreRuntimeStageSource).toContain('name="pending"')
    expect(awdOperationsRuntimeStageSource).toContain('name="inspector"')
    expect(awdOperationsDialogHubSource).toContain('<AWDRoundCreateDialog')
    expect(awdOperationsDialogHubSource).toContain('<AWDReadinessOverrideDialog')
    expect(awdOperationsPanelSource).not.toContain('id="awd-runtime-shell-create-round"')
    expect(awdOperationsPanelSource).not.toContain('id="awd-runtime-shell-run-check"')
    expect(awdOperationsPanelSource).not.toContain('id="awd-contest-selector"')
    expect(awdOperationsPanelSource).not.toContain('class="section-header"')
    expect(awdOperationsPanelSource).not.toContain('class="studio-ops-tabs"')
    expect(awdOperationsPanelSource).not.toContain('<AWDRoundCreateDialog')
    expect(awdOperationsPanelSource).not.toContain('<AWDServiceCheckDialog')
    expect(awdOperationsPanelSource).not.toContain('<AWDAttackLogDialog')
    expect(awdOperationsPanelSource).not.toContain('<AWDReadinessOverrideDialog')
    expect(awdOperationsPanelSource).not.toContain("const serviceCheckHint = computed(() => {")
    expect(awdOperationsPanelSource).not.toContain("const attackLogHint = computed(() => {")
  })
})
