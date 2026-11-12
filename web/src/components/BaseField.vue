<script setup lang="ts">
import { computed, useId } from 'vue'

const props = defineProps<{
  label: string
  hint?: string
  error?: string
  optional?: boolean
}>()

const fieldId = useId()
const hintId = useId()
const errorId = useId()

const describedBy = computed(() => {
  const ids: string[] = []
  if (props.hint) {
    ids.push(hintId)
  }
  if (props.error) {
    ids.push(errorId)
  }
  return ids.join(' ') || undefined
})
</script>

<template>
  <div :class="['field', { 'field--invalid': error }]">
    <label :for="fieldId">
      {{ label }}<template v-if="optional"> (opcional)</template>
    </label>
    <slot
      :described-by="describedBy"
      :id="fieldId"
      :invalid="Boolean(error)"
    />
    <p
      v-if="hint"
      :id="hintId"
      class="hint"
    >
      {{ hint }}
    </p>
    <p
      v-if="error"
      :id="errorId"
      class="field-error"
    >
      {{ error }}
    </p>
  </div>
</template>

<style scoped>
.field {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

label {
  font-weight: 600;
}

.field :deep(input),
.field :deep(select) {
  width: 100%;
  min-height: 44px;
  padding: var(--space-2);
  border: 1px solid var(--line);
  border-radius: var(--radius);
  background-color: var(--surface);
  color: var(--ink);
  font: inherit;
  font-size: 1rem;
}

.field :deep(select) {
  appearance: none;
  padding-right: var(--space-6);
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='8' viewBox='0 0 12 8' fill='none'%3E%3Cpath d='M1 1.5 6 6.5 11 1.5' stroke='%231a1814' stroke-width='1.5'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right var(--space-3) center;
  background-size: 0.75rem 0.5rem;
}

.field--invalid :deep(input),
.field--invalid :deep(select) {
  border-color: var(--alert-fg);
}
</style>
