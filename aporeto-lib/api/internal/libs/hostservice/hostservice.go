package hostservice

import (
	"context"
	"fmt"

	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/constants"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/internal/utils"
	"go.aporeto.io/elemental"
	"go.aporeto.io/gaia"
	"go.aporeto.io/manipulate"
)

// Create creates a host service.
// Params:
//   - namespace: namespace where profile will be created.
//   - name: name of host service.
//   - description: description of profile.
//   - services: defines all services
//   - hostModeEnabled: should entire host protection be turned on.
func Create(
	ctx context.Context,
	m manipulate.Manipulator,
	namespace, name, description string,
	services []string,
	hostModeEnabled bool,
) error {

	// Ensure namespaces are correctly formatted.
	namespace = utils.SetupNamespaceString(namespace)

	// Setup a new host service.
	hs := gaia.NewHostService()
	hs.Name = name
	hs.Description = description
	hs.AssociatedTags = append(utils.MakeNamespaceAssociatedTags(namespace), utils.MakeHostServiceAssociatedTags(name)...)
	hs.Metadata = utils.MakeTenantMetadata(namespace)
	hs.Services = services
	hs.HostModeEnabled = hostModeEnabled

	// Create a sub context so we dont retry too long.
	subctx, cancel := context.WithTimeout(ctx, constants.APIDefaultContextTimeout)
	defer cancel()

	// Create a namespace context where we are creating an object.
	mctx := manipulate.NewContext(
		subctx,
		manipulate.ContextOptionNamespace(namespace),
	)

	// Try creating multiple times in case of connection errors.
	return m.Create(mctx, hs)
}

// Delete deletes a host service.
func Delete(ctx context.Context, m manipulate.Manipulator, namespace, name string) error {

	// Ensure namespaces are correctly formatted.
	namespace = utils.SetupNamespaceString(namespace)

	// Get matching host services.
	hs, err := Get(ctx, m, namespace, name)
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
	return m.Delete(mctx, hs)
}

// Get fetches a list of host services matching the criteria.
func Get(ctx context.Context, m manipulate.Manipulator, namespace, name string) (*gaia.HostService, error) {

	hss := gaia.HostServicesList{}
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

	if err := m.RetrieveMany(mctx, &hss); err != nil {
		return nil, err
	}
	if len(hss) == 0 {
		return nil, fmt.Errorf("no host service '%s' found in namespace '%s'", name, namespace)
	}
	if len(hss) > 1 {
		return nil, fmt.Errorf("multiple (%d) host service found with name '%s' in namespace '%s'", len(hss), name, namespace)
	}

	return hss[0], nil
}
