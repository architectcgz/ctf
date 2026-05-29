<script setup lang="ts">
defineProps<{
  form: {
    attack_score: number
    defense_score: number
  }
  fieldErrors: {
    attack_score: string
    defense_score: string
  }
}>()
</script>

<template>
  <div class="grid gap-4 sm:grid-cols-2">
    <div class="ui-field awd-operations-field">
      <label
        class="ui-field__label"
        for="awd-attack-score"
      >攻击分</label>
      <span
        class="ui-control-wrap"
        :class="{ 'is-error': !!fieldErrors.attack_score }"
      >
        <input
          id="awd-attack-score"
          v-model.number="form.attack_score"
          type="number"
          min="0"
          step="1"
          class="ui-control"
        >
      </span>
      <p
        v-if="fieldErrors.attack_score"
        class="ui-field__error"
      >
        {{ fieldErrors.attack_score }}
      </p>
    </div>

    <div class="ui-field awd-operations-field">
      <label
        class="ui-field__label"
        for="awd-defense-score"
      >防守分</label>
      <span
        class="ui-control-wrap"
        :class="{ 'is-error': !!fieldErrors.defense_score }"
      >
        <input
          id="awd-defense-score"
          v-model.number="form.defense_score"
          type="number"
          min="0"
          step="1"
          class="ui-control"
        >
      </span>
      <p
        v-if="fieldErrors.defense_score"
        class="ui-field__error"
      >
        {{ fieldErrors.defense_score }}
      </p>
    </div>
  </div>
</template>
