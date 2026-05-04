import type { AWDDefenseSSHAccessData, SSHProfileData } from '@/api/contracts'

export function buildOpenSSHConfig(profile?: SSHProfileData): string {
  if (!profile) return ''
  return [
    `Host ${profile.alias}`,
    `  HostName ${profile.host_name}`,
    `  Port ${profile.port}`,
    `  User ${profile.user}`,
    '',
  ].join('\n')
}

export function getVSCodeSSHCommand(access?: AWDDefenseSSHAccessData): string {
  return access?.command || ''
}
