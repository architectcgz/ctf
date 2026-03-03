import { useToast } from '@/composables/useToast'

export function useClipboard() {
  const toast = useToast()

  async function copy(text: string): Promise<void> {
    try {
      if (navigator.clipboard?.writeText) {
        await navigator.clipboard.writeText(text)
      } else {
        const el = document.createElement('textarea')
        el.value = text
        el.style.position = 'fixed'
        el.style.opacity = '0'
        document.body.appendChild(el)
        el.focus()
        el.select()
        document.execCommand('copy')
        document.body.removeChild(el)
      }
      toast.success('已复制到剪贴板')
    } catch (err) {
      console.error('Copy failed:', err)
      toast.error('复制失败')
    }
  }

  return { copy }
}

