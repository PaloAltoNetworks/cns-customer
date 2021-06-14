package authpolicy

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/auth"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/tenant"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/zone"
	"go.aporeto.io/apocheck"
	"go.aporeto.io/apotests-lib/logging"
	"go.aporeto.io/manipulate/maniphttp"
)

const (
	tenantName  = "test-tenant"
	zoneName    = "test-zone"
	authpolName = "tenant-ro-access"
)

func init() {

	apocheck.RegisterTest(apocheck.Test{
		Name:        "Tenant Read Only Auth Policy Creation/Deletion Test",
		Description: "Test we can create/delete a read only auth policy for tenant access",
		Author:      "Satyam",
		Tags:        []string{"suite:auth-policy"},
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

			ap := auth.Policy{
				Account:          maniphttp.ExtractNamespace(t.PublicManipulator()),
				Zone:             zoneName + "-" + t.TestID(),
				Tenant:           tenantName,
				Name:             authpolName,
				Description:      "some description",
				AuthPolicyClaims: [][]string{{"@auth:realm=claim", "@auth:test=authpolicy"}},
			}

			apocheck.Step(t, "Given I setup an auth policy for tenant ro access", func() (err error) {

				err = ap.Create(ctx, t.PublicManipulator())
				if err != nil {
					logger.LogError(fmt.Sprintln(err))
				}
				return err
			})

			apocheck.Step(t, "Then I get the created auth policy", func() (err error) {

				e, err := ap.Get(ctx, t.PublicManipulator())
				if err != nil {
					logger.LogError(fmt.Sprintln(err))
					return err
				}
				if e == nil {
					logger.LogError(fmt.Sprintln("auth policy not found"))
					return errors.New("auth policy not found")
				}
				return nil
			})

			apocheck.Step(t, "Then I remove the auth policy", func() (err error) {

				err = ap.Delete(ctx, t.PublicManipulator())
				if err != nil {
					logger.LogError(fmt.Sprintln(err))
				}
				return err
			})

			apocheck.Step(t, "Then I get the deleted auth policy", func() (err error) {

				_, err = ap.Get(ctx, t.PublicManipulator())
				if err == nil {
					logger.LogError(fmt.Sprintln("auth policy found"))
					return errors.New("auth policy found")
				}
				return nil
			})

			return nil
		},
	})
}
