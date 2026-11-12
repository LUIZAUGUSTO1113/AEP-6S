<script setup lang="ts">
import { computed, onMounted, watch } from 'vue'
import BaseField from '../components/BaseField.vue'
import StationSelect from '../components/StationSelect.vue'
import { formatDate, stationCaption } from '../lib/input'
import { useAlertsStore } from '../stores/alerts'
import { useStationsStore } from '../stores/stations'

const stations = useStationsStore()
const alerts = useAlertsStore()

const emptyMessage = computed(() =>
  alerts.stationId ? 'Nenhum alerta para esta estação.' : 'Nenhum alerta registrado.',
)

function labelFor(stationId: string): string {
  return stationCaption(stations.byId(stationId), stationId)
}

onMounted(() => {
  void stations.load()
})

watch(
  () => alerts.stationId,
  (_id, _prev, onCleanup) => {
    const controller = new AbortController()
    onCleanup(() => controller.abort())
    void alerts.load(controller.signal)
  },
  { immediate: true },
)
</script>

<template>
  <div class="page page-alerts">
    <div>
      <h1>Alertas</h1>
      <p class="lede">Alertas gerados quando a coleta fica fora do limiar da estação.</p>
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
            v-model="alerts.stationId"
          />
        </BaseField>
      </div>
      <p
        v-if="alerts.loading"
        class="status"
      >
        Carregando alertas.
      </p>
      <p
        v-else-if="alerts.error"
        class="status-error"
      >
        {{ alerts.error }}
      </p>
      <p
        v-else-if="alerts.items.length === 0"
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
              <th>Mensagem</th>
              <th class="num">Data</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="alert in alerts.items"
              :key="alert.id"
            >
              <td>{{ labelFor(alert.station_id) }}</td>
              <td>{{ alert.parameter }}</td>
              <td class="num">{{ alert.value }}</td>
              <td>{{ alert.message }}</td>
              <td class="num">{{ formatDate(alert.created_at) }}</td>
            </tr>
          </tbody>
        </table>
        </div>
      </div>
    </section>
  </div>
</template>
