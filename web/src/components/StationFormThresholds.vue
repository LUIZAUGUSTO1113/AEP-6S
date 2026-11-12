<script setup lang="ts">
import { nextTick, useTemplateRef } from 'vue'
import { revealElement } from '../composables/useRevealForm'
import BaseButton from './BaseButton.vue'
import BaseField from './BaseField.vue'

export type ThresholdDraft = {
  key: string
  parameter: string
  min: string
  max: string
}

defineProps<{
  errors: Record<string, string>
}>()

const thresholds = defineModel<ThresholdDraft[]>({ required: true })
const list = useTemplateRef<HTMLElement>('thresholds')
const emit = defineEmits<{
  dismiss: [key: string]
}>()

async function add() {
  thresholds.value = [
    ...thresholds.value,
    { key: crypto.randomUUID(), parameter: '', min: '', max: '' },
  ]
  await nextTick()
  const rows = list.value?.querySelectorAll<HTMLElement>('.threshold')
  await revealElement(rows?.[rows.length - 1], 'nearest')
}

function remove(key: string) {
  thresholds.value = thresholds.value.filter((item) => item.key !== key)
}
</script>

<template>
  <fieldset
    ref="thresholds"
    class="thresholds"
  >
    <legend>Limiares</legend>
    <p class="hint">
      Sem limiar para o parâmetro, a coleta não gera alerta. O mínimo não pode
      passar do máximo. pH fica entre 0 e 14.
    </p>
    <div
      v-for="item in thresholds"
      :key="item.key"
      class="threshold"
    >
      <BaseField
        v-slot="{ id, describedBy, invalid }"
        :error="errors[`${item.key}:parameter`]"
        label="Parâmetro"
      >
        <input
          :id="id"
          :aria-describedby="describedBy"
          :aria-invalid="invalid"
          v-model="item.parameter"
          @input="emit('dismiss', `${item.key}:parameter`)"
        >
      </BaseField>
      <div class="pair-even">
        <BaseField
          v-slot="{ id, describedBy, invalid }"
          :error="errors[`${item.key}:min`] || errors[`${item.key}:range`]"
          label="Mínimo"
          optional
        >
          <input
            :id="id"
            :aria-describedby="describedBy"
            :aria-invalid="invalid"
            inputmode="decimal"
            v-model="item.min"
            @input="emit('dismiss', `${item.key}:min`)"
          >
        </BaseField>
        <BaseField
          v-slot="{ id, describedBy, invalid }"
          :error="errors[`${item.key}:max`]"
          label="Máximo"
          optional
        >
          <input
            :id="id"
            :aria-describedby="describedBy"
            :aria-invalid="invalid"
            inputmode="decimal"
            v-model="item.max"
            @input="emit('dismiss', `${item.key}:max`)"
          >
        </BaseField>
      </div>
      <BaseButton @click="remove(item.key)">Remover</BaseButton>
    </div>
    <BaseButton @click="add">Adicionar limiar</BaseButton>
  </fieldset>
</template>

<style scoped>
.thresholds {
  margin: 0;
  padding: 0;
  border: 0;
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

legend {
  padding: 0;
  font-weight: 600;
}

.threshold {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  padding-top: var(--space-3);
  border-top: 1px solid var(--line);
}
</style>
