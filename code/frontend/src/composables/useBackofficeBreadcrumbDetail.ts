import { readonly, ref } from 'vue'

const breadcrumbDetailTitle = ref<string | null>(null)

export function useBackofficeBreadcrumbDetail() {
  function setBreadcrumbDetailTitle(title?: string | null): void {
    const normalizedTitle = title?.trim()
    breadcrumbDetailTitle.value = normalizedTitle || null
  }

  return {
    breadcrumbDetailTitle: readonly(breadcrumbDetailTitle),
    setBreadcrumbDetailTitle,
  }
}
