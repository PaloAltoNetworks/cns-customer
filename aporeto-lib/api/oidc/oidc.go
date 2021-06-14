package oidc

import (
	"context"
	"fmt"

	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/internal/libs/oidc"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/internal/utils"
	"go.aporeto.io/gaia"
	"go.aporeto.io/manipulate"
)

// OIDC is the configuration for an OIDC provider
type OIDC struct {
	Account      string   `json:"account"`
	Zone         string   `json:"zone"`
	Tenant       string   `json:"tenant"`
	Name         string   `json:"name"`
	Endpoint     string   `json:"endpoint"`
	ClientID     string   `json:"clientID"`
	ClientSecret string   `json:"clientSecret"`
	Scopes       []string `json:"scopes"`
	Default      bool     `json:"default"`
	Subjects     []string `json:"subjects"`
}

// Create is an implementation of how to create a new OIDC provider which has the following:
func (o *OIDC) Create(ctx context.Context, m manipulate.Manipulator) error {

	// Generate tenant namespace
	tenantNamespace := utils.SetupNamespaceString(o.Account, o.Zone, o.Tenant)

	// Create and configure OIDC provider for tenants to allow OIDC users to login.
	err := oidc.Create(ctx, m, tenantNamespace, o.Name, o.Endpoint, o.ClientID, o.ClientSecret, o.Default, o.Scopes, o.Subjects)
	if err != nil {
		return fmt.Errorf("unable to create OIDC provider for tenant '%s' and children namespaces: %s", o.Tenant, err.Error())
	}

	return nil
}

// Delete is an implementation of how to delete OIDC provider for a given tenant.
func (o *OIDC) Delete(ctx context.Context, m manipulate.Manipulator) error {

	tenantNamespace := utils.SetupNamespaceString(o.Account, o.Zone, o.Tenant)

	return oidc.Delete(ctx, m, tenantNamespace, o.Name)
	//return nil
}

// Get is an implementation of how to get OIDC provider config for a given tenant.
func (o *OIDC) Get(ctx context.Context, m manipulate.Manipulator) (gaia.OIDCProvidersList, error) {

	tenantNamespace := utils.SetupNamespaceString(o.Account, o.Zone, o.Tenant)

	return oidc.Get(ctx, m, tenantNamespace, o.Name)
}
