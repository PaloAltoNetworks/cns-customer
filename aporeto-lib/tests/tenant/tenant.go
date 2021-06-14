package tenant

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/constants"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/extnetwork"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/namespace"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/tenant"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/zone"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/tests/testapi"
	"go.aporeto.io/apotests-lib/logging"

	"go.aporeto.io/apocheck"
	"go.aporeto.io/gaia"
	"go.aporeto.io/manipulate"
	"go.aporeto.io/manipulate/maniphttp"
)

const (
	zoneName      = "test-zone"
	tenantName    = "test-tenant"
	zoneName1     = "DMZ"
	tenantName1   = "tenant"
	privateRail   = "private"
	protectedRail = "protected"
	publicRail    = "public"

	associatedTagNamespaceKey = "cns-customer:namespace="
)

func init() {

	apocheck.RegisterTest(apocheck.Test{
		Name:        "Tenant Creation Test",
		Description: "Test to ensure everything that is needed in tenant creation is working per expectation",
		Author:      "Satyam",
		Tags:        []string{"suite:tenant", "tenant"},
		Setup: func(ctx context.Context, t apocheck.TestInfo) (interface{}, apocheck.TearDownFunction, error) {

			logger := logging.DefaultLoggerWithOutput(t.TestID(), os.Stdout)

			zn := zone.Zone{
				Account: maniphttp.ExtractNamespace(t.PublicManipulator()),
				Name:    zoneName + "-" + t.TestID(),
			}

			apocheck.Step(t, "Given I setup a zone on aporeto", func() (err error) {

				err = zn.Create(ctx, t.PublicManipulator())
				if err != nil {
					logger.LogError(fmt.Sprintln(err))
				}
				return err
			})

			return nil, func() {
				if err := zn.Delete(ctx, t.PublicManipulator()); err != nil {
					logger.LogError(fmt.Sprintln(err))
				}
			}, nil
		},

		Function: func(ctx context.Context, t apocheck.TestInfo) error {

			logger := logging.DefaultLoggerWithOutput(t.TestID(), os.Stdout)

			tn := tenant.Tenant{
				Account:               maniphttp.ExtractNamespace(t.PublicManipulator()),
				Zone:                  zoneName + "-" + t.TestID(),
				Name:                  tenantName,
				Description:           "some description",
				AuthPolicyClaims:      [][]string{{"@auth:realm=claim", "@auth:company=aporeto"}},
				AuthPolicyDescription: "some auth policy description",
			}

			apocheck.Step(t, "Given I setup a tenant on aporeto", func() (err error) {

				err = tn.Create(ctx, t.PublicManipulator())
				if err != nil {
					logger.LogError(fmt.Sprintln(err))
				}
				return err
			})

			apocheck.Step(t, "Then I make sure tenant namespaces exist correctly", func() (err error) {

				n := namespace.Namespace{
					ParentNamespace: setupNamespaceString(tn.Account, tn.Zone),
					Name:            tenantName,
				}
				_, err = n.Get(ctx, t.PublicManipulator())

				if n.Name != tenantName {
					return fmt.Errorf("tenant's name does NOT match the created name, expected %s, got %s", tenantName, n.Name)
				}

				if err != nil {
					logger.LogError(fmt.Sprintln(err))
				}
				return err
			})

			apocheck.Step(t, "Then I make sure public rail namespaces exist correctly", func() (err error) {

				n := namespace.Namespace{
					ParentNamespace: setupNamespaceString(tn.Account, tn.Zone, tn.Name),
					Name:            "public",
				}
				_, err = n.Get(ctx, t.PublicManipulator())
				if err != nil {
					logger.LogError(fmt.Sprintln(err))
				}
				return err
			})

			apocheck.Step(t, "Then I make sure private rail namespaces exist correctly", func() (err error) {

				n := namespace.Namespace{
					ParentNamespace: setupNamespaceString(tn.Account, tn.Zone, tn.Name),
					Name:            "private",
				}
				_, err = n.Get(ctx, t.PublicManipulator())
				if err != nil {
					logger.LogError(fmt.Sprintln(err))
				}
				return err
			})

			apocheck.Step(t, "Then I make sure protected rail namespaces exist correctly", func() (err error) {

				n := namespace.Namespace{
					ParentNamespace: setupNamespaceString(tn.Account, tn.Zone, tn.Name),
					Name:            "protected",
				}
				_, err = n.Get(ctx, t.PublicManipulator())
				if err != nil {
					logger.LogError(fmt.Sprintln(err))
				}
				return err
			})

			apocheck.Step(t, "Then I make sure all-tcp external network objects exist correctly", func() (err error) {

				e := extnetwork.ExternalNetwork{
					Account: tn.Account,
					Zone:    tn.Zone,
					Tenant:  tn.Name,
					Name:    constants.ExternalNetworkAllTCP,
				}
				_, err = e.Get(ctx, t.PublicManipulator())
				if err != nil {
					logger.LogError(fmt.Sprintln(err))
				}
				return err
			})

			apocheck.Step(t, "Then I make sure all-udp external network objects exist correctly", func() (err error) {

				e := extnetwork.ExternalNetwork{
					Account: tn.Account,
					Zone:    tn.Zone,
					Tenant:  tn.Name,
					Name:    constants.ExternalNetworkAllUDP,
				}
				_, err = e.Get(ctx, t.PublicManipulator())
				if err != nil {
					logger.LogError(fmt.Sprintln(err))
				}
				return err
			})

			apocheck.Step(t, "Then I make sure all policies exist correctly", func() (err error) {

				options := []manipulate.ContextOption{
					manipulate.ContextOptionNamespace(tn.Account + "/" + tn.Zone + "/" + tn.Name),
				}
				npl, err := testapi.GetMany(ctx, t.PublicManipulator(), &gaia.NetworkAccessPoliciesList{}, options)
				if err != nil {
					return err
				}

				npList := *npl.(*gaia.NetworkAccessPoliciesList)

				if len(npList) != 10 {
					return fmt.Errorf("error: number of external networks do not match, expected 10 got %d", len(npList))
				}

				return nil
			})

			apocheck.Step(t, "Then I disable the tenant from aporeto", func() (err error) {

				err = tn.Disable(ctx, t.PublicManipulator())
				if err != nil {
					logger.LogError(fmt.Sprintln(err))
				}
				return err
			})

			apocheck.Step(t, "Then I make sure there is a 'reject' policy prefixed 'disble ...' exists", func() (err error) {

				options := []manipulate.ContextOption{
					manipulate.ContextOptionNamespace(tn.Account),
				}

				n, err := testapi.GetMany(ctx, t.PublicManipulator(), &gaia.NetworkAccessPoliciesList{}, options)
				if err != nil {
					return err
				}

				netPolicyList := *n.(*gaia.NetworkAccessPoliciesList)

				for _, v := range netPolicyList {
					if v.GetName() == "disable "+tn.Account+"/"+tn.Zone+"/"+tn.Name && v.Action == "Reject" {
						return nil
					}
				}

				return fmt.Errorf("disable tenant policy not found for  %s", tn.Name)
			})

			apocheck.Step(t, "Then I remove the tenant from aporeto", func() (err error) {

				err = tn.Delete(ctx, t.PublicManipulator())
				if err != nil {
					logger.LogError(fmt.Sprintln(err))
				}
				return err
			})

			apocheck.Step(t, "Then I make sure namespace is deleted correctly", func() (err error) {

				n := namespace.Namespace{
					ParentNamespace: setupNamespaceString(tn.Account, tn.Zone),
					Name:            tn.Name,
				}
				_, err = n.Get(ctx, t.PublicManipulator())
				if err == nil {
					return errors.New("tenant still exists")
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
