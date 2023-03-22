import {
  Dispatch,
  useCallback,
  useEffect,
  useState,
} from 'react'

import { ClientBindingFactory, Endpoint } from '../services/client'

type Error = any // TODO: figure out the type

interface QueryResponse<T> {
  data?: T
  error: Error // TODO: figure out the type
  loading: boolean
  refetch: Dispatch<void>
}

type OperationVariables = Record<string, any>;

interface QueryOptions<TVariables = OperationVariables> {
  variables?: TVariables
}

function useClientQuery<TResult = unknown, TVariables extends OperationVariables = OperationVariables>(endpoint: Endpoint,
  options?: QueryOptions<TVariables>): QueryResponse<TResult> {
  const binding = ClientBindingFactory<TResult>(endpoint)
  const [loading, setLoading] = useState(false)
  const [data, setData] = useState<TResult>()
  const [error, setError] = useState()

  const fetch = useCallback(() => {
    setLoading(true)
    setError(undefined)
    setData(undefined)

    binding(options?.variables ?? {}).then(res => setData(res))
      .catch(err => setError(err))
      .finally(() => setLoading(false))
  }, [binding])

  const refetch = useCallback(() => {
    if (loading) return

    fetch()
  }, [fetch, loading])

  useEffect(() => fetch(), [fetch])

  return {
    data,
    loading,
    error,
    refetch,
  } as QueryResponse<TResult>
}

interface UpdateReponse {
  error: Error
  loading: boolean
  update: Dispatch<void>
}

function useClientUpdate(endpoint: Endpoint): UpdateReponse {
  const binding = ClientBindingFactory(endpoint)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState()

  const update = useCallback(() => {
    setLoading(true)

    binding()
      .catch(err => setError(err))
      .finally(() => setLoading(false))
  }, [binding])

  return {
    error,
    loading,
    update,
  }
}

export { useClientQuery, useClientUpdate }
