<script setup lang="ts">
import { computed } from 'vue'

import type { TeacherClassReviewData } from '@/api/contracts'
import AppEmpty from '@/components/common/AppEmpty.vue'
import AppCard from '@/components/common/AppCard.vue'
import SectionCard from '@/components/common/SectionCard.vue'

const props = defineProps<{
  review: TeacherClassReviewData | null
  className?: string
}>()

const reviewItems = computed(() => props.review?.items ?? [])
</script>

<template>
  <SectionCard
    title="教学复盘结论"
    :subtitle="
      className
        ? `${className} 当前班级可直接执行的复盘结论与介入建议。`
        : '当前班级可直接执行的复盘结论与介入建议。'
    "
  >
    <AppEmpty
      v-if="reviewItems.length === 0"
      icon="FileChartColumnIncreasing"
      title="暂无复盘结论"
      description="当前班级还没有足够的训练数据形成稳定结论。"
    />

    <div v-else class="grid gap-3 lg:grid-cols-2">
      <AppCard
        v-for="item in reviewItems"
        :key="item.key"
        variant="action"
        :accent="item.accent"
      >
        <div class="text-sm font-semibold text-text-primary">{{ item.title }}</div>
        <div class="mt-2 text-sm leading-6 text-text-secondary">{{ item.detail }}</div>

        <div
          v-if="item.students && item.students.length > 0"
          class="mt-3 flex flex-wrap gap-2"
        >
          <span
            v-for="student in item.students"
            :key="student.id"
            class="inline-flex items-center rounded-full border border-primary/18 bg-primary/10 px-3 py-1 text-xs font-medium text-primary"
          >
            {{ student.name || student.username }}
          </span>
        </div>

        <div
          v-if="item.recommendation"
          class="mt-3 rounded-2xl border border-border bg-base/60 px-4 py-3"
        >
          <div class="text-[11px] font-semibold uppercase tracking-[0.18em] text-primary/80">
            推荐训练题
          </div>
          <div class="mt-2 text-sm font-semibold text-text-primary">
            {{ item.recommendation.title }}
          </div>
          <div class="mt-1 text-xs text-text-secondary">
            {{ item.recommendation.category }} / {{ item.recommendation.difficulty }}
          </div>
          <div class="mt-2 text-sm leading-6 text-text-secondary">
            {{ item.recommendation.reason }}
          </div>
        </div>
      </AppCard>
    </div>
  </SectionCard>
</template>
