<script setup lang="ts">
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import { nextTick, onMounted, onUnmounted, ref, useTemplateRef, watch } from 'vue'
import { BRAZIL, ufView } from '../lib/brazil'
import { geocodeCity } from '../lib/geocode'
import { isAbort } from '../api/client'
import { parseOptionalNumber } from '../lib/input'

const props = defineProps<{
  city: string
  state: string
  lat: string
  lng: string
}>()

const canvas = useTemplateRef<HTMLDivElement>('canvas')
const anchor = useTemplateRef<HTMLDivElement>('anchor')
const caption = ref('Brasil')

let leafletMap: L.Map | undefined
let marker: L.CircleMarker | undefined
let geocodeTimer = 0
let geocodeAbort: AbortController | undefined
let sizeFrame = 0
let heightLocked = false
const narrowQuery = window.matchMedia('(max-width: 960px)')

function fitMap() {
  cancelAnimationFrame(sizeFrame)
  sizeFrame = requestAnimationFrame(() => {
    leafletMap?.invalidateSize()
  })
}

function lockToForm() {
  const el = anchor.value
  if (!el) {
    return
  }
  if (narrowQuery.matches) {
    heightLocked = false
    el.style.height = ''
    fitMap()
    return
  }
  if (heightLocked) {
    fitMap()
    return
  }
  const form = el.parentElement?.querySelector('form')
  if (!(form instanceof HTMLElement)) {
    return
  }
  const height = Math.round(form.getBoundingClientRect().height)
  if (height < 120) {
    return
  }
  heightLocked = true
  el.style.height = `${height}px`
  fitMap()
}

function reduceMotion() {
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches
}

function goTo(lat: number, lng: number, zoom: number) {
  if (!leafletMap) {
    return
  }
  if (reduceMotion()) {
    leafletMap.setView([lat, lng], zoom)
    return
  }
  leafletMap.flyTo([lat, lng], zoom)
}

function setMarker(lat: number, lng: number) {
  if (!leafletMap) {
    return
  }
  if (marker) {
    marker.setLatLng([lat, lng])
    return
  }
  marker = L.circleMarker([lat, lng], {
    radius: 8,
    color: '#0f4c5c',
    weight: 2,
    fillColor: '#0f4c5c',
    fillOpacity: 0.85,
  }).addTo(leafletMap)
}

function clearMarker() {
  marker?.remove()
  marker = undefined
}

function showUf(name: string, lat: number, lng: number, zoom: number) {
  goTo(lat, lng, zoom)
  clearMarker()
  caption.value = name
}

async function lookupCity(city: string, stateName: string, lat: number, lng: number, zoom: number) {
  geocodeAbort = new AbortController()
  try {
    const point = await geocodeCity(city, stateName, geocodeAbort.signal)
    if (point) {
      goTo(point.lat, point.lng, 12)
      setMarker(point.lat, point.lng)
      caption.value = point.label
      return
    }
  } catch (err) {
    if (isAbort(err)) {
      return
    }
  }
  showUf(stateName, lat, lng, zoom)
}

function sync() {
  const lat = parseOptionalNumber(props.lat)
  const lng = parseOptionalNumber(props.lng)
  const exact =
    lat !== undefined && lng !== undefined && !Number.isNaN(lat) && !Number.isNaN(lng)

  window.clearTimeout(geocodeTimer)
  geocodeAbort?.abort()

  if (exact) {
    goTo(lat, lng, 15)
    setMarker(lat, lng)
    caption.value = `${lat}, ${lng}`
    return
  }

  const uf = ufView(props.state)
  const city = props.city.trim()

  if (uf && city.length >= 3) {
    geocodeTimer = window.setTimeout(() => {
      void lookupCity(city, uf.name, uf.lat, uf.lng, uf.zoom)
    }, 700)
    return
  }

  if (uf) {
    showUf(uf.name, uf.lat, uf.lng, uf.zoom)
    return
  }

  showUf('Brasil', BRAZIL.lat, BRAZIL.lng, BRAZIL.zoom)
}

onMounted(() => {
  if (!canvas.value) {
    return
  }
  leafletMap = L.map(canvas.value, { scrollWheelZoom: false }).setView(
    [BRAZIL.lat, BRAZIL.lng],
    BRAZIL.zoom,
  )
  leafletMap.scrollWheelZoom.disable()
  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>',
    maxZoom: 19,
  }).addTo(leafletMap)
  void nextTick(() => lockToForm())
  window.addEventListener('resize', lockToForm)
  narrowQuery.addEventListener('change', lockToForm)
  sync()
})

watch(
  () => [props.city, props.state, props.lat, props.lng],
  () => {
    sync()
  },
)

onUnmounted(() => {
  window.clearTimeout(geocodeTimer)
  geocodeAbort?.abort()
  cancelAnimationFrame(sizeFrame)
  window.removeEventListener('resize', lockToForm)
  narrowQuery.removeEventListener('change', lockToForm)
  leafletMap?.remove()
  leafletMap = undefined
  marker = undefined
})
</script>

<template>
  <div
    ref="anchor"
    class="station-map-anchor"
  >
    <div
      class="station-map"
      role="img"
      :aria-label="`Mapa da localização: ${caption}`"
    >
      <div
        ref="canvas"
        class="station-map-canvas"
      />
    </div>
  </div>
</template>

<style scoped>
.station-map-anchor {
  min-width: 0;
  height: 22rem;
}

.station-map {
  height: 100%;
  border: 1px solid var(--line);
  border-radius: var(--radius);
  overflow: hidden;
  background: var(--paper);
}

.station-map-canvas {
  width: 100%;
  height: 100%;
}

.station-map :deep(.leaflet-control-zoom a) {
  width: 44px;
  height: 44px;
  line-height: 44px;
}

@media (max-width: 960px) {
  .station-map-anchor {
    height: 16rem;
  }
}

.station-map :deep(.leaflet-container) {
  font-family: var(--font);
  background: var(--paper);
}

.station-map :deep(.leaflet-control-attribution) {
  background: var(--surface);
  color: var(--ink-muted);
}
</style>
