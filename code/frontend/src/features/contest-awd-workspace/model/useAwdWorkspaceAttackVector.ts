import { computed, ref, toValue, watch, type MaybeRefOrGetter } from 'vue'

import type {
  ContestAWDWorkspaceTargetServiceData,
  ContestAWDWorkspaceTargetTeamData,
  ContestChallengeItem,
} from '@/api/contracts'
import { isAwdRuntimeChallenge, type AWDRuntimeChallenge } from './awdChallengeIdentity'

type AttackTargetItem = ContestAWDWorkspaceTargetTeamData & {
  active_service?: ContestAWDWorkspaceTargetServiceData
}

interface UseAwdWorkspaceAttackVectorOptions {
  challenges: MaybeRefOrGetter<ContestChallengeItem[]>
  targets: MaybeRefOrGetter<ContestAWDWorkspaceTargetTeamData[]>
  submitAttack: (serviceId: string, victimTeamId: number, flag: string) => Promise<unknown | null>
}

export function useAwdWorkspaceAttackVector(options: UseAwdWorkspaceAttackVectorOptions) {
  const activeChallengeKey = ref('')
  const flagInputs = ref<Record<string, string>>({})
  const targetKeyword = ref('')

  const runtimeChallenges = computed(() =>
    toValue(options.challenges).filter(isAwdRuntimeChallenge)
  )

  const attackToolbarChallengeOptions = computed(() =>
    runtimeChallenges.value.map((challenge) => ({
      key: getChallengeRuntimeKey(challenge),
      title: challenge.title,
    }))
  )

  const targetFilterKeyword = computed(() => targetKeyword.value.trim().toLowerCase())

  const activeChallenge = computed(
    () =>
      runtimeChallenges.value.find(
        (item) => getChallengeRuntimeKey(item) === activeChallengeKey.value
      ) || null
  )

  const activeChallengeRuntimeKey = computed(() => getChallengeRuntimeKey(activeChallenge.value))

  const filteredTargets = computed<AttackTargetItem[]>(() => {
    const challenge = activeChallenge.value
    if (!challenge) {
      return []
    }

    return toValue(options.targets)
      .map((target) => ({
        ...target,
        active_service: target.services.find((service) =>
          isTargetServiceForChallenge(service, challenge)
        ),
      }))
      .filter((target) => {
        if (targetFilterKeyword.value.length === 0) {
          return true
        }
        return target.team_name.toLowerCase().includes(targetFilterKeyword.value)
      })
  })

  watch(
    () => runtimeChallenges.value.map((item) => getChallengeRuntimeKey(item)),
    (challengeKeys) => {
      if (challengeKeys.length === 0) {
        activeChallengeKey.value = ''
        return
      }
      if (!challengeKeys.includes(activeChallengeKey.value)) {
        activeChallengeKey.value = challengeKeys[0]
      }
    },
    { immediate: true }
  )

  async function handleSubmit(serviceKey: string, teamId: string): Promise<void> {
    const stateKey = buildAttackStateKey(serviceKey, teamId)
    const flag = flagInputs.value[stateKey] || ''
    const result = await options.submitAttack(serviceKey, Number(teamId), flag)
    if (result) {
      flagInputs.value[stateKey] = ''
    }
  }

  return {
    activeChallengeKey,
    flagInputs,
    targetKeyword,
    runtimeChallenges,
    attackToolbarChallengeOptions,
    activeChallenge,
    activeChallengeRuntimeKey,
    filteredTargets,
    handleSubmit,
  }
}

function getChallengeRuntimeKey(challenge: ContestChallengeItem | null | undefined): string {
  return challenge?.awd_service_id || ''
}

function isTargetServiceForChallenge(
  service: { service_id?: string; awd_challenge_id: string },
  challenge: AWDRuntimeChallenge
): boolean {
  return service.service_id === challenge.awd_service_id
}

function buildAttackStateKey(serviceKey: string, teamId: string): string {
  return `${serviceKey}:${teamId}`
}
