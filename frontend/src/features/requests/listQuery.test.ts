import { describe, expect, it } from 'vitest'

import {
  draftFromSearchParams,
  filtersFromSearchParams,
  searchParamsFromDraft,
  validateRequestPeriod,
} from './listQuery'

describe('estado da consulta de solicitações', () => {
  it('converte filtros e paginação entre formulário e URL sem alterar datas civis', () => {
    const params = searchParamsFromDraft({
      dateFrom: '2026-10-01',
      dateTo: '2026-10-02',
      category: 'ti',
      status: 'aberto',
      query: '  acesso  ',
    }, 3)

    expect(params.toString()).toBe('date_from=2026-10-01&date_to=2026-10-02&category=ti&status=aberto&q=acesso&page=3')
    expect(draftFromSearchParams(params)).toEqual({
      dateFrom: '2026-10-01', dateTo: '2026-10-02', category: 'ti', status: 'aberto', query: 'acesso',
    })
    expect(filtersFromSearchParams(params)).toEqual({
      date_from: '2026-10-01', date_to: '2026-10-02', category: 'ti', status: 'aberto', q: 'acesso', page: 3, page_size: 20,
    })
  })

  it('rejeita período invertido e aceita limites ausentes ou iguais', () => {
    expect(validateRequestPeriod({ dateFrom: '2026-10-03', dateTo: '2026-10-02', category: '', status: '', query: '' })).toMatch(/posterior/)
    expect(validateRequestPeriod({ dateFrom: '2026-10-02', dateTo: '2026-10-02', category: '', status: '', query: '' })).toBeUndefined()
    expect(validateRequestPeriod({ dateFrom: '', dateTo: '2026-10-02', category: '', status: '', query: '' })).toBeUndefined()
  })
})
