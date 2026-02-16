package authz

import (
	"fmt"

	"cloudplane/credential-broker/internal/oidc"
)

// Authorizer handles authorization checks
type Authorizer struct{}

// NewAuthorizer creates a new authorizer
func NewAuthorizer() *Authorizer {
	return &Authorizer{}
}

// Authorize checks if the claims allow credential issuance
func (a *Authorizer) Authorize(claims *oidc.Claims) error {
	// TODO: Implement authorization logic
	//
	// Implementation should:
	// 1. Check claims.Scopes for required scope (e.g., "issue:credentials")
	// 2. Validate claims.Subject matches expected pattern
	// 3. Return error if not authorized
	//
	// Example check:
	// if !slices.Contains(claims.Scopes, "issue:credentials") {
	//     return fmt.Errorf("missing required scope")
	// }

	if claims == nil {
		return fmt.Errorf("no claims provided")
	}

	return fmt.Errorf("not implemented")
}

// AuthorizeForRole checks if the claims allow assuming a specific role
func (a *Authorizer) AuthorizeForRole(claims *oidc.Claims, roleARN string) error {
	// TODO: Implement role-specific authorization
	//
	// Implementation should:
	// 1. Call a.Authorize(claims) for base checks
	// 2. Check if claims.Subject is allowed to assume roleARN
	// 3. Could query database for project → role mappings
	//
	// For MVP, trust IAM trust policy (STS will reject if not allowed)

	_ = roleARN
	return a.Authorize(claims)
}
