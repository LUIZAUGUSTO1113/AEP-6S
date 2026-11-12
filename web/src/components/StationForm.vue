<script setup lang="ts">
import { nextTick, ref, useTemplateRef, watch } from 'vue'
import type { Station, StationPayload, Threshold } from '../api/client'
import { isFillPrompt, isPh, parseOptionalNumber, phOutOfRange } from '../lib/input'
import BaseButton from './BaseButton.vue'
import BaseField from './BaseField.vue'
import StationMap from './StationMap.vue'
import StationFormThresholds, { type ThresholdDraft } from './StationFormThresholds.vue'

type StationDraft = {
  name: string
  river: string
  code: string
  city: string
  state: string
  lat: string
  lng: string
  thresholds: ThresholdDraft[]
}

const props = defineProps<{
  station: Station | null
  saving: boolean
}>()

const emit = defineEmits<{
  submit: [payload: StationPayload]
  cancel: []
}>()

function emptyDraft(): StationDraft {
  return {
    name: '',
    river: '',
    code: '',
    city: '',
    state: '',
    lat: '',
    lng: '',
    thresholds: [],
  }
}

function fromStation(station: Station): StationDraft {
  return {
    name: station.name,
    river: station.river,
    code: station.code,
    city: station.location.city,
    state: station.location.state,
    lat: station.location.lat === 0 ? '' : String(station.location.lat),
    lng: station.location.lng === 0 ? '' : String(station.location.lng),
    thresholds: (station.thresholds ?? []).map((item) => ({
      key: crypto.randomUUID(),
      parameter: item.parameter,
      min: item.min == null ? '' : String(item.min),
      max: item.max == null ? '' : String(item.max),
    })),
  }
}

const draft = ref<StationDraft>(props.station ? fromStation(props.station) : emptyDraft())
const errors = ref<Record<string, string>>({})
const form = useTemplateRef<HTMLFormElement>('form')

watch(
  () => props.station,
  (station) => {
    draft.value = station ? fromStation(station) : emptyDraft()
    errors.value = {}
  },
)

function boundError(raw: string, parameter: string): string {
  if (raw.trim() === '') {
    return ''
  }
  const value = parseOptionalNumber(raw)
  if (Number.isNaN(value as number)) {
    return 'Informe um número.'
  }
  if (isPh(parameter) && phOutOfRange(value as number)) {
    return 'O pH precisa ficar entre 0 e 14.'
  }
  return ''
}

function validate(current: StationDraft): Record<string, string> {
  const next: Record<string, string> = {}
  if (!current.name.trim()) next.name = 'Informe o nome da estação.'
  if (!current.river.trim()) next.river = 'Informe o rio.'
  if (!current.code.trim()) next.code = 'Informe o código.'
  if (!current.city.trim()) next.city = 'Informe a cidade.'
  if (!current.state.trim()) next.state = 'Informe a UF.'
  const latError = boundError(current.lat, '')
  if (latError) next.lat = latError
  const lngError = boundError(current.lng, '')
  if (lngError) next.lng = lngError

  for (const item of current.thresholds) {
    const blank = !item.parameter.trim() && !item.min.trim() && !item.max.trim()
    if (blank) {
      continue
    }
    if (!item.parameter.trim()) {
      next[`${item.key}:parameter`] = 'Informe o parâmetro.'
    }
    const minError = boundError(item.min, item.parameter)
    if (minError) next[`${item.key}:min`] = minError
    const maxError = boundError(item.max, item.parameter)
    if (maxError) next[`${item.key}:max`] = maxError
    const min = parseOptionalNumber(item.min)
    const max = parseOptionalNumber(item.max)
    if (
      min !== undefined &&
      max !== undefined &&
      !Number.isNaN(min) &&
      !Number.isNaN(max) &&
      min > max
    ) {
      next[`${item.key}:range`] = 'O mínimo está acima do máximo.'
    }
  }
  return next
}

