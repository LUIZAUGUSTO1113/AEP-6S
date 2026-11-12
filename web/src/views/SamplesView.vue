<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import {
  errorMessage,
  isAbort,
  samplesApi,
  type Sample,
  type SamplePayload,
} from '../api/client'
import BaseButton from '../components/BaseButton.vue'
import BaseField from '../components/BaseField.vue'
import ConfirmDialog from '../components/ConfirmDialog.vue'
import RowActions from '../components/RowActions.vue'
import StationMap from '../components/StationMap.vue'
import StationSelect from '../components/StationSelect.vue'
import {
  formatDate,
  fromDateTimeLocal,
  isFillPrompt,
  isPh,
  parseOptionalNumber,
  phOutOfRange,
  stationCaption,
  toDateTimeLocal,
} from '../lib/input'
import { useRevealForm } from '../composables/useRevealForm'
import { useStationsStore } from '../stores/stations'
import { useToastsStore } from '../stores/toasts'

const stations = useStationsStore()
const toasts = useToastsStore()
const samples = ref<Sample[]>([])
const loading = ref(true)
const listError = ref('')
const filterId = ref('')
const stationId = ref('')
const parameter = ref('')
const value = ref('')
const collectedAt = ref('')
const editingId = ref('')
const errors = ref<Record<string, string>>({})
const saving = ref(false)
const pendingDelete = ref<Sample | null>(null)
const form = ref<HTMLFormElement | null>(null)
const { reveal } = useRevealForm()

function labelFor(sample: Sample): string {
  if (!sample.station_id) {
    return sample.river || 'Sem estação'
  }
  return stationCaption(stations.byId(sample.station_id), sample.river || sample.station_id)
}

async function loadSamples(signal?: AbortSignal) {
  listError.value = ''
  loading.value = true
  try {
    samples.value = await samplesApi.list(filterId.value || undefined, signal)
  } catch (err) {
    if (isAbort(err)) {
      return
    }
    listError.value = errorMessage(err)
  } finally {
    if (!signal?.aborted) {
      loading.value = false
    }
  }
}

onMounted(() => {
  void stations.load()
})

watch(
  filterId,
  (_id, _prev, onCleanup) => {
    const controller = new AbortController()
    onCleanup(() => controller.abort())
    void loadSamples(controller.signal)
  },
  { immediate: true },
)

function resetForm() {
  stationId.value = ''
  parameter.value = ''
  value.value = ''
  collectedAt.value = ''
  editingId.value = ''
  errors.value = {}
}

function startEdit(sample: Sample) {
  editingId.value = sample.id
  stationId.value = sample.station_id ?? ''
  parameter.value = sample.parameter
  value.value = String(sample.value)
  collectedAt.value = sample.collected_at ? toDateTimeLocal(sample.collected_at) : ''
  errors.value = {}
  void reveal()
}

function dismissPrompt(key: string) {
  if (!isFillPrompt(errors.value[key])) {
    return
  }
  const next = { ...errors.value }
  delete next[key]
  errors.value = next
}

function validate(): Record<string, string> {
  const next: Record<string, string> = {}
  if (stations.items.length === 0) {
    next.stationId = 'Cadastre uma estação antes de lançar a coleta.'
    return next
  }
  if (!stationId.value) next.stationId = 'Informe a estação.'
  if (!parameter.value.trim()) next.parameter = 'Informe o parâmetro.'
  if (!value.value.trim()) {
    next.value = 'Informe o valor.'
  } else {
    const parsed = parseOptionalNumber(value.value)
    if (Number.isNaN(parsed as number)) {
      next.value = 'Informe um número.'
    } else if (isPh(parameter.value) && phOutOfRange(parsed as number)) {
      next.value = 'O pH precisa ficar entre 0 e 14.'
    }
  }
  return next
}

function toPayload(): SamplePayload {
  const payload: SamplePayload = {
    station_id: stationId.value,
    parameter: parameter.value.trim(),
    value: parseOptionalNumber(value.value) as number,
  }
  if (collectedAt.value) {
    payload.collected_at = fromDateTimeLocal(collectedAt.value)
  }
  return payload
}

async function onSubmit() {
  errors.value = validate()
  if (Object.keys(errors.value).length > 0) {
    void nextTick(() => form.value?.querySelector<HTMLElement>('[aria-invalid="true"]')?.focus())
    return
  }
  saving.value = true
  try {
    if (editingId.value) {
      await samplesApi.update(editingId.value, toPayload())
    } else {
      await samplesApi.create(toPayload())
    }
    resetForm()
    await loadSamples()
  } catch (err) {
    toasts.show(errorMessage(err))
  } finally {
    saving.value = false
  }
}

