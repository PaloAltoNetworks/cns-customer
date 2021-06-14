package enforcerprofile

import (
	"context"
	"fmt"

	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/constants"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/internal/utils"
	"go.aporeto.io/elemental"
	"go.aporeto.io/gaia"
	"go.aporeto.io/manipulate"
)

// Create creates an enforcer profile.
// Params:
//   - namespace: namespace where profile will be created.
//   - subNamespace: name of child namespace.
//   - description: description of profile.
func Create(
	ctx context.Context,
	m manipulate.Manipulator,
	namespace, name, description string,
) error {

	// Ensure namespaces are correctly formatted.
	namespace = utils.SetupNamespaceString(namespace)

	// Setup a new enforcer profile.
	ep := gaia.NewEnforcerProfile()
	ep.Name = name
	ep.Description = description
	ep.AssociatedTags = utils.MakeNamespaceAssociatedTags(namespace)
	ep.Metadata = utils.MakeTenantMetadata(namespace)

	// Create a sub context so we dont retry too long.
	subctx, cancel := context.WithTimeout(ctx, constants.APIDefaultContextTimeout)
	defer cancel()

	// Create a namespace context where we are creating an object.
	mctx := manipulate.NewContext(
		subctx,
		manipulate.ContextOptionNamespace(namespace),
	)

	// Try creating multiple times in case of connection errors.
	return m.Create(mctx, ep)
}

// Delete deletes an enforcer profile.
func Delete(ctx context.Context, m manipulate.Manipulator, namespace, name string) error {

	// Ensure namespaces are correctly formatted.
	namespace = utils.SetupNamespaceString(namespace)

	// Get matching enforcer profiles.
	np, err := Get(ctx, m, namespace, name)
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
	return m.Delete(mctx, np)
}

// Get fetches a list of enforcer profiles matching the criteria.
func Get(ctx context.Context, m manipulate.Manipulator, namespace, name string) (*gaia.EnforcerProfile, error) {

	eps := gaia.EnforcerProfilesList{}

	subctx, cancel := context.WithTimeout(ctx, constants.APIDefaultContextTimeout)
	defer cancel()

	mctx := manipulate.NewContext(
		subctx,
		manipulate.ContextOptionNamespace(utils.SetupNamespaceString(namespace)),
		manipulate.ContextOptionFilter(
			elemental.NewFilterComposer().
				WithKey("metadata").Contains(utils.MetadataNamespaceKeyVal(namespace)).
				Done(),
		),
	)

	if err := m.RetrieveMany(mctx, &eps); err != nil {
		return nil, err
	}
	if len(eps) == 0 {
		return nil, fmt.Errorf("no enforcer profile '%s' found in namespace '%s'", name, namespace)
	}
	if len(eps) > 1 {
		return nil, fmt.Errorf("multiple (%d) enforcer profile policies found with name '%s' in namespace '%s'", len(eps), name, namespace)
	}

	return eps[0], nil
}
