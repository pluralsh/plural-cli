import { GraphQLErrors } from '@apollo/client/errors'
import { useMemo, useState } from 'react'
import { create } from '@absinthe/socket'
import { createAbsintheSocketLink } from '@absinthe/socket-apollo-link'
import {
  ApolloClient,
  ApolloLink,
  HttpLink,
  InMemoryCache,
  NormalizedCacheObject,
} from '@apollo/client'
import { setContext } from '@apollo/client/link/context'
import { onError } from '@apollo/client/link/error'
import { RetryLink } from '@apollo/client/link/retry'
import { Socket as PhoenixSocket } from 'phoenix'

const API_HOST = 'app.plural.sh'
const GQL_URL = `https://${API_HOST}/gql`
const WS_URI = `wss://${API_HOST}/socket`

// const splitLink = split(({ query }) => {
//   const definition = getMainDefinition(query)
//
//   return (
//     definition.kind === 'OperationDefinition'
//       && definition.operation === 'subscription'
//   )
// },
// socketLink,
// retryLink.concat(resetToken).concat(httpLink),)

interface ApolloClientHook {
  client: ApolloClient<NormalizedCacheObject>
  error: GraphQLErrors | undefined
}

export function useApolloClient(token: string): ApolloClientHook {
  const [error, setError] = useState<GraphQLErrors>()

  const authLink = setContext(() => ({ headers: token ? { authorization: `Bearer ${token}` } : {} }))
  const httpLink = useMemo(() => new HttpLink({ uri: GQL_URL }), [])
  const absintheSocket = create(new PhoenixSocket(WS_URI, { params: () => (token ? { Authorization: `Bearer ${token}` } : {}) }))
  const errorLink = onError(({ graphQLErrors }) => {
    if (!error && graphQLErrors) {
      const clearDelay = 15000

      setError(graphQLErrors)
      setTimeout(() => setError(undefined), clearDelay)
    }
  })
  const _socketLink = createAbsintheSocketLink(absintheSocket)
  const _retryLink = new RetryLink({
    delay: { initial: 200 },
    attempts: {
      max: Infinity,
      retryIf: error => !!error,
    },
  })

  return useMemo(() => ({
    client: new ApolloClient<NormalizedCacheObject>({
      link: ApolloLink.from([authLink, errorLink, httpLink]),
      cache: new InMemoryCache(),
    }),
    error,
  }), [authLink, error, httpLink, errorLink])
}
