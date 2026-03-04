import { describe, it, expect, beforeEach, vi } from 'vitest'
import { useTheme } from '../useTheme'

describe('useTheme', () => {
  beforeEach(() => {
    localStorage.clear()
    document.documentElement.removeAttribute('data-theme')
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
})
