import { useState, type FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'

import { ApiError } from '../api/client'
import { AlertMessage } from '../components/feedback/AlertMessage'
import { ErrorState } from '../components/feedback/ErrorState'
import { LoadingState } from '../components/feedback/LoadingState'
import { PageHeader } from '../components/layout/PageHeader'
import { RequestForm } from '../features/requests/RequestForm'
import {
  emptyRequestFormValues,
  hasRequestFormErrors,
  requestFormErrorsFromAPI,
  validateRequestForm,
  type RequestFormErrors,
  type RequestFormField,
  type RequestFormValues,
} from '../features/requests/requestFormValidation'
import { useRequestMetadata } from '../features/requests/useRequestMetadata'
import { useUnsavedChangesWarning } from '../features/requests/useUnsavedChangesWarning'
import { requestsService } from '../services/requests'
import type { Category } from '../types/api'

export function RequestCreatePage() {
  const navigate = useNavigate()
  const { metadata, loading, error: metadataError, retry } = useRequestMetadata()
  const [values, setValues] = useState<RequestFormValues>(emptyRequestFormValues)
  const [fieldErrors, setFieldErrors] = useState<RequestFormErrors>({})
  const [submitError, setSubmitError] = useState<string>()
  const [submitting, setSubmitting] = useState(false)
  const dirty = Object.values(values).some((value) => value.length > 0)
  useUnsavedChangesWarning(dirty && !submitting)

  function change(field: RequestFormField, value: string) {
    setValues((current) => ({ ...current, [field]: value }))
    setFieldErrors((current) => ({ ...current, [field]: undefined }))
    setSubmitError(undefined)
  }

  function cancel() {
    if (dirty && !window.confirm('Descartar as alterações desta solicitação?')) return
    navigate('/solicitacoes')
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (submitting || !metadata) return
    const validation = validateRequestForm(values, metadata.categories)
    setFieldErrors(validation.errors)
    setSubmitError(undefined)
    if (hasRequestFormErrors(validation.errors)) return

    setSubmitting(true)
    try {
      const created = await requestsService.create({
        title: validation.values.title,
        description: validation.values.description,
        category: validation.values.category as Category,
      })
      navigate(`/solicitacoes/${created.id}`, {
        replace: true,
        state: { message: 'Solicitação criada com sucesso.', from: '/solicitacoes' },
      })
    } catch (cause) {
      if (cause instanceof ApiError) {
        setSubmitError(cause.message)
        setFieldErrors(requestFormErrorsFromAPI(cause.fields))
      } else {
        setSubmitError('Não foi possível criar a solicitação. Verifique sua conexão e tente novamente.')
      }
    } finally {
      setSubmitting(false)
    }
  }

  if (loading) return <LoadingState label="Carregando formulário…" />
  if (metadataError || !metadata) return <ErrorState title="Não foi possível abrir o formulário" message={metadataError ?? 'Metadata indisponível.'} onRetry={retry} />

  return (
    <section aria-labelledby="create-request-title">
      <PageHeader eyebrow="Solicitações" title="Nova solicitação" titleId="create-request-title" description="Informe o que precisa ser atendido pela equipe responsável." />
      {submitError && <AlertMessage tone="error">{submitError}</AlertMessage>}
      <RequestForm
        values={values}
        errors={fieldErrors}
        categories={metadata.categories}
        submitting={submitting}
        submitLabel="Criar solicitação"
        onChange={change}
        onSubmit={submit}
        onCancel={cancel}
      />
    </section>
  )
}
