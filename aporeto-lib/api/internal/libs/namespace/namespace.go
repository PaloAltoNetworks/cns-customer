package namespace

import (
	"context"
	"fmt"

	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/constants"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/internal/utils"
	"go.aporeto.io/elemental"
	"go.aporeto.io/gaia"
	"go.aporeto.io/manipulate"
)

// Create creates a namespace
func Create(ctx context.Context, m manipulate.Manipulator, parentNamespace, name, description string) error {

	ns := gaia.NewNamespace()
	ns.Name = name
	ns.Description = description
	ns.Metadata = utils.MakeOwnerMetadata()

	subctx, cancel := context.WithTimeout(ctx, constants.APIDefaultContextTimeout)
	defer cancel()

	mctx := manipulate.NewContext(
		subctx,
		manipulate.ContextOptionNamespace(parentNamespace),
	)
	return m.Create(mctx, ns)
}

// Delete deletes a namespace
func Delete(ctx context.Context, m manipulate.Manipulator, parentNamespace, name string) error {

	ns, err := Get(ctx, m, parentNamespace, name)
	if err != nil {
		return err
	}

	subctx, cancel := context.WithTimeout(ctx, constants.APIDefaultContextTimeout)
	defer cancel()

	mctx := manipulate.NewContext(
		subctx,
		manipulate.ContextOptionNamespace(parentNamespace),
	)
	return m.Delete(mctx, ns)
}

// Get fetches a namespace
func Get(ctx context.Context, m manipulate.Manipulator, parentNamespace, name string) (*gaia.Namespace, error) {

	nsl := gaia.NamespacesList{}

	subctx, cancel := context.WithTimeout(ctx, constants.APIDefaultContextTimeout)
	defer cancel()

	mctx := manipulate.NewContext(
		subctx,
		manipulate.ContextOptionNamespace(parentNamespace),
		manipulate.ContextOptionFilter(
			elemental.NewFilterComposer().
				WithKey("name").Equals(utils.SetupNamespaceString(parentNamespace, name)).
				Done(),
		),
	)

	if err := m.RetrieveMany(mctx, &nsl); err != nil {
		return nil, err
	}
	if len(nsl) == 0 {
		return nil, fmt.Errorf("no namespace '%s' found in namespace '%s'", name, parentNamespace)
	}
	if len(nsl) > 1 {
		return nil, fmt.Errorf("multiple (%d) namespaces found with name '%s' in parent namespace '%s'", len(nsl), name, parentNamespace)
	}
	return nsl[0], nil
}
