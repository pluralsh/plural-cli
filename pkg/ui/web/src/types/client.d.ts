import { Provider } from '../graphql/generated/graphql'
import { Binding } from '../services/client'

type ClientBinding<TResult = unknown> = (...args: any) => Promise<TResult>
type Client = {[key in Binding]: ClientBinding}

interface NetworkConfig {
  subdomain: string
  pluralDns: bool
}

interface PluralProject {
  cluster: string
  bucket: string
  project: string
  provider: Provider
  region: string
  bucketPrefix: string
  network: NetworkConfig
  context: Map<string, unknown>
}

interface PluralContext {
  buckets: Array<string>
  domains: Array<string>
  configuration: Record<string, Record<string, unknown>>
}

export type {
  Client, ClientBinding, PluralProject, PluralContext,
}
