// Package graphql holds the gqlgen-generated GraphQL surface and
// the resolver wiring that translates between transport (request
// context, claims) and the conductor.
//
// Resolvers extract auth + workspace context up-front and pass typed
// values into the conductor. No package below this layer reads from
// context for tenant or identity.
//
// Populated in phase 3 from services/gateway/graphql/.
package graphql
