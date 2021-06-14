package networkpolicy

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/networkpolicy"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/tenant"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/zone"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/tests/testapi"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/tests/testutil"
	"go.aporeto.io/apocheck"
	"go.aporeto.io/apotests-lib/logging"
	"go.aporeto.io/gaia"
	"go.aporeto.io/manipulate"
	"go.aporeto.io/manipulate/maniphttp"
)

const (
	tenantName    = "test-tenant"
	tenantName2   = "dmz-tenant"
	tenantName3   = "sensitive-tenant"
	zoneName      = "test-zone"
	zoneName2     = "dmz"
	zoneName3     = "sensitive"
	extnetName    = "some-extnet"
	privateRail   = "private"
	protectedRail = "protected"
)

func init() {

	apocheck.RegisterTest(apocheck.Test{
		Name:        "Network Policy Creation/Deletion Test",
		Description: "Test we can create/delete a network policy",
		Author:      "Satyam",
		Tags:        []string{"suite:network-policy"},
		Setup: func(ctx context.Context, t apocheck.TestInfo) (interface{}, apocheck.TearDownFunction, error) {

			logger := logging.DefaultLoggerWithOutput(t.TestID(), os.Stdout)

			zn := zone.Zone{
				Account: maniphttp.ExtractNamespace(t.PublicManipulator()),
				Name:    zoneName + "-" + t.TestID(),
			}

			tn := tenant.Tenant{
				Account:               maniphttp.ExtractNamespace(t.PublicManipulator()),
				Zone:                  zoneName + "-" + t.TestID(),
				Name:                  tenantName,
				Description:           "some description",
				AuthPolicyClaims:      [][]string{{"@auth:realm=claim", "@auth:company=aporeto"}},
				AuthPolicyDescription: "some auth policy description",
			}

			apocheck.Step(t, "Given I setup a zone on aporeto", func() (err error) {

				err = zn.Create(ctx, t.PublicManipulator())
				if err != nil {
					logger.LogError(fmt.Sprintln(err))
				}
				return err
			})

			apocheck.Step(t, "Given I setup a tenant on aporeto", func() (err error) {

				err = tn.Create(ctx, t.PublicManipulator())
				if err != nil {
					logger.LogError(fmt.Sprintln(err))
				}
				return err
			})

			return nil, func() {

				// Disable and Delete Tenant
				if err := tn.Disable(ctx, t.PublicManipulator()); err != nil {
					logger.LogError(fmt.Sprintln(err))
				}
				if err := tn.Delete(ctx, t.PublicManipulator()); err != nil {
					logger.LogError(fmt.Sprintln(err))
				}

				// Delete Zone
				if err := zn.Delete(ctx, t.PublicManipulator()); err != nil {
					logger.LogError(fmt.Sprintln(err))
				}
			}, nil
		},

		Function: func(ctx context.Context, t apocheck.TestInfo) error {

			logger := logging.DefaultLoggerWithOutput(t.TestID(), os.Stdout)
			account := maniphttp.ExtractNamespace(t.PublicManipulator())
			zone := zoneName + "-" + t.TestID()

			ns := setupNamespaceString(account, zone, tenantName)
			np := networkpolicy.NetworkPolicy{
				Namespace:              ns,
				Name:                   "test-policy",
				Description:            "z",
				SubjectTenantNamespace: ns,
				SubjectTags:            []string{"app=hello"},
				ObjectTenantNamespace:  ns,
				ObjectTags:             []string{"app=world"},
				PolicyMode:             "Bidirectional",
			}

			apocheck.Step(t, "Given I setup a network policy", func() (err error) {

				err = np.Create(ctx, t.PublicManipulator())
				if err != nil {
					logger.LogError(fmt.Sprintln(err))
				}
				return err
			})

			apocheck.Step(t, "Then I get the created network policy", func() (err error) {

				e, err := np.Get(ctx, t.PublicManipulator())
				if err != nil {
					logger.LogError(fmt.Sprintln(err))
					return err
				}
				if e == nil {
					logger.LogError(fmt.Sprintln("network policy not found"))
					return errors.New("network policy not found")
				}
				return nil
			})

			apocheck.Step(t, "Then I remove the network policy", func() (err error) {

				err = np.Delete(ctx, t.PublicManipulator())
				if err != nil {
					logger.LogError(fmt.Sprintln(err))
				}
				return err
			})

			apocheck.Step(t, "Then I get the deleted network policy", func() (err error) {

				_, err = np.Get(ctx, t.PublicManipulator())
				if err == nil {
					logger.LogError(fmt.Sprintln("network policy found"))
					return errors.New("network policy found")
				}
				return nil
			})

			return nil
		},
	})

	apocheck.RegisterTest(apocheck.Test{
		Name:        "Network Policy Integration Test",
		Description: "Create multiple zones and tenants with default policies and setup, verify that network policies and custom network policies are correct.",
		Author:      "Arvind",
		Tags:        []string{"suite:network-policy", "network-policy"},
		Setup: func(ctx context.Context, t apocheck.TestInfo) (interface{}, apocheck.TearDownFunction, error) {

			logger := logging.DefaultLoggerWithOutput(t.TestID(), os.Stdout)
			accountNamespace := maniphttp.ExtractNamespace(t.PublicManipulator())
			zone2 := zoneName2 + "-" + t.TestID()
			zone3 := zoneName3 + "-" + t.TestID()

			znDmz := zone.Zone{
				Account: accountNamespace,
				Name:    zone2,
			}
			znSensitive := zone.Zone{
				Account: accountNamespace,
				Name:    zone3,
			}

			tnDmz := tenant.Tenant{
				Account:               accountNamespace,
				Zone:                  zone2,
				Name:                  tenantName2,
				Description:           zoneName2 + "/" + tenantName2,
				EnforcerAppCredPath:   "./appcreds",
				AuthPolicyClaims:      [][]string{{"@auth:realm=claim", "@auth:company=aporeto"}},
				AuthPolicyDescription: "",
			}

			tnSensitive := tenant.Tenant{
				Account:               accountNamespace,
				Zone:                  zone3,
				Name:                  tenantName3,
				Description:           zoneName3 + "/" + tenantName3,
				EnforcerAppCredPath:   "./appcreds",
				AuthPolicyClaims:      [][]string{{"@auth:realm=claim", "@auth:company=aporeto"}},
				AuthPolicyDescription: "",
			}

			apocheck.Step(t, "Given I setup a dmz zone on aporeto", func() (err error) {

				err = znDmz.Create(ctx, t.PublicManipulator())
				if err != nil {
					logger.LogError(fmt.Sprintln(err))
				}
				return err
			})

			apocheck.Step(t, "Given I setup a sensitive zone on aporeto", func() (err error) {

				err = znSensitive.Create(ctx, t.PublicManipulator())
				if err != nil {
					logger.LogError(fmt.Sprintln(err))
				}
				return err
			})

			apocheck.Step(t, "Given I setup a tenant on dmz zone", func() (err error) {

				err = tnDmz.Create(ctx, t.PublicManipulator())
				if err != nil {
					logger.LogError(fmt.Sprintln(err))
				}
				return err
			})

			apocheck.Step(t, "Given I setup a tenant on sensitive zone", func() (err error) {

				err = tnSensitive.Create(ctx, t.PublicManipulator())
				if err != nil {
					logger.LogError(fmt.Sprintln(err))
				}
				return err
			})

			return nil, func() {

				// Disable and Delete Tenants
				if err := tnDmz.Disable(ctx, t.PublicManipulator()); err != nil {
					logger.LogError(fmt.Sprintln(err))
				}
				if err := tnSensitive.Disable(ctx, t.PublicManipulator()); err != nil {
					logger.LogError(fmt.Sprintln(err))
				}

				if err := tnDmz.Delete(ctx, t.PublicManipulator()); err != nil {
					logger.LogError(fmt.Sprintln(err))
				}
				if err := tnSensitive.Delete(ctx, t.PublicManipulator()); err != nil {
					logger.LogError(fmt.Sprintln(err))
				}

				// Delete Zones
				if err := znDmz.Delete(ctx, t.PublicManipulator()); err != nil {
					logger.LogError(fmt.Sprintln(err))
				}
				if err := znSensitive.Delete(ctx, t.PublicManipulator()); err != nil {
					logger.LogError(fmt.Sprintln(err))
				}
			}, nil
		},

		Function: func(ctx context.Context, t apocheck.TestInfo) error {

			zone2 := zoneName2 + "-" + t.TestID()
			zone3 := zoneName3 + "-" + t.TestID()

			accountNamespace := maniphttp.ExtractNamespace(t.PublicManipulator())

			tenantA := accountNamespace + "/" + zone2 + "/" + tenantName2
			tenantB := accountNamespace + "/" + zone3 + "/" + tenantName3

			nsPrivateTagsTenantA := []string{"$namespace=" + tenantA + "/" + "private", "$identity=processingunit"}
			nsProtectedTagsTenantB := []string{"$namespace=" + tenantB + "/" + "protected", "$identity=processingunit"}

			policyInterZoneAB := networkpolicy.NetworkPolicy{
				Namespace:              accountNamespace,
				Name:                   "inter-zone A to B allow traffic from private to protected",
				Description:            "inter-zone A to B bidirectional traffic from private to protected",
				SubjectTenantNamespace: accountNamespace,
				SubjectTags:            nsPrivateTagsTenantA,
				ObjectTenantNamespace:  accountNamespace,
				ObjectTags:             nsProtectedTagsTenantB,
				PolicyMode:             "Bidirectional",
			}

			policyInterZoneBA := networkpolicy.NetworkPolicy{
				Namespace:              accountNamespace,
				Name:                   "inter-zone B to A allow traffic from protected to private",
				Description:            "inter-zone B to A bidirectional traffic from protected to private",
				SubjectTenantNamespace: accountNamespace,
				SubjectTags:            nsProtectedTagsTenantB,
				ObjectTenantNamespace:  accountNamespace,
				ObjectTags:             nsPrivateTagsTenantA,
				PolicyMode:             "Bidirectional",
			}

			apocheck.Step(t, "I now create policies and validate them", func() error {

				if err := policyInterZoneAB.Create(ctx, t.PublicManipulator()); err != nil {
					return fmt.Errorf("unable to create inter-zone policy from tenant '%s/%s' to tenant '%s/%s': %s", tenantA, privateRail, tenantB, protectedRail, err.Error())
				}

				if err := policyInterZoneBA.Create(ctx, t.PublicManipulator()); err != nil {
					return fmt.Errorf("unable to create inter-zone policy from tenant '%s/%s' to tenant '%s/%s': %s", tenantB, protectedRail, tenantA, privateRail, err.Error())
				}

				options := []manipulate.ContextOption{
					manipulate.ContextOptionNamespace(accountNamespace),
				}

				npl, err := testapi.GetMany(ctx, t.PublicManipulator(), &gaia.NetworkAccessPoliciesList{}, options)
				if err != nil {
					return err
				}

				npList := *npl.(*gaia.NetworkAccessPoliciesList)

				var foundAB = false
				var foundBA = false

				for _, np := range npList {
					if testutil.CheckTagExistInAnyTagClauses("$namespace="+tenantA+"/private", np.Subject) {
						foundAB = true
						break
					}
				}

				for _, np := range npList {
					if testutil.CheckTagExistInAnyTagClauses("$namespace="+tenantB+"/protected", np.Subject) {
						foundBA = true
						break
					}
				}

				if !(foundAB && foundBA) {
					return fmt.Errorf("error: Expected created policies not found")
				}

				return nil
			})

			apocheck.Step(t, "I delete the dynamically created policies and verify that account namespace contains no policies", func() error {

				if err := policyInterZoneAB.Delete(ctx, t.PublicManipulator()); err != nil {
					return fmt.Errorf("unable to delete inter-zone policy from tenant '%s/%s' to tenant '%s/%s': %s", tenantA, privateRail, tenantB, protectedRail, err.Error())
				}

				if err := policyInterZoneBA.Delete(ctx, t.PublicManipulator()); err != nil {
					return fmt.Errorf("unable to delete inter-zone policy from tenant '%s/%s' to tenant '%s/%s': %s", tenantB, protectedRail, tenantA, privateRail, err.Error())
				}

				return nil
			})

			return nil
		},
	})

}

// setupNamespaceString returns a namespace string. It
// accounts for any leading/trailing '/'
func setupNamespaceString(namespaces ...string) string {

	absoluteNamespace := ""
	for _, namespace := range namespaces {
		ns := strings.TrimLeft(namespace, "/")
		ns = strings.TrimRight(ns, "/")
		absoluteNamespace = absoluteNamespace + "/" + ns
	}
	return absoluteNamespace
}
