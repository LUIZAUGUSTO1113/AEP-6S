const API_UNAVAILABLE = 'A API não respondeu como esperado.'

const translations: Record<string, string> = {
  'Invalid JSON.': 'JSON inválido.',
  'invalid station id format': 'Identificador da estação inválido.',
  'invalid sample id format': 'Identificador da coleta inválido.',
  'invalid object id format': 'Identificador inválido.',
  'station not found': 'Estação não encontrada.',
  'sample not found': 'Coleta não encontrada.',
  'mongo: no documents in result': 'Registro não encontrado.',
  'station still has samples and cannot be deleted':
    'A estação ainda tem coletas e não pode ser excluída.',
  'a station with this code already exists': 'Já existe uma estação com este código.',
  'Name is required.': 'Informe o nome da estação.',
  'River is required.': 'Informe o rio.',
  'Code is required.': 'Informe o código.',
  'Location city is required.': 'Informe a cidade.',
  'Location state is required.': 'Informe a UF.',
  'Threshold parameter is required.': 'Informe o parâmetro do limiar.',
  'Threshold min must be less than or equal to max.': 'O mínimo está acima do máximo.',
  'pH threshold must be between 0 and 14 (inclusive).': 'O pH precisa ficar entre 0 e 14.',
  'Value is required.': 'Informe o valor.',
  'Parameter is required.': 'Informe o parâmetro.',
  'pH must be between 0 and 14 (inclusive).': 'O pH precisa ficar entre 0 e 14.',
}

export function translateApiError(message: string): string {
  if (message === API_UNAVAILABLE) {
    return message
  }

  const exact = translations[message]
  if (exact) {
    return exact
  }

  const threshold = /^thresholds\[(\d+)\]: (.+)$/.exec(message)
  if (threshold) {
    const inner = translations[threshold[2]] ?? translateApiError(threshold[2])
    return `Limiar ${Number(threshold[1]) + 1}: ${inner}`
  }

  if (/[áàâãéêíóôõúçÁÀÂÃÉÊÍÓÔÕÚÇ]/.test(message) || message.startsWith('Informe ')) {
    return message
  }

  return 'Não foi possível concluir a operação.'
}

export { API_UNAVAILABLE }
