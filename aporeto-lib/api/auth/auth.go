package auth

import (
	"context"
	"io/ioutil"
	"log"
	"time"

	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/internal/libs/authpolicy"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/internal/utils"
	"go.aporeto.io/gaia"
	"go.aporeto.io/manipulate"
	midgardclient "go.aporeto.io/midgard-lib/client"
)

// GenerateOneTimeToken generates a one time token for enforcer registration using an application credential located at a given path.
func GenerateOneTimeToken(ctx context.Context, appcred, namespace string) (string, error) {

	credsData, err := ioutil.ReadFile(appcred)
	if err != nil {
		return "", err
	}

	creds, tlsConfig, err := midgardclient.ParseCredentials(credsData)
	if err != nil {
		return "", err
	}

	if namespace == "" {
		namespace = creds.Namespace
	}

	mclient := midgardclient.NewClientWithTLS(creds.APIURL, tlsConfig)
	validity := 1 * time.Hour

	token, err := mclient.IssueFromCertificate(
		ctx,
		validity,
		midgardclient.OptQuota(4),
		midgardclient.OptOpaque(map[string]string{}),
		midgardclient.OptAudience("aud:*:*:"+namespace),
	)
	if err != nil || len(token) == 0 {
		return "", err
	}
	return token, nil
}

// Policy is the configuration for a tenant authorization policy.
type Policy struct {
	Account               string     `json:"account"`
	Zone                  string     `json:"zone"`
	Tenant                string     `json:"tenant"`
	Name                  string     `json:"name"`
	Description           string     `json:"description"`
	AuthPolicyClaims      [][]string `json:"auth-policy-claims"`
	AuthPolicyDescription string     `json:"auth-policy-description"`
}

// Create is an implementation of how to create a new tenant authorization policy which has the following:
//  - APIAuthorizationPolicy: tenants for readonly access to their namespace
func (p *Policy) Create(ctx context.Context, m manipulate.Manipulator) error {

	// Generate tenant namespace
	tenantNamespace := utils.SetupNamespaceString(p.Account, p.Zone, p.Tenant)

	// Create Read Only Authorization policies for tenants to access their namespace.
	err := authpolicy.Create(ctx, m, tenantNamespace, p.Name, p.AuthPolicyDescription, p.AuthPolicyClaims)
	if err != nil {
		log.Printf("unable to create authorization policy for tenant '%s' and children namespaces: %s\n", p.Tenant, err.Error())
		return err
	}

	return nil
}

// Delete is an implementation of how to delete all tenant authorization policies for a given tenant.
func (p *Policy) Delete(ctx context.Context, m manipulate.Manipulator) error {

	tenantNamespace := utils.SetupNamespaceString(p.Account, p.Zone, p.Tenant)

	return authpolicy.Delete(ctx, m, tenantNamespace, p.Name)
}

// Get is an implementation of how to get all tenant authorization policies for a given tenant.
func (p *Policy) Get(ctx context.Context, m manipulate.Manipulator) (gaia.APIAuthorizationPoliciesList, error) {

	tenantNamespace := utils.SetupNamespaceString(p.Account, p.Zone, p.Tenant)

	return authpolicy.Get(ctx, m, tenantNamespace, p.Name)
}
