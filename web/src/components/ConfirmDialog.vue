<script setup lang="ts">
import { onMounted, onUnmounted, useId, useTemplateRef } from 'vue'
import BaseButton from './BaseButton.vue'

defineProps<{
  kind: string
  question: string
}>()

const emit = defineEmits<{
  confirm: []
  cancel: []
}>()

const titleId = useId()
const bodyId = useId()
const dialog = useTemplateRef<HTMLElement>('dialog')

function onKey(event: KeyboardEvent) {
  if (event.key === 'Escape') {
    emit('cancel')
  }
}

onMounted(() => {
  document.addEventListener('keydown', onKey)
  dialog.value?.querySelector<HTMLButtonElement>('button')?.focus()
})

onUnmounted(() => {
  document.removeEventListener('keydown', onKey)
})
</script>

<template>
  <Teleport to="body">
    <div
      class="confirm-backdrop"
      @click.self="emit('cancel')"
    >
      <div
        ref="dialog"
        class="confirm-dialog"
        role="dialog"
        aria-modal="true"
        :aria-labelledby="titleId"
        :aria-describedby="bodyId"
      >
        <p
          :id="titleId"
          class="confirm-kind"
        >
          {{ kind }}
        </p>
        <p
          :id="bodyId"
          class="confirm-question"
        >
          {{ question }}
        </p>
        <p class="confirm-warning">Ação irreversível!</p>
        <div class="confirm-actions">
          <BaseButton @click="emit('cancel')">Cancelar</BaseButton>
          <BaseButton
            variant="primary"
            @click="emit('confirm')"
          >
            Excluir
          </BaseButton>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.confirm-backdrop {
  position: fixed;
  inset: 0;
  z-index: 1200;
  display: grid;
  place-items: center;
  padding: var(--space-5);
  background: rgb(26 24 20 / 0.4);
  backdrop-filter: blur(6px);
}

.confirm-dialog {
  width: min(28rem, 100%);
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  padding: var(--space-5);
  background: var(--surface);
  border: 1px solid var(--line);
  border-radius: var(--radius);
}

.confirm-kind {
  margin: 0;
  font-size: 1.125rem;
  font-weight: 600;
}

.confirm-question {
  margin: 0;
}

.confirm-warning {
  margin: 0;
  color: var(--alert-fg);
  font-weight: 600;
}

.confirm-actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-3);
}

@media (max-width: 960px) {
  .confirm-dialog {
    width: 100%;
  }

  .confirm-actions {
    flex-direction: column;
    align-items: stretch;
  }

  .confirm-actions :deep(.base-button) {
    width: 100%;
  }
}
</style>
