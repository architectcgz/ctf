import { afterEach, describe, expect, it } from 'vitest'
import { nextTick } from 'vue'

import { confirmDestructiveAction, useDestructiveConfirmState } from './useDestructiveConfirm'

describe('useDestructiveConfirm', () => {
  afterEach(() => {
    const { cancel } = useDestructiveConfirmState()
    cancel()
    document.body.innerHTML = ''
  })

  it('在确认后应 resolve true 并清空当前确认状态', async () => {
    const { current, visible, confirm } = useDestructiveConfirmState()

    const promise = confirmDestructiveAction({
      title: '删除用户',
      message: '确定要删除 alice 吗？',
      confirmButtonText: '确认删除',
    })

    expect(visible.value).toBe(true)
    expect(current.value?.title).toBe('删除用户')
    expect(current.value?.message).toBe('确定要删除 alice 吗？')
    expect(current.value?.warning).toBe('此操作不可恢复，请确认后继续。')

    confirm()

    await expect(promise).resolves.toBe(true)
    expect(visible.value).toBe(false)
    expect(current.value).toBeNull()
  })

  it('在第二次打开确认框时应先收口前一个 promise，避免悬空等待', async () => {
    const { current, cancel } = useDestructiveConfirmState()

    const first = confirmDestructiveAction({
      title: '删除题目',
      message: '确定要删除这道题目吗？',
    })

    const second = confirmDestructiveAction({
      title: '结束比赛',
      message: '确认结束当前比赛吗？',
      confirmButtonText: '确认结束',
    })

    await expect(first).resolves.toBe(false)
    expect(current.value?.title).toBe('结束比赛')
    expect(current.value?.warning).toBe('确认后将立即生效，请确认后继续。')

    cancel()

    await expect(second).resolves.toBe(false)
    expect(current.value).toBeNull()
  })

  it('关闭后应把焦点还给触发按钮', async () => {
    const trigger = document.createElement('button')
    document.body.appendChild(trigger)
    trigger.focus()

    const { cancel } = useDestructiveConfirmState()

    const promise = confirmDestructiveAction({
      title: '销毁实例',
      message: '确定要销毁当前实例吗？',
      confirmButtonText: '确认销毁',
    })

    cancel()

    await expect(promise).resolves.toBe(false)
    await nextTick()
    expect(document.activeElement).toBe(trigger)
  })
})
