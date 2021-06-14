package externalnetwork

import (
	"context"
	"fmt"

	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/constants"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/internal/utils"
	"go.aporeto.io/elemental"
	"go.aporeto.io/gaia"
	"go.aporeto.io/manipulate"
)

// Create creates an external network
// Params:
//   - namespace: namespace where external network will be created.
//   - name: name of external network.
//   - description: description of policy.
//   - cidrs: list of ip addresses.
//   - ports: list of ports.
//   - protocols: list of protocols.
func Create(
	ctx context.Context,
	m manipulate.Manipulator,
	namespace, name, description string,
	cidrs, ports, protocols []string,
) error {

	// Ensure namespaces are correctly formatted.
	namespace = utils.SetupNamespaceString(namespace)

	// Setup a new authorization policy.
	en := gaia.NewExternalNetwork()
	en.Name = name
	en.Description = description
	en.Entries = cidrs
	en.Ports = ports
	en.Protocols = protocols
	en.AssociatedTags = append(utils.MakeNamespaceAssociatedTags(namespace), utils.MakeExternalNetworkAssociatedTags(name)...)
	en.Metadata = utils.MakeTenantMetadata(namespace)

	// Create a sub context so we dont retry too long.
	subctx, cancel := context.WithTimeout(ctx, constants.APIDefaultContextTimeout)
	defer cancel()

	// Create a namespace context where we are creating an object.
	mctx := manipulate.NewContext(
		subctx,
		manipulate.ContextOptionNamespace(namespace),
	)

	// Try creating multiple times in case of connection errors.
	return m.Create(mctx, en)
}

// Delete deletes an external network.
func Delete(ctx context.Context, m manipulate.Manipulator, namespace, name string) error {

	// Ensure namespaces are correctly formatted.
	namespace = utils.SetupNamespaceString(namespace)

	// Get matching external networks.
	en, err := Get(ctx, m, namespace, name)
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
	return m.Delete(mctx, en)
}

// Get fetches a list of external networks matching the criteria.
func Get(ctx context.Context, m manipulate.Manipulator, namespace, name string) (*gaia.ExternalNetwork, error) {

	ens := gaia.ExternalNetworksList{}

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

	if err := m.RetrieveMany(mctx, &ens); err != nil {
		return nil, err
	}
	if len(ens) == 0 {
		return nil, fmt.Errorf("no external network '%s' found in namespace '%s'", name, namespace)
	}
	if len(ens) > 1 {
		return nil, fmt.Errorf("multiple (%d) external network found with name '%s' in namespace '%s'", len(ens), name, namespace)
	}

	return ens[0], nil
}
