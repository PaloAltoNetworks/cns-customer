package hostservice

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/hostservice"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/tenant"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/zone"
	"go.aporeto.io/apocheck"
	"go.aporeto.io/apotests-lib/logging"
	"go.aporeto.io/manipulate/maniphttp"
)

const (
	tenantName  = "test-tenant"
	zoneName    = "test-zone"
	hostsvcName = "http"
)

func init() {

	apocheck.RegisterTest(apocheck.Test{
		Name:        "Host Service Creation/Deletion Test",
		Description: "Test we can create/delete an host service",
		Author:      "Satyam",
		Tags:        []string{"suite:host-service"},
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

			hs := hostservice.Service{
				Account:         maniphttp.ExtractNamespace(t.PublicManipulator()),
				Zone:            zoneName + "-" + t.TestID(),
				Tenant:          tenantName,
				Rail:            "public",
				Name:            hostsvcName,
				Description:     "some description",
				Definition:      []string{"tcp/80"},
				HostModeEnabled: false,
			}

			apocheck.Step(t, "Given I setup an host service", func() (err error) {

				err = hs.Create(ctx, t.PublicManipulator())
				if err != nil {
					logger.LogError(fmt.Sprintln(err))
				}
				return err
			})

			apocheck.Step(t, "Then I get the created host service", func() (err error) {

				e, err := hs.Get(ctx, t.PublicManipulator())
				if err != nil {
					logger.LogError(fmt.Sprintln(err))
					return err
				}
				if e == nil {
					logger.LogError(fmt.Sprintln("host service not found"))
					return errors.New("host service not found")
				}
				return nil
			})

			apocheck.Step(t, "Then I remove the host service", func() (err error) {

				err = hs.Delete(ctx, t.PublicManipulator())
				if err != nil {
					logger.LogError(fmt.Sprintln(err))
				}
				return err
			})

			apocheck.Step(t, "Then I get the deleted host service", func() (err error) {

				_, err = hs.Get(ctx, t.PublicManipulator())
				if err == nil {
					logger.LogError(fmt.Sprintln("host service found"))
					return errors.New("host service found")
				}
				return nil
			})

			return nil
		},
	})
}
