package tenant

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"strings"

	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/appcred"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/constants"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/extnetwork"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/internal/libs/authpolicy"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/internal/libs/enforcerprofile"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/internal/libs/enforcerprofilemapping"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/internal/libs/hostservice"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/internal/libs/hostservicemapping"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/internal/libs/namespace"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/internal/libs/networkpolicy"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/internal/utils"
	"go.aporeto.io/gaia"
	"go.aporeto.io/manipulate"
)

// Tenant is the configuration for a tenant.
type Tenant struct {
	Account     string `json:"account"`
	Zone        string `json:"zone"`
	Name        string `json:"name"`
	Description string `json:"description"`

	// AuthPolicyClaims if has length 0, no auth policy will be created
	AuthPolicyClaims      [][]string `json:"auth-policy-claims"`
	AuthPolicyDescription string     `json:"auth-policy-description"`

	// EnforcerAppCredPath if set to "" will not generate appcreds
	EnforcerAppCredPath string `json:"enforcer-app-cred-path"`
}

// Create is an implementation of how to create a new tenant which has the following:
//  - Rails: public, private, protected.
//  - EnforcerProfiles: one each for each rail.
//  - EnforcerProfileMapping: one each for each rail. this maps all VMs in a rail namespace to its dedicated EnforcerProfile.
//  - HostServices: mgmt services (ssh). one each for each rail.
//  - HostServiceMapping: one each for each rail. this maps all VMs in a rail namespace to its dedicated set of services.
//  - ExternalNetworks: all-tcp and all-udp
//  - DefaultPolicies:
//               - allow all outgoing (external networks all-tcp and all-udp and all pus) unidirectional
//               - allow mgmt (all-tcp to mgmt)
//               - public -> protected
//               - public can talk to itself
//               - protected -> private
//               - protected can talk to itself
//               - private can talk to itself
//  - APIAuthorizationPolicy: tenants for readonly access to their namespace
//  - AppCreds: one for each rail to provision enforcer one time tokens
func (t *Tenant) Create(ctx context.Context, m manipulate.Manipulator) error {

	zoneNamespace := utils.SetupNamespaceString(t.Account, t.Zone)

	// Creation of tenant namespace in a zone alongwith child namespaces public, protected and private.
	err := createNamespaces(ctx, m, zoneNamespace, t.Name, t.Description)
	if err != nil {
		log.Printf("unable to create tenant '%s' and children namespaces: %s\n", t.Name, err.Error())
		return err
	}

	// Creation of enforcer profiles and enforcer profile mappings
	err = createEnforcerProfilesAndMappingPolicies(ctx, m, t.Account, t.Zone, t.Name)
	if err != nil {
		log.Printf("unable to create enforcer profiles for tenant '%s': %s\n", t.Name, err.Error())
		return err
	}

	// Creation of host service profiles and service profile mappings
	err = createHostServiceAndMappingPolicies(ctx, m, t.Account, t.Zone, t.Name)
	if err != nil {
		log.Printf("unable to create host service profiles for tenant '%s': %s\n", t.Name, err.Error())
		return err
	}

	// Create external networks.
	err = createExternalNetworks(ctx, m, t.Account, t.Zone, t.Name)
	if err != nil {
		log.Printf("unable to create external networks for tenant '%s': %s\n", t.Name, err.Error())
		return err
	}

	// Creation of tenant default policies.
	err = createDefaultPolicies(ctx, m, t.Account, t.Zone, t.Name, t.Description)
	if err != nil {
		log.Printf("unable to create tenant '%s' and children namespaces: %s\n", t.Name, err.Error())
		return err
	}

	// Generate tenant namespace
	tenantNamespace := utils.SetupNamespaceString(zoneNamespace, t.Name)

	if len(t.AuthPolicyClaims) != 0 {
		err = authpolicy.Create(ctx, m, tenantNamespace, constants.DefaultTenantROAuthPolicy, t.AuthPolicyDescription, t.AuthPolicyClaims)
		if err != nil {
			log.Printf("unable to create authorization policy for tenant '%s' and children namespaces: %s\n", t.Name, err.Error())
			return err
		}
	}

	if t.EnforcerAppCredPath != "" {
		// Creation of tenant Application Credentials to generate enforcer one time token.
		err = createEnforcerAppcreds(ctx, m, t.Account, t.Zone, t.Name, t.EnforcerAppCredPath)
		if err != nil {
			log.Printf("unable to create tenant '%s' and children namespaces: %s\n", t.Name, err.Error())
			return err
		}
	}

	// If claims contains an account tag then create an AWS Registration Policy within the tenant subnamespace
	// TODO: get accountid to pass into here.
	//CreateAWSAutoRegistrationAuth(ctx, m, t.Account, tenantNamespace, t.AuthPolicyDescription, t.AuthPolicyClaims)

	return nil
}

