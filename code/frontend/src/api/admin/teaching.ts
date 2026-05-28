import {
  destroyManagedInstance,
  getInstanceDirectory,
} from '../teaching/instances'

export { getClasses, getStudentsDirectory } from '../teaching/classes'
export type { StudentDirectoryParams } from '../teaching/classes'

type InstanceDirectoryQueryParams = Parameters<typeof getInstanceDirectory>[0]
type InstanceDirectoryRequestOptions = Parameters<typeof getInstanceDirectory>[1]

export async function getPlatformInstances(
  params?: InstanceDirectoryQueryParams,
  options?: InstanceDirectoryRequestOptions
) {
  return getInstanceDirectory(params, options)
}

export async function destroyPlatformInstance(id: string) {
  return destroyManagedInstance(id)
}

export type {
  InstanceDirectoryStatusFilter as PlatformInstanceStatusFilter,
} from '../teaching/instances'
