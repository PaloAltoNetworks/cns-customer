package hostservicemapping

import (
	"context"
	"fmt"

	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/constants"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/internal/utils"
	"go.aporeto.io/elemental"
	"go.aporeto.io/gaia"
	"go.aporeto.io/manipulate"
)

// Create creates a host service mapping policy.
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

	// Setup a new host service mapping policy.
	hsm := gaia.NewHostServiceMappingPolicy()
	hsm.Name = name
	hsm.Description = description
	hsm.Subject = [][]string{utils.MakeNamespaceKeyVal(namespace)}
	hsm.Object = [][]string{utils.MakeNamespaceAssociatedTags(namespace)}
	hsm.AssociatedTags = utils.MakeNamespaceAssociatedTags(namespace)
	hsm.Metadata = utils.MakeTenantMetadata(namespace)

	// Create a sub context so we dont retry too long.
	subctx, cancel := context.WithTimeout(ctx, constants.APIDefaultContextTimeout)
	defer cancel()

	// Create a namespace context where we are creating an object.
	mctx := manipulate.NewContext(
		subctx,
		manipulate.ContextOptionNamespace(namespace),
	)

	// Try creating multiple times in case of connection errors.
	return m.Create(mctx, hsm)
}

// Delete deletes a host service mapping policy.
func Delete(ctx context.Context, m manipulate.Manipulator, namespace, name string) error {

	// Ensure namespaces are correctly formatted.
	namespace = utils.SetupNamespaceString(namespace)

	// Get matching host service mapping policies.
	hsm, err := Get(ctx, m, namespace, name)
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
	return m.Delete(mctx, hsm)
}

// Get fetches a list of host service mapping policies matching the criteria.
func Get(ctx context.Context, m manipulate.Manipulator, namespace, name string) (*gaia.HostServiceMappingPolicy, error) {

	hsms := gaia.HostServiceMappingPoliciesList{}

	subctx, cancel := context.WithTimeout(ctx, constants.APIDefaultContextTimeout)
	defer cancel()

	mctx := manipulate.NewContext(
		subctx,
		manipulate.ContextOptionNamespace(utils.SetupNamespaceString(namespace)),
		manipulate.ContextOptionFilter(
			elemental.NewFilterComposer().
				WithKey("name").Equals(name).
				Done(),
		),
	)

	if err := m.RetrieveMany(mctx, &hsms); err != nil {
		return nil, err
	}
	if len(hsms) == 0 {
		return nil, fmt.Errorf("no host service mapping '%s' found in namespace '%s'", name, namespace)
	}
	if len(hsms) > 1 {
		return nil, fmt.Errorf("multiple (%d) host service mapping policies found with name '%s' in namespace '%s'", len(hsms), name, namespace)
	}

	return hsms[0], nil
}