// Disable is an implementation of how to disable a tenant (keep the tenant but not allow any communication). This is done by:
//  - Delete AppCreds: no enforcer can register anymore.
//  - Delete AuthPolicy: tenants cant log into this namespace (Admins can)
//  - Create NetworkAccessPolicy: no outbound or inbound communication can happen.
func (t *Tenant) Disable(ctx context.Context, m manipulate.Manipulator) error {

	zoneNamespace := utils.SetupNamespaceString(t.Account, t.Zone)

	// Delete Application Credentials to register new enforcers.
	_ = deleteEnforcerAppcreds(ctx, m, t.Account, t.Zone, t.Name) // nolint

	// Generate tenant namesapace
	tenantNamespace := utils.SetupNamespaceString(zoneNamespace, t.Name)

	// Delete Read Only Authorization policies for tenants to access their namespace.
	_ = authpolicy.Delete(ctx, m, tenantNamespace, constants.DefaultTenantROAuthPolicy) // nolint

	// Create rules that block traffic from and to this tenant.
	err := createDisablePolicies(ctx, m, t.Account, t.Zone, t.Name, t.Description)
	if err != nil {
		log.Printf("unable to create network access policies to disable tenant '%s': %s\n", t.Name, err.Error())
		return err
	}

	return nil
}

// Delete is an implementation of how to delete everything related to this tenant.
//  - Delete Namespace: everything recursively is removed.
//  - Delete All Tenant Exception Policies.
func (t *Tenant) Delete(ctx context.Context, m manipulate.Manipulator) error {

	zoneNamespace := utils.SetupNamespaceString(t.Account, t.Zone)

	// Delete the tenant namespace. This will delete all objects in that namespace and children.
	err := namespace.Delete(ctx, m, zoneNamespace, t.Name)
	if err != nil {
		log.Printf("unable to delete zone '%s' namespace in account '%s': %s\n", t.Zone, t.Account, err.Error())
		return err
	}

	// Delete rules that block traffic from and to this tenant.
	err = deleteDisablePolicies(ctx, m, t.Account, t.Zone, t.Name)
	if err != nil {
		log.Printf("unable to delete network access policies to disable tenant '%s': %s\n", t.Name, err.Error())
		return err
	}

	// Delete all rules for this tenant at account level which have this tenant.
	err = deleteTenantPolicies(ctx, m, t.Account, t.Zone, t.Name)
	if err != nil {
		log.Printf("unable to delete network access policies to disable tenant '%s': %s\n", t.Name, err.Error())
		return err
	}

	return nil
}

// createNamespaces creates a tenant namespace in the namespace hierarchy
// /account/zone/tenant with a description specified in the description parameter.
// It also creates the children namespaces.
func createNamespaces(ctx context.Context, m manipulate.Manipulator, zoneNamespace, tenant, description string) error {

	// Create Tenant
	err := namespace.Create(ctx, m, zoneNamespace, tenant, description)
	if err != nil {
		return fmt.Errorf("unable to create tenant '%s' in zone %s: %s", tenant, zoneNamespace, err.Error())
	}

	tenantNamespace := utils.SetupNamespaceString(zoneNamespace, tenant)

	// Create Public
	err = namespace.Create(ctx, m, tenantNamespace, constants.NamespacePublic, description)
	if err != nil {
		return fmt.Errorf("unable to create namespace '%s' in tenant '%s': %s", constants.NamespacePublic, tenantNamespace, err.Error())
	}

	// Create Protected
	err = namespace.Create(ctx, m, tenantNamespace, constants.NamespaceProtected, description)
	if err != nil {
		return fmt.Errorf("unable to create namespace '%s' in tenant '%s': %s", constants.NamespaceProtected, tenantNamespace, err.Error())
	}

	// Create Private
	err = namespace.Create(ctx, m, tenantNamespace, constants.NamespacePrivate, description)
	if err != nil {
		return fmt.Errorf("unable to create namespace '%s' in tenant '%s': %s", constants.NamespacePrivate, tenantNamespace, err.Error())
	}

	return nil
}

