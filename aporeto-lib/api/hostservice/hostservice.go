package hostservice

import (
	"context"

	"go.aporeto.io/gaia"
	"go.aporeto.io/manipulate"

	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/internal/libs/hostservice"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/internal/utils"
)

// Service definition.
type Service struct {
	Account         string   `json:"account"`
	Zone            string   `json:"zone"`
	Tenant          string   `json:"tenant"`
	Name            string   `json:"name"`
	Rail            string   `json:"rail"`
	Definition      []string `json:"definition"`
	Description     string   `json:"description"`
	HostModeEnabled bool     `json:"hostmodeenabled"`
}

// Create is an implementation of how to setup a tenant host service.
func (s *Service) Create(ctx context.Context, m manipulate.Manipulator) error {

	railNamespace := utils.SetupNamespaceString(s.Account, s.Zone, s.Tenant, s.Rail)
	return hostservice.Create(ctx, m, railNamespace, s.Name, s.Description, s.Definition, s.HostModeEnabled)
}

// Delete is an implementation of how to delete a tenant host service.
func (s *Service) Delete(ctx context.Context, m manipulate.Manipulator) error {

	railNamespace := utils.SetupNamespaceString(s.Account, s.Zone, s.Tenant, s.Rail)
	return hostservice.Delete(ctx, m, railNamespace, s.Name)
}

// Get is an implementation of how to get a tenant host service.
func (s *Service) Get(ctx context.Context, m manipulate.Manipulator) (*gaia.HostService, error) {

	railNamespace := utils.SetupNamespaceString(s.Account, s.Zone, s.Tenant, s.Rail)
	return hostservice.Get(ctx, m, railNamespace, s.Name)
}
