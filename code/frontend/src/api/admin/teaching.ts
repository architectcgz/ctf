export { getClasses, getStudentsDirectory } from '../teaching/classes'
export type { TeacherStudentDirectoryParams } from '../teaching/classes'

export {
  getTeacherInstances as getPlatformInstances,
  destroyTeacherInstance as destroyPlatformInstance,
} from '../teaching/instances'
export type { TeacherInstanceStatusFilter as PlatformInstanceStatusFilter } from '../teaching/instances'
