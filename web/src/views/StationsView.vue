<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { errorMessage, type Station, type StationPayload } from '../api/client'
import ConfirmDialog from '../components/ConfirmDialog.vue'
import RowActions from '../components/RowActions.vue'
import StationForm from '../components/StationForm.vue'
import { useRevealForm } from '../composables/useRevealForm'
import { useStationsStore } from '../stores/stations'
import { useToastsStore } from '../stores/toasts'

const stations = useStationsStore()
const toasts = useToastsStore()
const editing = ref<Station | null>(null)
const saving = ref(false)
const pendingDelete = ref<Station | null>(null)
const formKey = ref(0)
const { reveal } = useRevealForm()

onMounted(() => {
  void stations.load()
})

function startEdit(station: Station) {
  editing.value = station
  formKey.value += 1
  void reveal()
}

function cancelEdit() {
  editing.value = null
  formKey.value += 1
}

async function save(payload: StationPayload) {
  saving.value = true
  try {
    if (editing.value) {
      await stations.update(editing.value.id, payload)
    } else {
      await stations.create(payload)
    }
    editing.value = null
    formKey.value += 1
  } catch (err) {
    toasts.show(errorMessage(err))
  } finally {
    saving.value = false
  }
}

async function confirmDelete() {
  const station = pendingDelete.value
  pendingDelete.value = null
  if (!station) {
    return
  }
  try {
    await stations.remove(station.id)
    if (editing.value?.id === station.id) {
      cancelEdit()
    }
  } catch (err) {
    toasts.show(errorMessage(err))
  }
}
</script>

<template>
  <div class="page">
    <div
      class="editor"
      ref="editor"
    >
      <div>
        <h1>Estações</h1>
        <p class="lede">Cadastre o ponto de coleta e os limiares de cada parâmetro.</p>
      </div>
      <StationForm
        :key="formKey"
        :saving="saving"
        :station="editing"
        @cancel="cancelEdit"
        @submit="save"
      />
    </div>
    <section>
      <h2>Lista</h2>
      <p
        v-if="stations.loading"
        class="status"
      >
        Carregando estações.
      </p>
      <p
        v-else-if="stations.error"
        class="status-error"
      >
        {{ stations.error }}
      </p>
      <p
        v-else-if="stations.items.length === 0"
        class="status"
      >
        Nenhuma estação cadastrada.
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
              <th>Nome</th>
              <th>Rio</th>
              <th>Código</th>
              <th>Cidade</th>
              <th>UF</th>
              <th>Ações</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="station in stations.items"
              :key="station.id"
            >
              <td>{{ station.name }}</td>
              <td>{{ station.river }}</td>
              <td>{{ station.code }}</td>
              <td>{{ station.location.city }}</td>
              <td>{{ station.location.state }}</td>
              <td>
                <RowActions
                  @edit="startEdit(station)"
                  @remove="pendingDelete = station"
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
      kind="Estação"
      :question="`Você deseja realmente excluir a estação ${pendingDelete.name} (${pendingDelete.code})?`"
      @cancel="pendingDelete = null"
      @confirm="confirmDelete"
    />
  </div>
</template>
