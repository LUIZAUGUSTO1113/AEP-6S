import { defineStore } from 'pinia'
import { ref } from 'vue'
import { alertsApi, errorMessage, isAbort, type Alert } from '../api/client'

export const useAlertsStore = defineStore('alerts', () => {
  const items = ref<Alert[]>([])
  const loading = ref(true)
  const error = ref('')
  const stationId = ref('')

  async function load(signal?: AbortSignal) {
    error.value = ''
    loading.value = true
    try {
      items.value = await alertsApi.list(stationId.value || undefined, signal)
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

  return { items, loading, error, stationId, load }
})