// createEnforcerProfilesAndMappingPolicies creates a enforcer profiles in tenant namespace one each for public, protected and private.
func createEnforcerProfilesAndMappingPolicies(ctx context.Context, m manipulate.Manipulator, account, zone, tenant string) error {

	tenantNs := utils.SetupNamespaceString(account, zone, tenant)
	children := []string{constants.NamespacePublic, constants.NamespaceProtected, constants.NamespacePrivate}
	for i := range children {

		child := children[i]
		childNs := utils.SetupNamespaceString(tenantNs, child)

		description := fmt.Sprintf("enforcer profile utilized by all enforcers for tenant %s in %s namespace", tenantNs, child)
		err := enforcerprofile.Create(ctx, m, childNs, child, description)
		if err != nil {
			return fmt.Errorf("unable to create enforcer profile '%s' in tenant '%s': %s", child, tenantNs, err.Error())
		}

		description = fmt.Sprintf("enforcer profile mapping to map all enforcers for tenant %s in %s namespace", tenantNs, child)
		err = enforcerprofilemapping.Create(ctx, m, childNs, child, description)
		if err != nil {
			return fmt.Errorf("unable to create enforcer profile mapping '%s' in tenant '%s': %s", child, tenantNs, err.Error())
		}
	}

	return nil
}

// createHostServiceAndMappingPolicies creates a host profiles in tenant namespace one each for public, protected and private.
func createHostServiceAndMappingPolicies(ctx context.Context, m manipulate.Manipulator, account, zone, tenant string) error {

	tenantNs := utils.SetupNamespaceString(account, zone, tenant)

	children := []string{constants.NamespacePublic, constants.NamespaceProtected, constants.NamespacePrivate}
	for i := range children {

		child := children[i]
		childNs := utils.SetupNamespaceString(tenantNs, child)

		description := fmt.Sprintf("management host service for rail %s in %s namespace", tenantNs, child)
		err := hostservice.Create(ctx, m, childNs, constants.ManagementServiceName, description, []string{constants.ManagementServices}, false)
		if err != nil {
			return fmt.Errorf("unable to create host service '%s' in tenant '%s': %s", child, tenantNs, err.Error())
		}

		description = fmt.Sprintf("host service mapping to map all enforcers for tenant %s in %s namespace", tenantNs, child)
		err = hostservicemapping.Create(ctx, m, childNs, child, description)
		if err != nil {
			return fmt.Errorf("unable to create host service mapping '%s' in tenant '%s': %s", child, tenantNs, err.Error())
		}
	}

	return nil
}

// createExternalNetworks creates external networks in tenant namespace.
func createExternalNetworks(ctx context.Context, m manipulate.Manipulator, account, zone, tenant string) error {

	tenantNs := utils.SetupNamespaceString(account, zone, tenant)

	e := extnetwork.ExternalNetwork{
		Account:     account,
		Zone:        zone,
		Tenant:      tenant,
		Name:        constants.ExternalNetworkAllTCP,
		Description: fmt.Sprintf("default %s external network tenant %s", constants.ExternalNetworkAllTCP, tenantNs),
		CIDRs:       []string{constants.ExternalNetworkAnyCIDR},
		Ports:       []string{constants.ExternalNetworkAllPorts},
		Protocols:   []string{constants.ExternalNetworkProtcolTCP},
	}
	err := e.Create(ctx, m)
	if err != nil {
		return err
	}

	e = extnetwork.ExternalNetwork{
		Account:     account,
		Zone:        zone,
		Tenant:      tenant,
		Name:        constants.ExternalNetworkAllUDP,
		Description: fmt.Sprintf("default %s external network tenant %s", constants.ExternalNetworkAllUDP, tenantNs),
		CIDRs:       []string{constants.ExternalNetworkAnyCIDR},
		Ports:       []string{constants.ExternalNetworkAllPorts},
		Protocols:   []string{constants.ExternalNetworkProtcolUDP},
	}
	err = e.Create(ctx, m)
	if err != nil {
		return err
	}

	return nil
}

