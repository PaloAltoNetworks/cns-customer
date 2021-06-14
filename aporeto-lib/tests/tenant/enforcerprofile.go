package tenant

import (
	"context"
	"fmt"
	"strings"

	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/tenant"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/zone"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/tests/testapi"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/tests/testutil"
	"go.aporeto.io/apocheck"
	"go.aporeto.io/gaia"
	"go.aporeto.io/manipulate"
	"go.aporeto.io/manipulate/maniphttp"
)

const (
	defaultNumOfRails = 3
)

func init() {

	apocheck.RegisterTest(apocheck.Test{
		Name:        "Enforcer Profile Mapping Policy Validation Test",
		Description: "Create a zone and tenant with default policies and setup, verify that enforcer profiles and enforcer profile mapping policies are correct.",
		Author:      "Arvind",
		Tags:        []string{"suite:enforcer-profile"},
		Setup: func(ctx context.Context, t apocheck.TestInfo) (interface{}, apocheck.TearDownFunction, error) {

			apocheck.Step(t, "Given I create a test zone", func() error {

				zn := zone.New(maniphttp.ExtractNamespace(t.PublicManipulator()), zoneName, "This is a test zone, tenant will go in here")
				if err := zn.Create(ctx, t.PublicManipulator()); err != nil {
					return fmt.Errorf("zone creation failed: %s", err.Error())
				}

				return nil
			})

			apocheck.Step(t, "Given I create a test tenant", func() error {

				tn := tenant.Tenant{
					Account:               maniphttp.ExtractNamespace(t.PublicManipulator()),
					Zone:                  zoneName,
					Name:                  tenantName,
					Description:           zoneName + "/" + tenantName,
					EnforcerAppCredPath:   "./appcreds",
					AuthPolicyClaims:      [][]string{{"@auth:realm=claim", "@auth:company=aporeto"}},
					AuthPolicyDescription: "",
				}

				if err := tn.Create(ctx, t.PublicManipulator()); err != nil {
					return fmt.Errorf("tenant creation failed: %s", err.Error())
				}

				return nil
			})

			return nil, func() {

				zn := zone.New(maniphttp.ExtractNamespace(t.PublicManipulator()), zoneName, "This is a test zone, tenant will go in here")
				if err := zn.Delete(ctx, t.PublicManipulator()); err != nil {
					fmt.Errorf("zone deletion failed: %s", err.Error())
				}
			}, nil
		},

		Function: func(ctx context.Context, t apocheck.TestInfo) error {

			accountNamespace := maniphttp.ExtractNamespace(t.PublicManipulator())

			// get tenant namespace
			options := []manipulate.ContextOption{
				manipulate.ContextOptionNamespace(accountNamespace + "/" + zoneName),
			}

			nsList, err := testapi.GetMany(ctx, t.PublicManipulator(), &gaia.NamespacesList{}, options)
			if err != nil {
				return err
			}

			tenantList := *nsList.(*gaia.NamespacesList)
			tnName := accountNamespace + "/" + zoneName + "/" + tenantName

			tnfound := false
			for _, tn := range tenantList {
				if tn.GetName() == tnName {
					tnfound = true
					break
				}
			}

			if !tnfound {
				return fmt.Errorf("Expected created tenant's name does NOT match")
			}
			// get rail namespaces
			options = []manipulate.ContextOption{
				manipulate.ContextOptionNamespace(accountNamespace + "/" + zoneName + "/" + tenantName),
			}

			nsList, err = testapi.GetMany(ctx, t.PublicManipulator(), &gaia.NamespacesList{}, options)
			if err != nil {
				return err
			}

			railList := *nsList.(*gaia.NamespacesList)

			if len(railList) != defaultNumOfRails {
				return fmt.Errorf("error: number of expected rail namespaces does not match, expected %d got %d", defaultNumOfRails, len(railList))
			}

			// This map is so we can use this to reference policies and other things declared inside these rails
			railNSMap := map[string]*gaia.Namespace{}

			for i := range railList {
				railNSMap[railList[i].GetName()] = railList[i]
			}

			apocheck.Step(t, "I verify enforcer profiles and mapping policies", func() error {

				// enforcer profile validation
				for _, rail := range railNSMap {

					options = []manipulate.ContextOption{
						manipulate.ContextOptionNamespace(rail.GetName()),
					}

					profileList, err := testapi.GetMany(ctx, t.PublicManipulator(), &gaia.EnforcerProfilesList{}, options)
					if err != nil {
						return err
					}

					enfProfiles := *profileList.(*gaia.EnforcerProfilesList)

					if len(enfProfiles) != 1 {
						return fmt.Errorf("error: number of expected enforcer profiles for ns %s does not match, expected 1 got %d",
							rail.GetNamespace(),
							len(railList))
					}

					enforcerProfile := enfProfiles[0]

					found := false
					for _, enfProTag := range enforcerProfile.GetAssociatedTags() {
						if enfProTag == associatedTagNamespaceKey+rail.GetName() {
							found = true
						}
					}

					if !found {
						return fmt.Errorf("proper tag not found for enforcer profile %s", associatedTagNamespaceKey+rail.GetName())
					}
				}

				// enforcer profile mapping validation
				for _, rail := range railNSMap {

					options = []manipulate.ContextOption{
						manipulate.ContextOptionNamespace(rail.GetName()),
					}

					list, err := testapi.GetMany(ctx, t.PublicManipulator(), &gaia.EnforcerProfileMappingPoliciesList{}, options)
					if err != nil {
						return err
					}

					enfProfilesMPList := *list.(*gaia.EnforcerProfileMappingPoliciesList)

					enfPMP := gaia.EnforcerProfileMappingPolicy{}

					for i := range enfProfilesMPList {
						if enfProfilesMPList[i].GetName() == strings.TrimPrefix(rail.GetName(), rail.GetNamespace()+"/") {
							enfPMP = *enfProfilesMPList[i]
						}
					}

					// Verify that subject tag is proper and set to namespace of enforcer launch
					if enfPMP.Subject == nil {
						return fmt.Errorf("Subject is nil")
					}
					if len(enfPMP.Subject) == 0 {
						return fmt.Errorf("Subject is empty")
					}
					if len(enfPMP.Subject[0]) == 0 {
						return fmt.Errorf("a Subject line is empty")
					}

					if !testutil.CheckTagInTagClauses("$namespace="+rail.GetName(), enfPMP.Subject) {
						return fmt.Errorf("enforcer mapping policy subject tag is invalid, expected %s got %s",
							"namespace="+rail.GetName(), enfPMP.Subject[0][0])
					}

					// Verify that object tag is the CNS customer tag of enforcer profile
					if enfPMP.Object == nil {
						return fmt.Errorf("Object is nil")
					}
					if len(enfPMP.Object) == 0 {
						return fmt.Errorf("Object is empty")
					}
					if len(enfPMP.Object[0]) == 0 {
						return fmt.Errorf("an Object line is empty")
					}

					if !testutil.CheckTagInTagClauses(associatedTagNamespaceKey+rail.GetName(), enfPMP.Object) {
						return fmt.Errorf("enforcer mapping policy object tag is invalid, expected %s got %s",
							associatedTagNamespaceKey+rail.GetName(), enfPMP.Object[0][0])
					}

				}
				return nil
			})

			return nil
		},
	})
}
