<template>
  <component :is="tag" class="student-insight-state-surface" :class="stateClasses" v-bind="$attrs">
    <template v-if="loading">
      <slot name="loading" />
    </template>

    <template v-else-if="empty">
      <slot name="empty" />
    </template>

    <template v-else>
      <slot />
    </template>
  </component>
</template>

<style src="./studentInsightSurface.css"></style>

<script setup lang="ts">
import { computed } from 'vue'

defineOptions({
  inheritAttrs: false,
})

const props = withDefaults(
  defineProps<{
    loading?: boolean
    empty?: boolean
    tag?: string
    surface?: 'glass' | 'plain'
  }>(),
  {
    loading: false,
    empty: false,
    tag: 'div',
    surface: 'glass',
  }
)

const stateClasses = computed(() => ({
  'student-insight-state-surface--glass': props.surface === 'glass',
  'student-insight-state-surface--plain': props.surface === 'plain',
  'student-insight-state-surface--loading': props.loading,
  'student-insight-state-surface--empty': !props.loading && props.empty,
  'student-insight-state-surface--content': !props.loading && !props.empty,
}))
</script>
