import { API_UNAVAILABLE, translateApiError } from '../lib/errors'

function baseUrl(): string {
  return (import.meta.env.VITE_API_URL ?? 'http://localhost:8080').replace(/\/$/, '')
}

export function isAbort(err: unknown): boolean {
  return err instanceof DOMException && err.name === 'AbortError'
}

export function errorMessage(err: unknown): string {
  if (isAbort(err)) {
    throw err
  }
  return translateApiError(err instanceof Error ? err.message : API_UNAVAILABLE)
}

async function readError(response: Response): Promise<string> {
  try {
    const body: unknown = await response.json()
    if (
      body !== null &&
      typeof body === 'object' &&
      'error' in body &&
      typeof body.error === 'string' &&
      body.error !== ''
    ) {
      return body.error
    }
  } catch {
    return API_UNAVAILABLE
  }
  return API_UNAVAILABLE
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  let response: Response
  try {
    response = await fetch(`${baseUrl()}${path}`, init)
  } catch (err) {
    if (isAbort(err)) {
      throw err
    }
    throw new Error(API_UNAVAILABLE)
  }

  if (!response.ok) {
    throw new Error(await readError(response))
  }

  if (response.status === 204) {
    return undefined as T
  }

  try {
    return (await response.json()) as T
  } catch {
    throw new Error(API_UNAVAILABLE)
  }
}

function jsonInit(method: string, body: unknown, signal?: AbortSignal): RequestInit {
  return {
    method,
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
    signal,
  }
}

function withStation(path: string, stationId?: string): string {
  if (!stationId) {
    return path
  }
  return `${path}?station_id=${encodeURIComponent(stationId)}`
}

export type Threshold = {
  parameter: string
  min?: number
  max?: number
}

export type StationLocation = {
  city: string
  state: string
  lat: number
  lng: number
}

export type Station = {
  id: string
  name: string
  river: string
  code: string
  location: StationLocation
  thresholds: Threshold[]
}

export type StationPayload = {
  name: string
  river: string
  code: string
  location: StationLocation
  thresholds: Threshold[]
}

export type Sample = {
  id: string
  station_id?: string
  river: string
  parameter: string
  value: number
  collected_at: string
}

export type SamplePayload = {
  station_id: string
  parameter: string
  value: number
  collected_at?: string
}

export type Alert = {
  id: string
  sample_id: string
  station_id: string
  parameter: string
  value: number
  message: string
  created_at: string
}

export const stationsApi = {
  list: (signal?: AbortSignal) => request<Station[]>('/stations', { signal }),
  create: (body: StationPayload) => request<Station>('/stations', jsonInit('POST', body)),
  update: (id: string, body: StationPayload) =>
    request<Station>(`/stations/${id}`, jsonInit('PUT', body)),
  remove: (id: string) => request<void>(`/stations/${id}`, { method: 'DELETE' }),
}

export const samplesApi = {
  list: (stationId?: string, signal?: AbortSignal) =>
    request<Sample[]>(withStation('/samples', stationId), { signal }),
  create: (body: SamplePayload) => request<Sample>('/samples', jsonInit('POST', body)),
  update: (id: string, body: SamplePayload) =>
    request<Sample>(`/samples/${id}`, jsonInit('PUT', body)),
  remove: (id: string) => request<void>(`/samples/${id}`, { method: 'DELETE' }),
}

export const alertsApi = {
  list: (stationId?: string, signal?: AbortSignal) =>
    request<Alert[]>(withStation('/alerts', stationId), { signal }),
}
