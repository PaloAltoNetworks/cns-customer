package extnetwork

import (
	"context"
	"fmt"

	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/constants"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/internal/libs/externalnetwork"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/internal/utils"
	"go.aporeto.io/elemental"
	"go.aporeto.io/gaia"
	"go.aporeto.io/manipulate"
)

// ExternalNetwork is the configuration for a tenant.
type ExternalNetwork struct {
	Account     string   `json:"account"`
	Zone        string   `json:"zone"`
	Tenant      string   `json:"tenant"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	CIDRs       []string `json:"cidrs"`
	Ports       []string `json:"ports"`
	Protocols   []string `json:"protocols"`
}

// Create creates an external network
func (e *ExternalNetwork) Create(ctx context.Context, m manipulate.Manipulator) error {

	namespace := utils.SetupNamespaceString(e.Account, e.Zone, e.Tenant)

	err := externalnetwork.Create(
		ctx,
		m,
		namespace,
		e.Name,
		e.Description,
		e.CIDRs,
		e.Ports,
		e.Protocols,
	)
	if err != nil {
		return fmt.Errorf("unable to create external network '%s' in tenant '%s': %s", e.Name, namespace, err.Error())
	}

	return nil
}

// Delete deletes an external network.
func (e *ExternalNetwork) Delete(ctx context.Context, m manipulate.Manipulator) error {

	namespace := utils.SetupNamespaceString(e.Account, e.Zone, e.Tenant)

	// Get matching external networks.
	en, err := e.Get(ctx, m)
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
func (e *ExternalNetwork) Get(ctx context.Context, m manipulate.Manipulator) (*gaia.ExternalNetwork, error) {

	namespace := utils.SetupNamespaceString(e.Account, e.Zone, e.Tenant)

	ens := gaia.ExternalNetworksList{}

	subctx, cancel := context.WithTimeout(ctx, constants.APIDefaultContextTimeout)
	defer cancel()

	mctx := manipulate.NewContext(
		subctx,
		manipulate.ContextOptionNamespace(utils.SetupNamespaceString(namespace)),
		manipulate.ContextOptionFilter(
			elemental.NewFilterComposer().
				WithKey("name").Equals(e.Name).
				Done(),
		),
	)

	if err := m.RetrieveMany(mctx, &ens); err != nil {
		return nil, err
	}
	if len(ens) == 0 {
		return nil, fmt.Errorf("no external network '%s' found in namespace '%s'", e.Name, namespace)
	}
	if len(ens) > 1 {
		return nil, fmt.Errorf("multiple (%d) external network found with name '%s' in namespace '%s'", len(ens), e.Name, namespace)
	}

	return ens[0], nil
}
