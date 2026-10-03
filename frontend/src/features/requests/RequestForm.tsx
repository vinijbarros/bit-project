import type { FormEvent } from 'react'

import { FormField } from '../../components/forms/FormField'
import type { Category, MetadataOption } from '../../types/api'
import { unicodeLength, type RequestFormErrors, type RequestFormField, type RequestFormValues } from './requestFormValidation'

interface RequestFormProps {
  values: RequestFormValues
  errors: RequestFormErrors
  categories: MetadataOption<Category>[]
  submitting: boolean
  submitLabel: string
  onChange: (field: RequestFormField, value: string) => void
  onSubmit: (event: FormEvent<HTMLFormElement>) => void | Promise<void>
  onCancel: () => void
}

export function RequestForm({
  values,
  errors,
  categories,
  submitting,
  submitLabel,
  onChange,
  onSubmit,
  onCancel,
}: RequestFormProps) {
  return (
    <form className="request-form" onSubmit={(event) => void onSubmit(event)} noValidate>
      <FormField
        id="request-title"
        label="Título"
        hint={`${unicodeLength(values.title)} de 150 caracteres`}
        error={errors.title}
        required
      >
        <input
          name="title"
          type="text"
          value={values.title}
          onChange={(event) => onChange('title', event.target.value)}
          disabled={submitting}
          autoFocus
        />
      </FormField>
      <FormField
        id="request-description"
        label="Descrição"
        hint={`${unicodeLength(values.description)} de 5000 caracteres. Quebras de linha serão preservadas.`}
        error={errors.description}
        required
      >
        <textarea
          name="description"
          rows={9}
          value={values.description}
          onChange={(event) => onChange('description', event.target.value)}
          disabled={submitting}
        />
      </FormField>
      <FormField id="request-category" label="Categoria" error={errors.category} required>
        <select
          name="category"
          value={values.category}
          onChange={(event) => onChange('category', event.target.value)}
          disabled={submitting}
        >
          <option value="">Selecione uma categoria</option>
          {categories.map((category) => <option key={category.value} value={category.value}>{category.label}</option>)}
        </select>
      </FormField>
      <p className="automatic-fields-note">Solicitante, data de criação e status são definidos automaticamente pelo sistema.</p>
      <div className="form-actions">
        <button className="button" type="submit" disabled={submitting}>{submitting ? 'Salvando…' : submitLabel}</button>
        <button className="button button--quiet" type="button" onClick={onCancel} disabled={submitting}>Cancelar</button>
      </div>
    </form>
  )
}