async function confirmDelete() {
  const sample = pendingDelete.value
  pendingDelete.value = null
  if (!sample) {
    return
  }
  try {
    await samplesApi.remove(sample.id)
    if (editingId.value === sample.id) {
      resetForm()
    }
    await loadSamples()
  } catch (err) {
    toasts.show(errorMessage(err))
  }
}

const emptyMessage = computed(() =>
  filterId.value ? 'Nenhuma coleta para esta estação.' : 'Nenhuma coleta registrada.',
)

const mapLocation = computed(() => {
  const place = stations.byId(stationId.value)?.location
  if (!place) {
    return { city: '', state: '', lat: '', lng: '' }
  }
  return {
    city: place.city,
    state: place.state,
    lat: place.lat === 0 ? '' : String(place.lat),
    lng: place.lng === 0 ? '' : String(place.lng),
  }
})
</script>

<template>
  <div class="page">
    <div
      class="editor"
      ref="editor"
    >
      <div>
        <h1>Coletas</h1>
        <p class="lede">Lance a medição de uma estação já cadastrada.</p>
      </div>
      <div class="station-compose">
      <form
        ref="form"
        class="form"
        novalidate
        @submit.prevent="onSubmit"
      >
      <BaseField
        v-slot="{ id, describedBy, invalid }"
        :error="errors.stationId"
        label="Estação"
      >
        <StationSelect
          :described-by="describedBy"
          :id="id"
          :invalid="invalid"
          :stations="stations.items"
          v-model="stationId"
          @change="dismissPrompt('stationId')"
        />
      </BaseField>
      <BaseField
        v-slot="{ id, describedBy, invalid }"
        :error="errors.parameter"
        hint="pH fica entre 0 e 14."
        label="Parâmetro"
      >
        <input
          :id="id"
          :aria-describedby="describedBy"
          :aria-invalid="invalid"
          v-model="parameter"
          @input="dismissPrompt('parameter')"
        >
      </BaseField>
      <BaseField
        v-slot="{ id, describedBy, invalid }"
        :error="errors.value"
        label="Valor"
      >
        <input
          :id="id"
          :aria-describedby="describedBy"
          :aria-invalid="invalid"
          inputmode="decimal"
          v-model="value"
          @input="dismissPrompt('value')"
        >
      </BaseField>
      <BaseField
        v-slot="{ id, describedBy, invalid }"
        hint="Se ficar vazio, a API usa o horário do envio."
        label="Data da coleta"
        optional
      >
        <input
          :id="id"
          :aria-describedby="describedBy"
          :aria-invalid="invalid"
          type="datetime-local"
          v-model="collectedAt"
        >
      </BaseField>
      <div class="form-actions">
        <BaseButton
          :disabled="saving"
          type="submit"
          variant="primary"
        >
          {{ saving ? 'Salvando…' : editingId ? 'Salvar alterações' : 'Salvar coleta' }}
        </BaseButton>
        <BaseButton
          v-if="editingId"
          :disabled="saving"
          @click="resetForm"
        >
          Cancelar
        </BaseButton>
      </div>
    </form>
      <StationMap
        :city="mapLocation.city"
        :lat="mapLocation.lat"
        :lng="mapLocation.lng"
        :state="mapLocation.state"
      />
      </div>
    </div>
    <section>
      <h2>Lista</h2>
      <div class="filter">
        <BaseField
          v-slot="{ id, describedBy }"
          label="Filtrar por estação"
        >
          <StationSelect
            :described-by="describedBy"
            :id="id"
            :stations="stations.items"
            include-all
            v-model="filterId"
          />
        </BaseField>
      </div>
      <p
        v-if="loading"
        class="status"
      >
        Carregando coletas.
      </p>
      <p
        v-else-if="listError"
        class="status-error"
      >
        {{ listError }}
      </p>
      <p
        v-else-if="samples.length === 0"
        class="status"
      >
        {{ emptyMessage }}
      </p>
      <div
        v-else
        class="table-block"
      >
        <p class="table-scroll-hint">A tabela continua para o lado.</p>
        <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Estação</th>
              <th>Parâmetro</th>
              <th class="num">Valor</th>
              <th class="num">Data</th>
              <th>Ações</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="sample in samples"
              :key="sample.id"
            >
              <td>{{ labelFor(sample) }}</td>
              <td>{{ sample.parameter }}</td>
              <td class="num">{{ sample.value }}</td>
              <td class="num">{{ formatDate(sample.collected_at) }}</td>
              <td>
                <RowActions
                  @edit="startEdit(sample)"
                  @remove="pendingDelete = sample"
                />
              </td>
            </tr>
          </tbody>
        </table>
        </div>
      </div>
    </section>
    <ConfirmDialog
      v-if="pendingDelete"
      kind="Coleta"
      :question="`Você deseja realmente excluir a coleta de ${pendingDelete.parameter} em ${labelFor(pendingDelete)}?`"
      @cancel="pendingDelete = null"
      @confirm="confirmDelete"
    />
  </div>
</template>
