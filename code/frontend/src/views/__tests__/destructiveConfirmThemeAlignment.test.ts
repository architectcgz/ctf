import { readFileSync } from 'node:fs'

import { describe, expect, it } from 'vitest'

const deleteConfirmModalSource = readFileSync(
  `${process.cwd()}/src/components/common/DeleteConfirmModal.vue`,
  'utf-8'
)
const destructiveConfirmHostSource = readFileSync(
  `${process.cwd()}/src/components/common/AppDestructiveConfirm.vue`,
  'utf-8'
)

describe('destructive confirm theme alignment', () => {
  it('危险确认框应复用本地 modal shell 和主题 token，而不是继续依赖 ElMessageBox', () => {
    expect(deleteConfirmModalSource).toContain(
      "from '@/components/common/modal-templates/ModalTemplateShell.vue'"
    )
    expect(deleteConfirmModalSource).toContain('var(--color-danger)')
    expect(deleteConfirmModalSource).toContain('color-mix(')
    expect(deleteConfirmModalSource).toContain('frosted')
    expect(deleteConfirmModalSource).toContain('--delete-confirm-modal-outer-border')
    expect(deleteConfirmModalSource).toContain(
      '0 0 0 var(--space-0-5) var(--delete-confirm-modal-outer-border)'
    )
    expect(deleteConfirmModalSource).toContain('.delete-confirm-modal {')
    expect(deleteConfirmModalSource).toContain(
      'color: color-mix(in srgb, var(--color-danger) 72%, var(--color-text-primary))'
    )
    expect(deleteConfirmModalSource).not.toContain('ElMessageBox')
    expect(deleteConfirmModalSource).not.toContain('element-plus')
    expect(deleteConfirmModalSource).not.toContain('#dc2626')
    expect(destructiveConfirmHostSource).toContain('<DeleteConfirmModal')
  })
})
