<script setup lang="ts">
import type { Station } from '../api/client'
import { stationCaption } from '../lib/input'

defineProps<{
  stations: Station[]
  id?: string
  describedBy?: string
  invalid?: boolean
  includeAll?: boolean
}>()

const model = defineModel<string>({ default: '' })
</script>

<template>
  <select
    :id="id"
    :aria-describedby="describedBy"
    :aria-invalid="invalid ? true : undefined"
    v-model="model"
  >
    <option
      v-if="includeAll"
      value=""
    >
      Todas
    </option>
    <option
      v-else
      value=""
    >
      Selecione
    </option>
    <option
      v-for="station in stations"
      :key="station.id"
      :value="station.id"
    >
      {{ stationCaption(station) }}
    </option>
  </select>
</template>
