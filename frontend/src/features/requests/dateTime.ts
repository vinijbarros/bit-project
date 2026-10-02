const saoPauloDateTime = new Intl.DateTimeFormat('pt-BR', {
  timeZone: 'America/Sao_Paulo',
  dateStyle: 'short',
  timeStyle: 'short',
})

export function formatDateTimeSaoPaulo(value: string): string {
  return saoPauloDateTime.format(new Date(value))
}
