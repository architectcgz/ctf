import type { AWDDefenseSSHAccessData } from '@/api/contracts'

export function getVSCodeSSHCommand(access?: AWDDefenseSSHAccessData): string {
  return access?.command || ''
}
