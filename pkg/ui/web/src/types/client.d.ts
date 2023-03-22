import { Endpoint } from '../services/client'

type ClientBinding<TResult = unknown> = (...args: any) => Promise<TResult>
type Client = {[key in Endpoint]: ClientBinding}

interface ListRepositoriesVariables {
  query: string
}

interface ListRecipesVariables {
  repo: string
  provider: string // TODO: use provider enum
}

interface Repository {
  Id: string
  Name: string
  Description: string
  Notes: string
  Icon: string
  DarkIcon: string
  Recipes: Array<Recipe>
}

interface Recipe {
  Id: string
  Name: string
  Provider: string
  Description: string
  Restricted: boolean
}

export type {
  Client, ClientBinding, Repository, Recipe, ListRecipesVariables, ListRepositoriesVariables,
}
