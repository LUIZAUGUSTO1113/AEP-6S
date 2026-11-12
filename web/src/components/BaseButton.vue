<script setup lang="ts">
withDefaults(
  defineProps<{
    type?: 'button' | 'submit'
    variant?: 'primary' | 'ghost'
    disabled?: boolean
  }>(),
  {
    type: 'button',
    variant: 'ghost',
    disabled: false,
  },
)
</script>

<template>
  <button
    :class="['base-button', `base-button--${variant}`]"
    :disabled="disabled"
    :type="type"
  >
    <slot />
  </button>
</template>

<style scoped>
.base-button {
  margin: 0;
  min-height: 44px;
  padding: var(--space-2) var(--space-4);
  border: 1px solid transparent;
  border-radius: var(--radius);
  font: inherit;
  cursor: pointer;
}

.base-button:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.base-button--primary {
  background: var(--river);
  color: var(--surface);
}

.base-button--primary:hover:not(:disabled) {
  background: var(--river-press);
}

.base-button--ghost {
  background: transparent;
  border-color: var(--line);
  color: var(--river);
}

.base-button--ghost:hover:not(:disabled) {
  border-color: var(--river);
}
</style>
