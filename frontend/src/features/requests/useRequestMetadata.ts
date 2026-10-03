import { useCallback, useEffect, useState } from 'react'

import { ApiError } from '../../api/client'
import { getMetadata } from '../../services/metadata'
import type { Metadata } from '../../types/api'

export function useRequestMetadata() {
  const [metadata, setMetadata] = useState<Metadata | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string>()
  const [reloadKey, setReloadKey] = useState(0)

  useEffect(() => {
    const controller = new AbortController()
    let active = true
    setLoading(true)
    setError(undefined)
    void getMetadata(controller.signal)
      .then((value) => {
        if (active) setMetadata(value)
      })
      .catch((cause: unknown) => {
        if (!active || (cause instanceof DOMException && cause.name === 'AbortError')) return
        setMetadata(null)
        setError(cause instanceof ApiError ? cause.message : 'Não foi possível carregar as categorias.')
      })
      .finally(() => {
        if (active) setLoading(false)
      })
    return () => {
      active = false
      controller.abort()
    }
  }, [reloadKey])

  const retry = useCallback(() => setReloadKey((value) => value + 1), [])
  return { metadata, loading, error, retry }
}