function toPayload(current: StationDraft): StationPayload {
  const thresholds: Threshold[] = []
  for (const item of current.thresholds) {
    if (!item.parameter.trim() && !item.min.trim() && !item.max.trim()) {
      continue
    }
    const threshold: Threshold = { parameter: item.parameter.trim() }
    const min = parseOptionalNumber(item.min)
    const max = parseOptionalNumber(item.max)
    if (min !== undefined) threshold.min = min
    if (max !== undefined) threshold.max = max
    thresholds.push(threshold)
  }
  return {
    name: current.name.trim(),
    river: current.river.trim(),
    code: current.code.trim(),
    location: {
      city: current.city.trim(),
      state: current.state.trim(),
      lat: parseOptionalNumber(current.lat) ?? 0,
      lng: parseOptionalNumber(current.lng) ?? 0,
    },
    thresholds,
  }
}

function dismissPrompt(key: string) {
  if (!isFillPrompt(errors.value[key])) {
    return
  }
  const next = { ...errors.value }
  delete next[key]
  errors.value = next
}

function onSubmit() {
  errors.value = validate(draft.value)
  if (Object.keys(errors.value).length > 0) {
    void nextTick(() => form.value?.querySelector<HTMLElement>('[aria-invalid="true"]')?.focus())
    return
  }
  emit('submit', toPayload(draft.value))
}
</script>

<template>
  <div class="station-compose">
    <form
      ref="form"
      class="form"
      novalidate
      @submit.prevent="onSubmit"
    >
    <BaseField
      v-slot="{ id, describedBy, invalid }"
      :error="errors.name"
      label="Nome"
    >
      <input
        :id="id"
        :aria-describedby="describedBy"
        :aria-invalid="invalid"
        v-model="draft.name"
        @input="dismissPrompt('name')"
      >
    </BaseField>
    <BaseField
      v-slot="{ id, describedBy, invalid }"
      :error="errors.river"
      label="Rio"
    >
      <input
        :id="id"
        :aria-describedby="describedBy"
        :aria-invalid="invalid"
        v-model="draft.river"
        @input="dismissPrompt('river')"
      >
    </BaseField>
    <BaseField
      v-slot="{ id, describedBy, invalid }"
      :error="errors.code"
      label="Código"
    >
      <input
        :id="id"
        :aria-describedby="describedBy"
        :aria-invalid="invalid"
        v-model="draft.code"
        @input="dismissPrompt('code')"
      >
    </BaseField>
    <div class="pair">
      <BaseField
        v-slot="{ id, describedBy, invalid }"
        :error="errors.city"
        label="Cidade"
      >
        <input
          :id="id"
          :aria-describedby="describedBy"
          :aria-invalid="invalid"
          v-model="draft.city"
          @input="dismissPrompt('city')"
        >
      </BaseField>
      <BaseField
        v-slot="{ id, describedBy, invalid }"
        :error="errors.state"
        label="UF"
      >
        <input
          :id="id"
          :aria-describedby="describedBy"
          :aria-invalid="invalid"
          v-model="draft.state"
          @input="dismissPrompt('state')"
        >
      </BaseField>
    </div>
    <div class="pair-even">
      <BaseField
        v-slot="{ id, describedBy, invalid }"
        :error="errors.lat"
        label="Latitude"
        optional
      >
        <input
          :id="id"
          :aria-describedby="describedBy"
          :aria-invalid="invalid"
          inputmode="decimal"
          v-model="draft.lat"
          @input="dismissPrompt('lat')"
        >
      </BaseField>
      <BaseField
        v-slot="{ id, describedBy, invalid }"
        :error="errors.lng"
        label="Longitude"
        optional
      >
        <input
          :id="id"
          :aria-describedby="describedBy"
          :aria-invalid="invalid"
          inputmode="decimal"
          v-model="draft.lng"
          @input="dismissPrompt('lng')"
        >
      </BaseField>
    </div>
    <StationFormThresholds
      v-model="draft.thresholds"
      :errors="errors"
      @dismiss="dismissPrompt"
    />
    <div class="form-actions">
      <BaseButton
        :disabled="saving"
        type="submit"
        variant="primary"
      >
        {{ saving ? 'Salvando…' : station ? 'Salvar alterações' : 'Salvar estação' }}
      </BaseButton>
      <BaseButton
        v-if="station"
        :disabled="saving"
        @click="emit('cancel')"
      >
        Cancelar
      </BaseButton>
    </div>
    </form>
    <StationMap
      :city="draft.city"
      :lat="draft.lat"
      :lng="draft.lng"
      :state="draft.state"
    />
  </div>
</template>
