import { ref } from 'vue'

export type Theme = 'light' | 'dark'
export type ThemeBrand = 'green' | 'cyan' | 'blue'

type ThemeBrandOption = {
  value: ThemeBrand
  label: string
  description: string
}

const THEME_STORAGE_KEY = 'theme'
const BRAND_STORAGE_KEY = 'theme-brand'
const DEFAULT_THEME: Theme = 'dark'
const DEFAULT_BRAND: ThemeBrand = 'green'

const theme = ref<Theme>(DEFAULT_THEME)
const brand = ref<ThemeBrand>(DEFAULT_BRAND)

const availableBrands: ThemeBrandOption[] = [
  { value: 'green', label: '绿色', description: '更贴近学校 CTF 的默认技术主题' },
  { value: 'cyan', label: '青色', description: '保留当前较冷静的青蓝技术感' },
  { value: 'blue', label: '蓝色', description: '更传统、更中性的控制台色调' },
]

function isTheme(value: string | null): value is Theme {
  return value === 'light' || value === 'dark'
}

function isThemeBrand(value: string | null): value is ThemeBrand {
  return value === 'green' || value === 'cyan' || value === 'blue'
}

function syncDocumentTheme(): void {
  document.documentElement.setAttribute('data-theme', theme.value)
  document.documentElement.setAttribute('data-brand', brand.value)
}

export function useTheme() {
  const setTheme = (newTheme: Theme) => {
    theme.value = newTheme
    syncDocumentTheme()
    localStorage.setItem(THEME_STORAGE_KEY, newTheme)
  }

  const setBrand = (newBrand: ThemeBrand) => {
    brand.value = newBrand
    syncDocumentTheme()
    localStorage.setItem(BRAND_STORAGE_KEY, newBrand)
  }

  const toggleTheme = () => {
    setTheme(theme.value === 'light' ? 'dark' : 'light')
  }

  const initTheme = () => {
    const savedTheme = localStorage.getItem(THEME_STORAGE_KEY)
    const savedBrand = localStorage.getItem(BRAND_STORAGE_KEY)

    theme.value = isTheme(savedTheme) ? savedTheme : DEFAULT_THEME
    brand.value = isThemeBrand(savedBrand) ? savedBrand : DEFAULT_BRAND

    syncDocumentTheme()
    localStorage.setItem(THEME_STORAGE_KEY, theme.value)
    localStorage.setItem(BRAND_STORAGE_KEY, brand.value)
  }

  return { theme, brand, availableBrands, setTheme, setBrand, toggleTheme, initTheme }
}
