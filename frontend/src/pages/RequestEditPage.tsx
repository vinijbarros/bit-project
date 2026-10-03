import { useEffect, useState, type FormEvent } from 'react'
import { Link, useLocation, useNavigate, useParams } from 'react-router-dom'

import { ApiError } from '../api/client'
import { AlertMessage } from '../components/feedback/AlertMessage'
import { ErrorState } from '../components/feedback/ErrorState'
import { LoadingState } from '../components/feedback/LoadingState'
import { PageHeader } from '../components/layout/PageHeader'
import { RequestForm } from '../features/requests/RequestForm'
import {
  hasRequestFormErrors,
  requestFormErrorsFromAPI,
  validateRequestForm,
  type RequestFormErrors,
  type RequestFormField,
  type RequestFormValues,
} from '../features/requests/requestForm'
import { useRequestMetadata } from '../features/requests/useRequestMetadata'
import { useUnsavedChangesWarning } from '../features/requests/useUnsavedChangesWarning'
import { requestsService } from '../services/requests'
import type { Category, Request, UpdateRequestInput } from '../types/api'

function listReturnPath(value: unknown): string {
  if (typeof value === 'string' && (value === '/solicitacoes' || value.startsWith('/solicitacoes?'))) return value
  return '/solicitacoes'
}

export function RequestEditPage() {
  const { id } = useParams()
  const navigate = useNavigate()
  const location = useLocation()
  const numericID = Number(id)
  const validID = Number.isSafeInteger(numericID) && numericID > 0
  const { metadata, loading: loadingMetadata, error: metadataError, retry: retryMetadata } = useRequestMetadata()
  const [request, setRequest] = useState<Request | null>(null)
  const [values, setValues] = useState<RequestFormValues>({ title: '', description: '', category: '' })
  const [fieldErrors, setFieldErrors] = useState<RequestFormErrors>({})
  const [loadError, setLoadError] = useState<string>()
  const [submitError, setSubmitError] = useState<string>()
  const [conflict, setConflict] = useState(false)
  const [loadingRequest, setLoadingRequest] = useState(validID)
  const [submitting, setSubmitting] = useState(false)
  const [reloadKey, setReloadKey] = useState(0)
  const state = (location.state ?? {}) as { from?: unknown }
  const returnPath = listReturnPath(state.from)

  useEffect(() => {
    if (!validID) return
    const controller = new AbortController()
    let active = true
    setLoadingRequest(true)
    setLoadError(undefined)
    void requestsService.get(numericID, controller.signal)
      .then((value) => {
        if (!active) return
        setRequest(value)
        setValues({ title: value.title, description: value.description, category: value.category.value })
      })
      .catch((cause: unknown) => {
        if (!active || (cause instanceof DOMException && cause.name === 'AbortError')) return
        setRequest(null)
        if (cause instanceof ApiError && cause.status === 404) setLoadError('A solicitação informada não existe ou foi excluída.')
        else if (cause instanceof ApiError) setLoadError(cause.message)
        else setLoadError('Não foi possível carregar a solicitação.')
      })
      .finally(() => {
        if (active) setLoadingRequest(false)
      })
    return () => {
      active = false
      controller.abort()
    }
  }, [numericID, reloadKey, validID])

  const dirty = request
    ? values.title !== request.title || values.description !== request.description || values.category !== request.category.value
    : false
  useUnsavedChangesWarning(dirty && !submitting)

  function change(field: RequestFormField, value: string) {
    setValues((current) => ({ ...current, [field]: value }))
    setFieldErrors((current) => ({ ...current, [field]: undefined }))
    setSubmitError(undefined)
    setConflict(false)
  }

  function cancel() {
    if (dirty && !window.confirm('Descartar as alterações desta solicitação?')) return
    navigate(`/solicitacoes/${numericID}`, { state: { from: returnPath } })
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (submitting || !metadata || !request) return
    const validation = validateRequestForm(values, metadata.categories)
    setFieldErrors(validation.errors)
    setSubmitError(undefined)
    setConflict(false)
    if (hasRequestFormErrors(validation.errors)) return

    const input: UpdateRequestInput = {}
    if (validation.values.title !== request.title) input.title = validation.values.title
    if (validation.values.description !== request.description) input.description = validation.values.description
    if (validation.values.category !== request.category.value) input.category = validation.values.category as Category
    if (Object.keys(input).length === 0) {
      setSubmitError('Faça ao menos uma alteração antes de salvar.')
      return
    }

    setSubmitting(true)
    try {
      const updated = await requestsService.update(numericID, input)
      navigate(`/solicitacoes/${updated.id}`, {
        replace: true,
        state: { message: 'Solicitação atualizada com sucesso.', from: returnPath },
      })
    } catch (cause) {
      if (cause instanceof ApiError) {
        setFieldErrors(requestFormErrorsFromAPI(cause.fields))
        if (cause.status === 409) {
          setConflict(true)
          setSubmitError('A solicitação mudou de status e não pode mais ser editada. Recarregue os detalhes para conferir o estado atual.')
        } else if (cause.status === 403) {
          setSubmitError('Você não tem permissão para editar esta solicitação.')
        } else if (cause.status === 404) {
          setSubmitError('A solicitação não existe mais.')
        } else {
          setSubmitError(cause.message)
        }
      } else {
        setSubmitError('Não foi possível salvar. Verifique sua conexão e tente novamente.')
      }
    } finally {
      setSubmitting(false)
    }
  }

  if (!validID) return <ErrorState title="Solicitação inválida" message="O identificador informado na URL não é válido." />
  if (loadingRequest || loadingMetadata) return <LoadingState label="Carregando solicitação para edição…" />
  if (loadError || !request) return <ErrorState title="Não foi possível editar a solicitação" message={loadError ?? 'Resposta inválida da API.'} onRetry={() => setReloadKey((value) => value + 1)} />
  if (metadataError || !metadata) return <ErrorState title="Não foi possível abrir o formulário" message={metadataError ?? 'Metadata indisponível.'} onRetry={retryMetadata} />

  if (!request.permissions.can_edit) {
    const message = request.status.value !== 'aberto'
      ? 'Somente solicitações com status Aberto podem ser editadas.'
      : 'Somente o autor pode editar esta solicitação.'
    return (
      <ErrorState
        title="Edição não permitida"
        message={message}
        onRetry={() => navigate(`/solicitacoes/${numericID}`, { state: { from: returnPath } })}
        retryLabel="Voltar aos detalhes"
      />
    )
  }

  return (
    <section aria-labelledby="edit-request-title">
      <PageHeader eyebrow={request.code} title="Editar solicitação" titleId="edit-request-title" description="Título, descrição e categoria podem ser alterados enquanto a solicitação estiver aberta." />
      {submitError && (
        <AlertMessage tone="error">
          {submitError} {conflict && <Link to={`/solicitacoes/${numericID}`} state={{ from: returnPath }}>Recarregar detalhes</Link>}
        </AlertMessage>
      )}
      <RequestForm
        values={values}
        errors={fieldErrors}
        categories={metadata.categories}
        submitting={submitting}
        submitLabel="Salvar alterações"
        onChange={change}
        onSubmit={submit}
        onCancel={cancel}
      />
    </section>
  )
}
