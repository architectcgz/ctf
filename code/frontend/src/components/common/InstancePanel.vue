<template>
  <ElCard>
    <template #header>
      <div class="flex items-center justify-between">
        <span class="font-semibold">我的实例</span>
        <ElButton size="small" @click="refresh">刷新</ElButton>
      </div>
    </template>

    <div v-if="instances.length === 0" class="text-center text-gray-500 py-8">
      暂无运行中的实例
    </div>

    <div v-else class="space-y-4">
      <ElCard v-for="instance in instances" :key="instance.id" shadow="never" class="border">
        <div class="space-y-3">
          <div class="flex items-start justify-between">
            <div>
              <h3 class="font-semibold text-lg">{{ instance.challenge_title }}</h3>
              <ElTag :type="getStatusColor(instance.status)" size="small" class="mt-1">
                {{ getStatusLabel(instance.status) }}
              </ElTag>
            </div>
            <div class="text-right">
              <div :class="getTimeColor(instance.expires_at)" class="text-lg font-semibold">
                {{ formatCountdown(instance.expires_at) }}
              </div>
              <div class="text-xs text-gray-500">剩余时间</div>
            </div>
          </div>

          <div v-if="instance.access_url" class="space-y-2">
            <div class="text-sm">
              <span class="text-gray-600">访问地址：</span>
              <a :href="instance.access_url" target="_blank" class="text-primary hover:underline">
                {{ instance.access_url }}
              </a>
            </div>
          </div>

          <div class="flex gap-2">
            <ElButton
              v-if="instance.remaining_extends > 0"
              size="small"
              @click="extend(instance.id)"
            >
              延时 (剩余 {{ instance.remaining_extends }} 次)
            </ElButton>
            <ElButton
              size="small"
              type="danger"
              @click="destroy(instance.id)"
            >
              销毁
            </ElButton>
          </div>
        </div>
      </ElCard>
    </div>
  </ElCard>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { ElMessageBox } from 'element-plus'

import { getMyInstances, destroyInstance, extendInstance } from '@/api/instance'
import { useToast } from '@/composables/useToast'
import type { InstanceListItem, InstanceStatus } from '@/api/contracts'

const toast = useToast()
const instances = ref<InstanceListItem[]>([])
let timer: number | null = null

async function refresh() {
  try {
    instances.value = await getMyInstances()
    checkExpiringSoon()
  } catch (error) {
    toast.error('加载实例列表失败')
  }
}

async function destroy(id: string) {
  try {
    await ElMessageBox.confirm('确定要销毁此实例吗？', '确认', { type: 'warning' })
    await destroyInstance(id)
    toast.success('实例已销毁')
    refresh()
  } catch (error) {
    if (error !== 'cancel') {
      toast.error('销毁实例失败')
    }
  }
}

async function extend(id: string) {
  try {
    await extendInstance(id)
    toast.success('延时成功')
    refresh()
  } catch (error) {
    toast.error('延时失败')
  }
}

function formatCountdown(expiresAt: string): string {
  const now = Date.now()
  const expires = new Date(expiresAt).getTime()
  const diff = expires - now

  if (diff <= 0) return '已过期'

  const hours = Math.floor(diff / 3600000)
  const minutes = Math.floor((diff % 3600000) / 60000)
  const seconds = Math.floor((diff % 60000) / 1000)

  if (hours > 0) return `${hours}:${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`
  return `${minutes}:${String(seconds).padStart(2, '0')}`
}

function getTimeColor(expiresAt: string): string {
  const now = Date.now()
  const expires = new Date(expiresAt).getTime()
  const diff = expires - now

  if (diff <= 0) return 'text-gray-500'
  if (diff < 300000) return 'text-red-600'
  if (diff < 600000) return 'text-orange-600'
  return 'text-green-600'
}

function checkExpiringSoon() {
  instances.value.forEach(instance => {
    const now = Date.now()
    const expires = new Date(instance.expires_at).getTime()
    const diff = expires - now

    if (diff > 0 && diff < 300000) {
      toast.warning(`实例 ${instance.challenge_title} 即将在 ${Math.floor(diff / 60000)} 分钟后过期`)
    }
  })
}

function getStatusLabel(status: InstanceStatus): string {
  const labels: Record<InstanceStatus, string> = {
    pending: '等待中',
    creating: '创建中',
    running: '运行中',
    expired: '已过期',
    destroying: '销毁中',
    destroyed: '已销毁',
    failed: '失败',
    crashed: '崩溃',
  }
  return labels[status]
}

function getStatusColor(status: InstanceStatus): string {
  const colors: Record<InstanceStatus, string> = {
    pending: 'info',
    creating: 'warning',
    running: 'success',
    expired: 'info',
    destroying: 'warning',
    destroyed: 'info',
    failed: 'danger',
    crashed: 'danger',
  }
  return colors[status]
}

onMounted(() => {
  refresh()
  timer = window.setInterval(refresh, 10000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

