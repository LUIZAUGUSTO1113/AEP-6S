import { defineStore } from 'pinia'
import { ref } from 'vue'
import {
  errorMessage,
  isAbort,
  stationsApi,
  type Station,
  type StationPayload,
} from '../api/client'

export const useStationsStore = defineStore('stations', () => {
  const items = ref<Station[]>([])
  const loading = ref(true)
  const error = ref('')

  async function load(signal?: AbortSignal) {
    error.value = ''
    if (items.value.length === 0) {
      loading.value = true
    }
    try {
      items.value = await stationsApi.list(signal)
    } catch (err) {
      if (isAbort(err)) {
        return
      }
      error.value = errorMessage(err)
    } finally {
      if (!signal?.aborted) {
        loading.value = false
      }
    }
  }

  function byId(id: string) {
    return items.value.find((station) => station.id === id)
  }

  async function create(payload: StationPayload) {
    await stationsApi.create(payload)
    await load()
  }

  async function update(id: string, payload: StationPayload) {
    await stationsApi.update(id, payload)
    await load()
  }

  async function remove(id: string) {
    await stationsApi.remove(id)
    await load()
  }

  return { items, loading, error, load, byId, create, update, remove }
})
