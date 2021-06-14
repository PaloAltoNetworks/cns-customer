package extnetwork

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/extnetwork"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/tenant"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/zone"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/tests/testutil"
	"go.aporeto.io/apocheck"
	"go.aporeto.io/apotests-lib/logging"
	"go.aporeto.io/manipulate/maniphttp"
)

const (
	tenantName                           = "test-tenant"
	zoneName                             = "test-zone"
	extnetName                           = "some-extnet"
	numOfExternalNetworkCreatedByDefault = 2
)

func init() {

	apocheck.RegisterTest(apocheck.Test{
		Name:        "External Network Creation/Deletion Test",
		Description: "Test we can create/delete an external network",
		Author:      "Satyam",
		Tags:        []string{"suite:external-network"},
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

			en := extnetwork.ExternalNetwork{
				Account:     maniphttp.ExtractNamespace(t.PublicManipulator()),
				Zone:        zoneName + "-" + t.TestID(),
				Tenant:      tenantName,
				Name:        extnetName,
				Description: "some description",
				CIDRs:       []string{"10.0.0.0/8"},
				Ports:       []string{"80"},
				Protocols:   []string{"tcp"},
			}

			apocheck.Step(t, "Given I setup an external network", func() (err error) {

				err = en.Create(ctx, t.PublicManipulator())
				if err != nil {
					logger.LogError(fmt.Sprintln(err))
				}
				return err
			})

			apocheck.Step(t, "Then I get the created external network and verify the created details", func() (err error) {

				e, err := en.Get(ctx, t.PublicManipulator())
				if err != nil {
					logger.LogError(fmt.Sprintln(err))
					return err
				}
				if e == nil {
					logger.LogError(fmt.Sprintln("external network not found"))
					return errors.New("external network not found")
				}

				if e.Name != en.Name {
					logger.LogError(fmt.Sprintln("created external network name does not match"))
					return errors.New("created external network name does not match")
				}

				if len(en.Ports) == 1 && len(en.Protocols) == 1 && len(en.CIDRs) == 1 {
					if !testutil.CheckTagInTags(en.Ports[0], e.Ports) {
						logger.LogError(fmt.Sprintln("created external network port not found"))
						return errors.New("created external network port not found")
					}

					if !testutil.CheckTagInTags(en.Protocols[0], e.Protocols) {
						logger.LogError(fmt.Sprintln("created external network protocol not found"))
						return errors.New("created external network protocol not found")
					}

					if !testutil.CheckTagInTags(en.CIDRs[0], e.Entries) {
						logger.LogError(fmt.Sprintln("created external network CIDR not found"))
						return errors.New("created external network CIDR not found")
					}
				} else {
					logger.LogError(fmt.Sprintln("number of Ports, Protocols or CIDR is differenet than expected"))
					return errors.New("external network found")
				}

				return nil
			})

			apocheck.Step(t, "Then I remove the external network", func() (err error) {

				err = en.Delete(ctx, t.PublicManipulator())
				if err != nil {
					logger.LogError(fmt.Sprintln(err))
				}
				return err
			})

			apocheck.Step(t, "Then I get the deleted external network", func() (err error) {

				_, err = en.Get(ctx, t.PublicManipulator())
				if err == nil {
					logger.LogError(fmt.Sprintln("external network found"))
					return errors.New("external network found")
				}
				return nil
			})

			return nil
		},
	})
}
