export function isFillPrompt(message: string | undefined): boolean {
  return !!message && (message.startsWith('Informe') || message.startsWith('Cadastre uma'))
}

export function parseOptionalNumber(raw: string): number | undefined {
  const trimmed = raw.trim().replace(',', '.')
  if (trimmed === '') {
    return undefined
  }
  const value = Number(trimmed)
  return Number.isFinite(value) ? value : Number.NaN
}

export function isPh(parameter: string): boolean {
  return parameter.trim().toLowerCase() === 'ph'
}

export function phOutOfRange(value: number): boolean {
  return value < 0 || value > 14
}

export function formatDate(iso: string): string {
  const date = new Date(iso)
  if (Number.isNaN(date.getTime())) {
    return iso
  }
  return new Intl.DateTimeFormat('pt-BR', { dateStyle: 'short', timeStyle: 'short' }).format(date)
}

export function toDateTimeLocal(iso: string): string {
  const date = new Date(iso)
  if (Number.isNaN(date.getTime())) {
    return ''
  }
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`
}

export function fromDateTimeLocal(value: string): string {
  return new Date(value).toISOString()
}

export function stationCaption(
  station: { name: string; code: string } | undefined,
  fallback = 'Sem estação',
): string {
  return station ? `${station.name} — ${station.code}` : fallback
}