// createIntraTenantPolicy creates the specified network access policy to allow traffic within the tenant
func createIntraTenantPolicy(ctx context.Context, m manipulate.Manipulator, tenantNs, name, description, srcNs, dstNs string) error {

	srcNsTag := "$namespace=" + srcNs
	srcTag := [][]string{{srcNsTag}}

	dstNsTag := "$namespace=" + dstNs
	dstTag := [][]string{{dstNsTag}}

	return networkpolicy.Create(
		ctx,
		m,
		tenantNs,
		name,
		description,
		srcNs,
		dstNs,
		srcTag,
		dstTag,
		gaia.NetworkAccessPolicyApplyPolicyModeIncomingTraffic,
		gaia.NetworkAccessPolicyActionAllow,
		false,
	)
}

// createMgmtTenantPolicy creates the network access policy to allow management traffic (SSH) into the tenant namespace
func createMgmtTenantPolicy(ctx context.Context, m manipulate.Manipulator, tenantNs string) error {

	srcNsTag := "$namespace=" + tenantNs
	srcTag := [][]string{{srcNsTag}}
	srcTag[0] = append(srcTag[0], utils.MakeExternalNetworkAssociatedTags(constants.ExternalNetworkAllTCP)...)

	dstNsWildcardTag := "$namespace=" + utils.SetupNamespaceString(tenantNs, "*")
	dstTag := [][]string{{dstNsWildcardTag}}
	dstTag[0] = append(dstTag[0], utils.MakeHostServiceAssociatedTags(constants.ManagementServiceName)...)

	name := fmt.Sprintf("management %s for tenant %s", constants.ManagementServiceName, tenantNs)
	description := fmt.Sprintf("allow all bidirectional management traffic to/from tenant %s", tenantNs)
	return networkpolicy.Create(
		ctx,
		m,
		tenantNs,
		name,
		description,
		tenantNs,
		tenantNs,
		srcTag,
		dstTag,
		gaia.NetworkAccessPolicyApplyPolicyModeBidirectional,
		gaia.NetworkAccessPolicyActionAllow,
		false,
	)
}

// createDefaultOutgoingTenantPolicy creates a network access policy to allow outgoing traffic from the tenant namespace
func createDefaultOutgoingTenantPolicy(ctx context.Context, m manipulate.Manipulator, tenantNs string) (ret error) {

	srcNsWildcardTag := "$namespace=" + utils.SetupNamespaceString(tenantNs, "*")
	srcTag := [][]string{{srcNsWildcardTag}}
	protocols := []string{constants.ExternalNetworkAllUDP, constants.ExternalNetworkAllTCP}

	for i := range protocols {
		dstNsTag := "$namespace=" + tenantNs
		dstTag := [][]string{{dstNsTag}}
		dstTag[0] = append(dstTag[0], utils.MakeExternalNetworkAssociatedTags(protocols[i])...)
		dstTag = append(dstTag, []string{"$identity=processingunit"})

		name := fmt.Sprintf("outgoing %s for tenant %s", protocols[i], tenantNs)
		description := fmt.Sprintf("allow all unidirectional outgoing traffic from tenant %s for %s", tenantNs, protocols[i])
		err := networkpolicy.Create(
			ctx,
			m,
			tenantNs,
			name,
			description,
			tenantNs,
			tenantNs,
			srcTag,
			dstTag,
			gaia.NetworkAccessPolicyApplyPolicyModeOutgoingTraffic,
			gaia.NetworkAccessPolicyActionAllow,
			false,
		)
		if err != nil {
			if ret == nil {
				ret = fmt.Errorf("unable to create a default policy for tenant '%s' %s", tenantNs, name)
			} else {
				ret = fmt.Errorf("%s and %s", err.Error(), name)
			}
		}
	}

	return ret
}

