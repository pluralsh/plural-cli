import {
  Dispatch,
  useCallback,
  useEffect,
  useState,
} from 'react'

import { Binding, ClientBindingFactory } from '../services/wails'

type Error = any // TODO: figure out the type

interface QueryResponse<T> {
  data?: T
  error: Error
  loading: boolean
  refetch: Dispatch<void>
}

type OperationVariables = Record<string, any>;

interface QueryOptions<TVariables = OperationVariables> {
  variables?: TVariables
}

function useWailsQuery<TResult = unknown, TVariables extends OperationVariables = OperationVariables>(binding: Binding,
  options?: QueryOptions<TVariables>): QueryResponse<TResult> {
  const bindingFn = ClientBindingFactory<TResult>(binding)
  const [loading, setLoading] = useState(false)
  const [data, setData] = useState<TResult>()
  const [error, setError] = useState()

  const fetch = useCallback(() => {
    setLoading(true)
    setError(undefined)
    setData(undefined)

    bindingFn(options?.variables ?? {}).then(res => setData(res))
      .catch(err => setError(err))
      .finally(() => setLoading(false))
  }, [bindingFn, options?.variables])

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

interface UpdateReponse<T> {
  error: Error
  loading: boolean
  update: Dispatch<void>
  data?: T
}

function useWailsUpdate<TResult = void, TVariables extends OperationVariables = OperationVariables>(binding: Binding,
  options?: QueryOptions<TVariables>): UpdateReponse<TResult> {
  const bindingFn = ClientBindingFactory<TResult>(binding)
  const [loading, setLoading] = useState(false)
  const [data, setData] = useState<TResult>()
  const [error, setError] = useState()

  const update = useCallback(() => {
    setLoading(true)

    bindingFn(options?.variables ?? {})
      .then(data => setData(data))
      .catch(err => setError(err))
      .finally(() => setLoading(false))
  }, [bindingFn, options?.variables])

  return {
    error,
    loading,
    update,
    data,
  }
}

export { useWailsQuery, useWailsUpdate }
