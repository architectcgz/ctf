import { getCurrentInstance, onBeforeUnmount, reactive, ref } from 'vue'

import { useProbeEasterEggs } from '@/composables/useProbeEasterEggs'

import { useAuth } from './useAuth'
import { useLoginViewPage } from './useLoginViewPage'

interface LoginFormState {
  username: string
  password: string
}

interface SubmitFallbackValues {
  username?: string | null
  password?: string | null
}

function emitLoginConsoleHints() {
  // eslint-disable-next-line no-console
  console.log(
    '%c[CTF COMMAND CENTER] %cSystem online. Initializing monitoring...',
    'font-weight: bold; font-size: 14px;',
    'font-style: italic;'
  )
  // eslint-disable-next-line no-console
  console.log(
    `%c
      :::::::: ::::::::::: :::::::::: 
    :+:    :+:    :+:     :+:         
   +:+           +:+     +:+          
  +#+           +#+     +#++:++#      
 +#+           +#+     +#+            
#+#    #+#    #+#     #+#             
########     ###     ###              
`,
    'font-weight: bold;'
  )
  // eslint-disable-next-line no-console
  console.log(
    '%cWARNING: %cUnauthorized debugging may lead to "unexpected" results. Good luck, cadet.',
    'font-weight: bold;',
    ''
  )
  // eslint-disable-next-line no-console
  console.log(
    '%cAudit note: %ccuriosity detected. Keep it academic.',
    'font-weight: bold;',
    ''
  )
  // eslint-disable-next-line no-console
  console.log(
    '%cMemo: %cIf this page were the weak point, we would all be having a worse day.',
    'font-weight: bold;',
    ''
  )
}

export function useLoginPage() {
  const { login } = useAuth()
  const { redirectTo } = useLoginViewPage()
  const { track } = useProbeEasterEggs()

  const loading = ref(false)
  const submitError = ref('')
  const probeMessage = ref('')
  const form = reactive<LoginFormState>({
    username: '',
    password: '',
  })
  let probeMessageTimer: number | null = null

  emitLoginConsoleHints()

  function clearSubmitError() {
    submitError.value = ''
  }

  function showProbeMessage(message: string) {
    probeMessage.value = message
    if (probeMessageTimer) {
      window.clearTimeout(probeMessageTimer)
    }
    probeMessageTimer = window.setTimeout(() => {
      probeMessage.value = ''
      probeMessageTimer = null
    }, 3000)
  }

  function handleHeroProbe() {
    const result = track('login-brand', 4)
    if (!result.unlocked) {
      return
    }
    showProbeMessage('隐藏入口排查完毕，结果让你失望了。')
  }

  if (getCurrentInstance()) {
    onBeforeUnmount(() => {
      if (probeMessageTimer) {
        window.clearTimeout(probeMessageTimer)
      }
    })
  }

  async function onSubmit(fallbackValues?: SubmitFallbackValues) {
    const fallbackUsername = fallbackValues?.username?.trim() ?? ''
    const fallbackPassword = fallbackValues?.password ?? ''
    const username = form.username.trim() || fallbackUsername
    const password = form.password || fallbackPassword

    if (loading.value || !username || !password) {
      return
    }

    form.username = username
    form.password = password
    loading.value = true
    submitError.value = ''
    try {
      await login({ username, password }, redirectTo.value === '/' ? undefined : redirectTo.value)
    } catch (err) {
      submitError.value =
        err instanceof Error && err.message.trim() ? err.message : '身份验证失败，请核对信息'
    } finally {
      loading.value = false
    }
  }

  return {
    form,
    loading,
    probeMessage,
    submitError,
    clearSubmitError,
    handleHeroProbe,
    onSubmit,
  }
}
