import { describe, it, expect, beforeEach, vi } from 'vitest'
import { useTheme } from '../useTheme'

describe('useTheme', () => {
  beforeEach(() => {
    localStorage.clear()
    document.documentElement.removeAttribute('data-theme')
    document.documentElement.removeAttribute('data-brand')
    vi.clearAllMocks()

    // Mock window.matchMedia
    Object.defineProperty(window, 'matchMedia', {
      writable: true,
      value: vi.fn().mockImplementation(query => ({
        matches: false,
        media: query,
        onchange: null,
        addListener: vi.fn(),
        removeListener: vi.fn(),
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
        dispatchEvent: vi.fn(),
      })),
    })
  })

  it('应该初始化为 dark 模式', () => {
    const { theme } = useTheme()
    expect(theme.value).toBe('dark')
  })

  it('应该切换主题', () => {
    const { theme, toggleTheme } = useTheme()

    expect(theme.value).toBe('dark')
    toggleTheme()
    expect(theme.value).toBe('light')
    toggleTheme()
    expect(theme.value).toBe('dark')
  })

  it('应该设置 data-theme 属性', () => {
    const { toggleTheme } = useTheme()

    toggleTheme()
    expect(document.documentElement.getAttribute('data-theme')).toBe('light')

    toggleTheme()
    expect(document.documentElement.getAttribute('data-theme')).toBe('dark')
  })

  it('应该持久化主题到 localStorage', () => {
    const { toggleTheme } = useTheme()

    toggleTheme()
    expect(localStorage.getItem('theme')).toBe('light')

    toggleTheme()
    expect(localStorage.getItem('theme')).toBe('dark')
  })

  it('应该从 localStorage 恢复主题', () => {
    localStorage.setItem('theme', 'dark')

    const { theme, initTheme } = useTheme()
    initTheme()

    expect(theme.value).toBe('dark')
    expect(document.documentElement.getAttribute('data-theme')).toBe('dark')
  })

  it('应该设置品牌主题并写入 data-brand', () => {
    const { brand, setBrand } = useTheme()

    expect(brand.value).toBe('green')
    setBrand('green')

    expect(brand.value).toBe('green')
    expect(document.documentElement.getAttribute('data-brand')).toBe('green')
  })

  it('应该持久化品牌主题到 localStorage', () => {
    const { setBrand } = useTheme()

    setBrand('green')

    expect(localStorage.getItem('theme-brand')).toBe('green')
  })

  it('应该从 localStorage 恢复品牌主题', () => {
    localStorage.setItem('theme-brand', 'green')

    const { brand, initTheme } = useTheme()
    initTheme()

    expect(brand.value).toBe('green')
    expect(document.documentElement.getAttribute('data-brand')).toBe('green')
  })
})
