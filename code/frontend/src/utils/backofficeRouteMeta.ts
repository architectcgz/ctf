import {
  getBackofficeActiveSecondaryRouteName,
  isBackofficePath,
} from '@/config/backofficeNavigation'

export function isBackofficeRoute(path: string): boolean {
  return isBackofficePath(path)
}

export function getBackofficeLayoutMode(path: string): 'backoffice' | 'default' {
  return isBackofficeRoute(path) ? 'backoffice' : 'default'
}

export { getBackofficeActiveSecondaryRouteName }
