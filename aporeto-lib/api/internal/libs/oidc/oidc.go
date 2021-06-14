package oidc

import (
	"context"
	"fmt"

	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/constants"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/internal/utils"
	"go.aporeto.io/elemental"
	"go.aporeto.io/gaia"
	"go.aporeto.io/manipulate"
)

// Create OIDC provider by configuring all the required parameters.
// ctx, m, tenantNamespace, o.Name, o.Endpoint, o.ClientID, o.ClientSecret, o.Scopes, o.Default, o.Subjects
func Create(ctx context.Context, m manipulate.Manipulator, namespace, name, endpoint, clientid, clientsecret string, defaultflag bool, scopes, subjects []string) error {

	// Ensure namespaces are correctly formatted.
	namespace = utils.SetupNamespaceString(namespace)

	// Setup a new OIDC provider.
	op := gaia.NewOIDCProvider()
	op.Name = name
	op.Endpoint = endpoint
	op.ClientID = clientid
	op.ClientSecret = clientsecret
	op.Default = defaultflag
	op.Scopes = scopes
	op.Subjects = subjects

	// Create a sub context so we dont retry too long.
	subctx, cancel := context.WithTimeout(ctx, constants.APIDefaultContextTimeout)
	defer cancel()

	// Create a namespace context where we are creating an object.
	mctx := manipulate.NewContext(
		subctx,
		manipulate.ContextOptionNamespace(namespace),
	)

	// Try creating multiple times in case of connection errors.
	return m.Create(mctx, op)
}

// Delete deletes an OIDC provider configuration.
func Delete(ctx context.Context, m manipulate.Manipulator, namespace, name string) error {

	// Ensure namespaces are correctly formatted.
	namespace = utils.SetupNamespaceString(namespace)

	// Get matching OIDC provider.
	oidcproviderlist, err := Get(ctx, m, namespace, name)
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
	for _, op := range oidcproviderlist {
		// Try deleting multiple times in case of connection errors.
		err := m.Delete(mctx, op)
		if err != nil {
			ret = err
		}
	}

	return ret
}

// Get fetches a list of OIDC provider config matching the name criteria.
func Get(ctx context.Context, m manipulate.Manipulator, namespace, name string) (gaia.OIDCProvidersList, error) {

	// Ensure namespaces are correctly formatted.
	namespace = utils.SetupNamespaceString(namespace)

	oidcproviderlist := gaia.OIDCProvidersList{}

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

	if err := m.RetrieveMany(mctx, &oidcproviderlist); err != nil {
		return nil, err
	}

	if len(oidcproviderlist) == 0 {
		return nil, fmt.Errorf("no OIDC providers found in namespace '%s'", namespace)
	}

	return oidcproviderlist, nil
}
