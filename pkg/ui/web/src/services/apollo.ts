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

export function newApolloClient(token: string): ApolloClient<NormalizedCacheObject> {
  const authLink = setContext(() => ({ headers: token ? { authorization: `Bearer ${token}` } : {} }))
  const httpLink = new HttpLink({ uri: GQL_URL })
  const absintheSocket = create(new PhoenixSocket(WS_URI, { params: () => (token ? { Authorization: `Bearer ${token}` } : {}) }))
  const socketLink = createAbsintheSocketLink(absintheSocket)
  const retryLink = new RetryLink({
    delay: { initial: 200 },
    attempts: {
      max: Infinity,
      retryIf: error => !!error,
    },
  })

  return new ApolloClient<NormalizedCacheObject>({
    link: ApolloLink.from([authLink, httpLink]),
    cache: new InMemoryCache(),
  })
}
