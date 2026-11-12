export type GeoPoint = {
  lat: number
  lng: number
  label: string
}

const cache = new Map<string, GeoPoint | null>()

type NominatimHit = {
  lat: string
  lon: string
  display_name?: string
}

export async function geocodeCity(
  city: string,
  stateName: string,
  signal?: AbortSignal,
): Promise<GeoPoint | null> {
  const key = `${city.trim().toLowerCase()}|${stateName.toLowerCase()}`
  if (cache.has(key)) {
    return cache.get(key) ?? null
  }

  const params = new URLSearchParams({
    city: city.trim(),
    state: stateName,
    countrycodes: 'br',
    format: 'jsonv2',
    limit: '1',
  })
  const response = await fetch(`https://nominatim.openstreetmap.org/search?${params}`, {
    signal,
    headers: { Accept: 'application/json' },
  })
  if (!response.ok) {
    return null
  }

  const rows = (await response.json()) as NominatimHit[]
  const first = rows[0]
  const point = first
    ? { lat: Number(first.lat), lng: Number(first.lon), label: first.display_name ?? city }
    : null
  cache.set(key, point)
  return point
}
