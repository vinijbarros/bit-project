import type { Category, FieldErrors, MetadataOption } from '../../types/api'

export interface RequestFormValues {
  title: string
  description: string
  category: string
}

export type RequestFormField = keyof RequestFormValues
export type RequestFormErrors = Partial<Record<RequestFormField, string>>

export const emptyRequestFormValues: RequestFormValues = {
  title: '',
  description: '',
  category: '',
}

export function unicodeLength(value: string): number {
  return Array.from(value).length
}

export function validateRequestForm(
  values: RequestFormValues,
  categories: MetadataOption<Category>[],
): { values: RequestFormValues; errors: RequestFormErrors } {
  const normalized = {
    title: values.title.trim(),
    description: values.description.trim(),
    category: values.category.trim(),
  }
  const errors: RequestFormErrors = {}
  const titleLength = unicodeLength(normalized.title)
  const descriptionLength = unicodeLength(normalized.description)

  if (titleLength < 3 || titleLength > 150) errors.title = 'Informe um título entre 3 e 150 caracteres.'
  if (descriptionLength < 10 || descriptionLength > 5000) errors.description = 'Informe uma descrição entre 10 e 5000 caracteres.'
  if (!categories.some((category) => category.value === normalized.category)) errors.category = 'Escolha uma categoria válida.'

  return { values: normalized, errors }
}

export function requestFormErrorsFromAPI(fields?: FieldErrors): RequestFormErrors {
  return {
    title: fields?.title?.[0],
    description: fields?.description?.[0],
    category: fields?.category?.[0],
  }
}

export function hasRequestFormErrors(errors: RequestFormErrors): boolean {
  return Object.values(errors).some(Boolean)
}