// createDefaultPolicies creates default policies for a tenant
func createDefaultPolicies(ctx context.Context, m manipulate.Manipulator, account, zone, tenant, description string) error {

	tenantNs := utils.SetupNamespaceString(account, zone, tenant)
	publicNs := utils.SetupNamespaceString(tenantNs, constants.NamespacePublic)
	protectedNs := utils.SetupNamespaceString(tenantNs, constants.NamespaceProtected)
	privateNs := utils.SetupNamespaceString(tenantNs, constants.NamespacePrivate)

	// Public to Public Allow
	if err := createIntraTenantPolicy(ctx, m, tenantNs, "accept intra-public", "unidirectional incoming traffic from public to public", publicNs, publicNs); err != nil {
		return fmt.Errorf("unable to create intra-public policy for tenant '%s': %s", tenantNs, err.Error())
	}

	// Public to Protected Allow
	if err := createIntraTenantPolicy(ctx, m, tenantNs, "accept from public to protected", "unidirectional incoming traffic from public to protected", publicNs, protectedNs); err != nil {
		return fmt.Errorf("unable to create public to protected policy for tenant '%s': %s", tenantNs, err.Error())
	}

	// Protected to Public Allow
	if err := createIntraTenantPolicy(ctx, m, tenantNs, "accept from protected to public", "unidirectional incoming traffic from protected to public", protectedNs, publicNs); err != nil {
		return fmt.Errorf("unable to create protected to public policy for tenant '%s': %s", tenantNs, err.Error())
	}

	// Protected to Protected Allow
	if err := createIntraTenantPolicy(ctx, m, tenantNs, "accept intra-protected", "unidirectional incoming traffic from protected to protected", protectedNs, protectedNs); err != nil {
		return fmt.Errorf("unable to create intra-protected policy for tenant '%s': %s", tenantNs, err.Error())
	}

	// Protected to Private Allow
	if err := createIntraTenantPolicy(ctx, m, tenantNs, "accept from protected to private", "unidirectional incoming traffic from protected to private", protectedNs, privateNs); err != nil {
		return fmt.Errorf("unable to create protected to private policy for tenant '%s': %s", tenantNs, err.Error())
	}

	// Private to Protected Allow
	if err := createIntraTenantPolicy(ctx, m, tenantNs, "accept from private to protected", "unidirectional incoming traffic from private to protected", privateNs, protectedNs); err != nil {
		return fmt.Errorf("unable to create private to protected policy for tenant '%s': %s", tenantNs, err.Error())
	}

	// Private to Private Allow
	if err := createIntraTenantPolicy(ctx, m, tenantNs, "accept intra-private", "unidirectional incoming traffic from private to private", privateNs, privateNs); err != nil {
		return fmt.Errorf("unable to create intra-private policy for tenant '%s': %s", tenantNs, err.Error())
	}

	// Allow All Management Policy
	if err := createMgmtTenantPolicy(ctx, m, tenantNs); err != nil {
		return fmt.Errorf("unable to create allow all management traffic policy for tenant '%s': %s", tenantNs, err.Error())
	}

	// Allow All Outgoing Policy
	if err := createDefaultOutgoingTenantPolicy(ctx, m, tenantNs); err != nil {
		return fmt.Errorf("unable to create unidirectional allow all outgoing policy for tenant '%s': %s", tenantNs, err.Error())
	}
	return nil
}

// createDisablePolicies creates disable policies for a tenant
func createDisablePolicies(ctx context.Context, m manipulate.Manipulator, account, zone, tenant, description string) error {

	accountNs := utils.SetupNamespaceString(account)
	accountWildcardNs := utils.SetupNamespaceString(accountNs, "*")
	accountWildcardNsTag := "$namespace=" + accountWildcardNs
	accountWildcardNsTags := [][]string{{accountWildcardNsTag}}

	tenantNs := utils.SetupNamespaceString(account, zone, tenant)
	tenantWildcardNs := utils.SetupNamespaceString(tenantNs, "*")
	tenantWildcardNsTag := "$namespace=" + tenantWildcardNs
	tenantWildcardNsTags := [][]string{{tenantWildcardNsTag}}
	name := "disable " + tenantNs

	return networkpolicy.Create(
		ctx,
		m,
		accountNs,
		name,
		name,
		tenantNs,
		accountNs,
		tenantWildcardNsTags,
		accountWildcardNsTags,
		gaia.NetworkAccessPolicyApplyPolicyModeBidirectional,
		gaia.NetworkAccessPolicyActionReject,
		false,
	)
}

