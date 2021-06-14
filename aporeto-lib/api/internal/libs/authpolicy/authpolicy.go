package authpolicy

import (
	"context"
	"fmt"

	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/constants"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/internal/utils"
	"go.aporeto.io/elemental"
	"go.aporeto.io/gaia"
	"go.aporeto.io/manipulate"
)

// Create creates a read only authorization policy for a tenant with claims specified as oidcClaims
// oidcClaims is a 2d string array. For user to be allowed to access, at least one array claims must be satisfied.
func Create(ctx context.Context, m manipulate.Manipulator, namespace, name, description string, oidcClaims [][]string) error {

	// Ensure namespaces are correctly formatted.
	namespace = utils.SetupNamespaceString(namespace)

	// Setup a new authorization policy.
	ap := gaia.NewAPIAuthorizationPolicy()
	ap.Name = name
	ap.Description = description
	ap.Subject = oidcClaims
	ap.AuthorizedIdentities = []string{constants.AuthNamespaceViewer}
	ap.AuthorizedNamespace = namespace
	ap.PropagationHidden = true
	ap.Metadata = utils.MakeTenantMetadata(namespace)

	// Create a sub context so we dont retry too long.
	subctx, cancel := context.WithTimeout(ctx, constants.APIDefaultContextTimeout)
	defer cancel()

	// Create a namespace context where we are creating an object.
	mctx := manipulate.NewContext(
		subctx,
		manipulate.ContextOptionNamespace(namespace),
	)

	// Try creating multiple times in case of connection errors.
	return m.Create(mctx, ap)
}

// Delete deletes an authorization policy.
func Delete(ctx context.Context, m manipulate.Manipulator, namespace, name string) error {

	// Ensure namespaces are correctly formatted.
	namespace = utils.SetupNamespaceString(namespace)

	// Get matching authorization policies.
	aplist, err := Get(ctx, m, namespace, name)
	if err != nil {
		return err
	}

	// Create a sub context so we dont retry too long.
	subctx, cancel := context.WithTimeout(ctx, constants.APIDefaultContextTimeout)
	defer cancel()

	// Create a namespace context where we are creating an object.
	mctx := manipulate.NewContext(
		subctx,
		manipulate.ContextOptionNamespace(utils.SetupNamespaceString(namespace)),
	)

	var ret error
	for _, ap := range aplist {
		// Try deleting multiple times in case of connection errors.
		err := m.Delete(mctx, ap)
		if err != nil {
			ret = err
		}
	}

	return ret
}

// Get fetches a list of authorization policies matching the criteria.
func Get(ctx context.Context, m manipulate.Manipulator, namespace, name string) (gaia.APIAuthorizationPoliciesList, error) {

	// Ensure namespaces are correctly formatted.
	namespace = utils.SetupNamespaceString(namespace)

	aplist := gaia.APIAuthorizationPoliciesList{}

	// Create a sub context so we dont retry too long.
	subctx, cancel := context.WithTimeout(ctx, constants.APIDefaultContextTimeout)
	defer cancel()

	mctx := manipulate.NewContext(
		subctx,
		manipulate.ContextOptionNamespace(namespace),
		manipulate.ContextOptionFilter(
			elemental.NewFilterComposer().
				WithKey("name").Equals(name).
				Done(),
		),
	)

	if err := m.RetrieveMany(mctx, &aplist); err != nil {
		return nil, err
	}
	if len(aplist) == 0 {
		return nil, fmt.Errorf("no authorization policies found in namespace '%s'", namespace)
	}

	return aplist, nil
}
