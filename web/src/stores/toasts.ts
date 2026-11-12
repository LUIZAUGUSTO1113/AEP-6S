import { defineStore } from 'pinia'
import { ref } from 'vue'

const visibleMs = 6000

type Toast = {
  id: string
  text: string
}

export const useToastsStore = defineStore('toasts', () => {
  const items = ref<Toast[]>([])
  const timers = new Map<string, number>()

  function dismiss(id: string) {
    const timer = timers.get(id)
    if (timer !== undefined) {
      window.clearTimeout(timer)
      timers.delete(id)
    }
    items.value = items.value.filter((item) => item.id !== id)
  }

  function show(text: string) {
    const id = crypto.randomUUID()
    items.value = [...items.value, { id, text }]
    timers.set(
      id,
      window.setTimeout(() => dismiss(id), visibleMs),
    )
  }

  return { items, show, dismiss }
})
