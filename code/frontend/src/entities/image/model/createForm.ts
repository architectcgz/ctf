export interface ImageCreateForm {
  name: string
  tag: string
  description: string
}

export function createEmptyImageCreateForm(): ImageCreateForm {
  return {
    name: '',
    tag: '',
    description: '',
  }
}
