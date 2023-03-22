import { ListRecipes, ListRepositories } from '../../wailsjs/go/ui/Client'
import {
  Client,
  ClientBinding,
  Recipe,
  Repository,
} from '../types/client'

/**
 * List of supported client methods based on API Go client.
 * @see pkg/api/client.go
 */
enum Endpoint {
  ListRepositories = 'ListRepositories',
  ListRecipes = 'ListRecipes',
}

/**
 * Client mapping from API "endpoint" to the proper backend binding.
 * Abstracts the backend calls and wraps them with proper return types
 * to simplify usage in the UI.
 * @see Endpoint
 */
const Plural: Client = {
  [Endpoint.ListRepositories]: ({ query }): Promise<Array<Repository>> => ListRepositories(query),
  [Endpoint.ListRecipes]: ({ repo, provider }): Promise<Array<Recipe>> => ListRecipes(repo, provider),
}

/**
 * Factory that simplifies getting wrapped client binding methods.
 * @param endpoint
 * @constructor
 */
function ClientBindingFactory<TResult>(endpoint: Endpoint): ClientBinding<TResult> {
  const binding: ClientBinding<TResult> = Plural[endpoint] as ClientBinding<TResult>

  if (!binding) throw new Error(`Unsupported client endpoint: ${endpoint}`)

  return binding
}

export { ClientBindingFactory, Endpoint }
