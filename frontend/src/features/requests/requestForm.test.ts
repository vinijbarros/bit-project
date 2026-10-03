import { describe, expect, it } from 'vitest'

import type { MetadataOption, Category } from '../../types/api'
import { unicodeLength, validateRequestForm } from './requestForm'

const categories: MetadataOption<Category>[] = [
  { value: 'ti', label: 'TI' },
  { value: 'rh', label: 'RH' },
  { value: 'compras', label: 'Compras' },
  { value: 'financeiro', label: 'Financeiro' },
  { value: 'infraestrutura', label: 'Infraestrutura' },
]

describe('validação compartilhada do formulário de solicitação', () => {
  it('normaliza espaços externos e conta pontos de código Unicode como o backend', () => {
    expect(unicodeLength('á😀')).toBe(2)
    const result = validateRequestForm({
      title: '  Acesso 😀  ',
      description: '  123456789😀  ',
      category: 'ti',
    }, categories)

    expect(result.values).toEqual({ title: 'Acesso 😀', description: '123456789😀', category: 'ti' })
    expect(result.errors).toEqual({})
  })

  it('aceita limites exatos e rejeita vazio, excesso e categoria desconhecida', () => {
    expect(validateRequestForm({
      title: '😀'.repeat(150), description: 'á'.repeat(5000), category: 'infraestrutura',
    }, categories).errors).toEqual({})

    const errors = validateRequestForm({
      title: '😀'.repeat(151), description: 'curta', category: 'juridico',
    }, categories).errors
    expect(errors.title).toMatch(/3 e 150/)
    expect(errors.description).toMatch(/10 e 5000/)
    expect(errors.category).toMatch(/categoria válida/)

    const empty = validateRequestForm({ title: '  ', description: '\n ', category: '' }, categories).errors
    expect(empty.title).toBeTruthy()
    expect(empty.description).toBeTruthy()
    expect(empty.category).toBeTruthy()
  })
})