// deleteDisablePolicies deletes disable policies for a tenant
func deleteDisablePolicies(ctx context.Context, m manipulate.Manipulator, account, zone, tenant string) error {

	accountNs := utils.SetupNamespaceString(account)
	tenantNs := utils.SetupNamespaceString(account, zone, tenant)
	name := "disable " + tenantNs

	return networkpolicy.Delete(
		ctx,
		m,
		accountNs,
		name,
	)
}

func write(name string, data []byte, out string) error {

	name = strings.Replace(name, " ", "-", -1)

	if out != "-" {
		return ioutil.WriteFile(path.Join(out, name), data, 0744)
	}

	fmt.Println(string(data))

	return nil
}

// createEnforcerAppcreds generates application credentials that can be used by CI pipeline to generate enforcer tokens.
func createEnforcerAppcreds(ctx context.Context, m manipulate.Manipulator, account, zone, tenant, dir string) error {

	if dir == "" {
		return fmt.Errorf("no output directory specified")
	}

	tenantNs := utils.SetupNamespaceString(account, zone, tenant)

	var js []byte
	errs := []error{nil, nil, nil}
	rails := []string{constants.NamespacePublic, constants.NamespaceProtected, constants.NamespacePrivate}
	errFlag := false
	for i, rail := range rails {

		railNs := utils.SetupNamespaceString(tenantNs, rail)
		js, errs[i] = appcred.Create(
			ctx,
			m,
			railNs,
			"enforcer-registration-"+rail,
			"appcred to generate one-time-token for enforcers in "+railNs,
			[]string{constants.AuthEnforcerd, constants.AuthEnforcerdRuntime},
		)

		if errs[i] == nil {
			errs[i] = write(fmt.Sprintf("enforcer-%s-%s-%s-creds.json", zone, tenant, rail), js, dir)
		}

		if errs[i] != nil {
			errFlag = true
			log.Printf("[error] failed to create application credential for tenant '%s' rail '%s'\n", tenant, rail)
			continue
		}
	}

	if errFlag {
		err := fmt.Errorf("failed to create some application credentials: ")
		for i := range errs {
			if errs[i] != nil {
				err = fmt.Errorf("%s \n %s", err.Error(), errs[i].Error())
			}
		}
		return err
	}

	return nil
}

// deleteEnforcerAppcreds removes application credentials that can be used by CI pipeline to generate enforcer tokens.
func deleteEnforcerAppcreds(ctx context.Context, m manipulate.Manipulator, account, zone, tenant string) (ret error) {

	tenantNs := utils.SetupNamespaceString(account, zone, tenant)

	rails := []string{constants.NamespacePublic, constants.NamespaceProtected, constants.NamespacePrivate}
	for _, rail := range rails {

		railNs := utils.SetupNamespaceString(tenantNs, rail)
		err := appcred.Delete(
			ctx,
			m,
			railNs,
			"enforcer-registration-"+rail,
		)
		if err != nil {
			if ret == nil {
				ret = fmt.Errorf("unable to delete an appcred for tenant '%s' rail(s) %s", tenantNs, rail)
			} else {
				ret = fmt.Errorf("%s and '%s'", err.Error(), rail)
			}
		}
	}
	return ret
}

// deleteTenantPolicies deletes all policies related to the tenant at account and zone levels if any
func deleteTenantPolicies(ctx context.Context, m manipulate.Manipulator, account, zone, tenant string) error {

	accountNs := utils.SetupNamespaceString(account)
	tenantNs := utils.SetupNamespaceString(account, zone, tenant)
	tenantMetadata := utils.MetadataTenantKeyVal(tenantNs)
	return networkpolicy.DeleteManyWithMetadata(ctx, m, accountNs, tenantMetadata)
}
