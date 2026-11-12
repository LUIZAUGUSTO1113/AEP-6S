export const BRAZIL = { lat: -14.2, lng: -51.9, zoom: 4 }

export type UfView = {
  name: string
  lat: number
  lng: number
  zoom: number
}

export const UFS: Record<string, UfView> = {
  AC: { name: 'Acre', lat: -8.77, lng: -70.55, zoom: 7 },
  AL: { name: 'Alagoas', lat: -9.57, lng: -36.78, zoom: 8 },
  AP: { name: 'Amapá', lat: 1.41, lng: -51.77, zoom: 7 },
  AM: { name: 'Amazonas', lat: -3.47, lng: -65.1, zoom: 6 },
  BA: { name: 'Bahia', lat: -12.97, lng: -41.57, zoom: 6 },
  CE: { name: 'Ceará', lat: -5.2, lng: -39.53, zoom: 7 },
  DF: { name: 'Distrito Federal', lat: -15.83, lng: -47.86, zoom: 10 },
  ES: { name: 'Espírito Santo', lat: -19.19, lng: -40.34, zoom: 8 },
  GO: { name: 'Goiás', lat: -15.98, lng: -49.86, zoom: 7 },
  MA: { name: 'Maranhão', lat: -5.42, lng: -45.44, zoom: 7 },
  MT: { name: 'Mato Grosso', lat: -12.64, lng: -55.42, zoom: 6 },
  MS: { name: 'Mato Grosso do Sul', lat: -20.51, lng: -54.54, zoom: 7 },
  MG: { name: 'Minas Gerais', lat: -18.1, lng: -44.38, zoom: 6 },
  PA: { name: 'Pará', lat: -3.79, lng: -52.48, zoom: 6 },
  PB: { name: 'Paraíba', lat: -7.24, lng: -36.78, zoom: 8 },
  PR: { name: 'Paraná', lat: -24.89, lng: -51.55, zoom: 7 },
  PE: { name: 'Pernambuco', lat: -8.38, lng: -37.86, zoom: 7 },
  PI: { name: 'Piauí', lat: -7.72, lng: -42.73, zoom: 7 },
  RJ: { name: 'Rio de Janeiro', lat: -22.25, lng: -42.66, zoom: 8 },
  RN: { name: 'Rio Grande do Norte', lat: -5.81, lng: -36.59, zoom: 8 },
  RS: { name: 'Rio Grande do Sul', lat: -30.17, lng: -53.5, zoom: 7 },
  RO: { name: 'Rondônia', lat: -10.83, lng: -63.34, zoom: 7 },
  RR: { name: 'Roraima', lat: 2.09, lng: -61.33, zoom: 7 },
  SC: { name: 'Santa Catarina', lat: -27.45, lng: -50.95, zoom: 8 },
  SP: { name: 'São Paulo', lat: -22.19, lng: -48.79, zoom: 7 },
  SE: { name: 'Sergipe', lat: -10.57, lng: -37.45, zoom: 9 },
  TO: { name: 'Tocantins', lat: -10.18, lng: -48.33, zoom: 7 },
}

export function ufView(state: string): UfView | undefined {
  return UFS[state.trim().toUpperCase()]
}
