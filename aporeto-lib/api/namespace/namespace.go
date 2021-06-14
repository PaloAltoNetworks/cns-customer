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

// Namespace is the configuration for a tenant.
type Namespace struct {
	ParentNamespace string `json:"parent"`
	Name            string `json:"name"`
}

// Get fetches a list of external networks matching the criteria.
func (n *Namespace) Get(ctx context.Context, m manipulate.Manipulator) (*gaia.Namespace, error) {

	ens := gaia.NamespacesList{}

	subctx, cancel := context.WithTimeout(ctx, constants.APIDefaultContextTimeout)
	defer cancel()

	mctx := manipulate.NewContext(
		subctx,
		manipulate.ContextOptionNamespace(utils.SetupNamespaceString(n.ParentNamespace)),
		manipulate.ContextOptionFilter(
			elemental.NewFilterComposer().
				WithKey("name").Equals(utils.SetupNamespaceString(n.ParentNamespace, n.Name)).
				Done(),
		),
	)

	if err := m.RetrieveMany(mctx, &ens); err != nil {
		return nil, err
	}
	if len(ens) == 0 {
		return nil, fmt.Errorf("no namespace '%s' found in namespace '%s'", n.Name, n.ParentNamespace)
	}
	if len(ens) > 1 {
		return nil, fmt.Errorf("multiple (%d) namespace found with name '%s' in namespace '%s'", len(ens), n.Name, n.ParentNamespace)
	}

	return ens[0], nil
}
