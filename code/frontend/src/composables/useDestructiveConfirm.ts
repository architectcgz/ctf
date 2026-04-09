import { ElMessageBox } from 'element-plus'

export interface DestructiveConfirmOptions {
  message: string
  title?: string
  confirmButtonText?: string
  cancelButtonText?: string
}

export async function confirmDestructiveAction({
  message,
  title = '确认删除',
  confirmButtonText = '确认删除',
  cancelButtonText = '取消',
}: DestructiveConfirmOptions): Promise<boolean> {
  try {
    await ElMessageBox.confirm(message, title, {
      type: 'warning',
      showClose: true,
      closeOnClickModal: false,
      confirmButtonText,
      cancelButtonText,
      customClass: 'app-destructive-confirm-box',
      modalClass: 'app-destructive-confirm-modal',
      cancelButtonClass: 'app-destructive-confirm-cancel',
      confirmButtonClass: 'app-destructive-confirm-primary',
    })
    return true
  } catch (error) {
    if (error === 'cancel' || error === 'close') {
      return false
    }
    throw error
  }
}
