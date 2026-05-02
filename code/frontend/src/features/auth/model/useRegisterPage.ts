import { reactive, ref } from 'vue'

import { useAuth } from './useAuth'

interface RegisterFormState {
  username: string
  password: string
  class_name: string
}

export function useRegisterPage() {
  const { register } = useAuth()

  const loading = ref(false)
  const submitError = ref('')
  const form = reactive<RegisterFormState>({
    username: '',
    password: '',
    class_name: '',
  })

  function clearSubmitError() {
    submitError.value = ''
  }

  async function onSubmit() {
    if (loading.value || !form.username || !form.password) {
      return
    }

    loading.value = true
    submitError.value = ''
    try {
      await register({
        username: form.username,
        password: form.password,
        class_name: form.class_name.trim() || undefined,
      })
    } catch (err) {
      submitError.value = err instanceof Error && err.message.trim() ? err.message : '注册失败，请稍后重试'
    } finally {
      loading.value = false
    }
  }

  return {
    form,
    loading,
    submitError,
    clearSubmitError,
    onSubmit,
  }
}
