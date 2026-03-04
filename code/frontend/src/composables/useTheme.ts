import { ref } from 'vue'

type Theme = 'light' | 'dark'

const theme = ref<Theme>('dark')
const initialized = ref(false)

export function useTheme() {
  const setTheme = (newTheme: Theme) => {
    theme.value = newTheme
    document.documentElement.setAttribute('data-theme', newTheme)
    localStorage.setItem('theme', newTheme)
  }

  const toggleTheme = () => {
    setTheme(theme.value === 'light' ? 'dark' : 'light')
  }

  const initTheme = () => {
    if (initialized.value) return
    initialized.value = true
    const saved = localStorage.getItem('theme') as Theme | null
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
    setTheme(saved || (prefersDark ? 'dark' : 'light'))
  }

  return { theme, toggleTheme, initTheme }
}
