import {
  destroyManagedInstance,
  exportClassReport,
  getInstanceDirectory,
} from '../teaching/instances'

type InstanceDirectoryQueryParams = Parameters<typeof getInstanceDirectory>[0]
type InstanceDirectoryRequestOptions = Parameters<typeof getInstanceDirectory>[1]

export async function getTeacherInstances(
  params?: InstanceDirectoryQueryParams,
  options?: InstanceDirectoryRequestOptions
) {
  return getInstanceDirectory(params, options)
}

export async function destroyTeacherInstance(id: string) {
  return destroyManagedInstance(id)
}

export { exportClassReport }
export type { InstanceDirectoryStatusFilter } from '../teaching/instances'
