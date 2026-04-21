import { mount } from '@vue/test-utils'
import { defineComponent, h } from 'vue'
import { describe, expect, it } from 'vitest'

import AWDServiceTemplateEditorDialog from '../AWDServiceTemplateEditorDialog.vue'

const AdminSurfaceModalStub = defineComponent({
  props: {
    open: { type: Boolean, default: false },
    title: { type: String, default: '' },
    subtitle: { type: String, default: '' },
    eyebrow: { type: String, default: '' },
  },
  emits: ['close', 'update:open'],
  setup(props, { slots }) {
    return () =>
      h('section', { 'data-test': 'admin-surface-modal', 'data-open': String(props.open) }, [
        h('header', [props.title, props.subtitle, props.eyebrow]),
        slots.default?.(),
        slots.footer?.(),
      ])
  },
})

describe('AWDServiceTemplateEditorDialog', () => {
  it('uses category enum select instead of free text input', async () => {
    const wrapper = mount(AWDServiceTemplateEditorDialog, {
      props: {
        open: true,
        mode: 'create',
        saving: false,
        draft: {
          name: '',
          slug: '',
          category: 'web',
          difficulty: 'medium',
          description: '',
          service_type: 'web_http',
          deployment_mode: 'single_container',
          status: 'draft',
        },
      },
      global: {
        stubs: {
          AdminSurfaceModal: AdminSurfaceModalStub,
        },
      },
    })

    const categoryField = wrapper.get('#awd-template-category')
    expect(categoryField.element.tagName).toBe('SELECT')

    const optionValues = categoryField
      .findAll('option')
      .map((option) => (option.element as HTMLOptionElement).value)
    expect(optionValues).toEqual(['web', 'pwn', 'reverse', 'crypto', 'misc', 'forensics'])

    await wrapper.get('#awd-template-name').setValue('Crypto Bank AWD')
    await wrapper.get('#awd-template-slug').setValue('crypto-bank-awd')
    await categoryField.setValue('crypto')
    await wrapper.get('#awd-template-dialog-submit').trigger('click')

    expect(wrapper.emitted('save')).toEqual([
      [
        expect.objectContaining({
          category: 'crypto',
        }),
      ],
    ])
  })
})
