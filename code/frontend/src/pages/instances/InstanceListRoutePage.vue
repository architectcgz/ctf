<template>
  <InstanceListWorkspaceShell
    :loading="loading"
    :max-instances="maxInstances"
    :instances="instances"
    :running-count="runningCount"
    :waiting-count="waitingCount"
    :show-warning="showWarning"
    :warning-instance="warningInstance"
    :warning-threshold-seconds="WARNING_THRESHOLD_SECONDS"
    :extend-duration-seconds="EXTEND_DURATION_SECONDS"
    :copy-address="copyAddress"
    :extend-time="extendTime"
    :open-target="openTarget"
    :destroy-instance="destroyInstance"
    :extend-from-warning="extendFromWarning"
    :close-warning="closeWarning"
    :set-warning-close-button="setWarningCloseButton"
    :can-open-instance-in-browser="canOpenInstanceInBrowser"
    :format-instance-access-display="formatInstanceAccessDisplay"
    :format-remaining-time="formatInstanceRemainingTime"
    :get-instance-status-class="getInstanceStatusDotClass"
    :get-instance-provisioning-label="getInstanceProvisioningLabel"
    :get-instance-waiting-hint="getInstanceWaitingHint"
    :is-instance-manual-action-allowed="isInstanceManualActionAllowed"
  />
</template>

<script setup lang="ts">
import { ref, type ComponentPublicInstance } from 'vue'

import {
  formatInstanceAccessDisplay,
  formatInstanceRemainingTime,
  getInstanceProvisioningLabel,
  getInstanceStatusDotClass,
  getInstanceWaitingHint,
} from '@/entities/instance'
import {
  EXTEND_DURATION_SECONDS,
  WARNING_THRESHOLD_SECONDS,
  canOpenInstanceInBrowser,
  isInstanceManualActionAllowed,
  InstanceListWorkspaceShell,
  useInstanceListPage,
  useInstanceWarningFocus,
} from '@/features/instance-list'

const {
  loading,
  maxInstances,
  instances,
  runningCount,
  waitingCount,
  showWarning,
  warningInstance,
  copyAddress,
  extendTime,
  openTarget,
  destroyInstance,
  extendFromWarning,
  closeWarning,
} = useInstanceListPage()

const warningCloseButton = ref<HTMLButtonElement | null>(null)
useInstanceWarningFocus({ showWarning, warningCloseButton })

function setWarningCloseButton(refTarget: Element | ComponentPublicInstance | null) {
  warningCloseButton.value = refTarget instanceof HTMLButtonElement ? refTarget : null
}
</script>
