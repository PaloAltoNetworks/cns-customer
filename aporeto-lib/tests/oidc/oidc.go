package oidc

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/auth"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/oidc"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/tenant"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/zone"
	"go.aporeto.io/apocheck"
	"go.aporeto.io/apotests-lib/logging"
	"go.aporeto.io/manipulate/maniphttp"
)

const (
	tenantName       = "test-tenant"
	zoneName         = "test-zone"
	authpolName      = "group-based-access"
	oidcproviderName = "AzureAD"
)

func init() {

	apocheck.RegisterTest(apocheck.Test{
		Name:        "Tenant OIDC Test",
		Description: "Test we can create/delete an OIDC provider config",
		Author:      "Pritesh",
		Tags:        []string{"suite:oidc"},
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

			ap := auth.Policy{
				Account:          maniphttp.ExtractNamespace(t.PublicManipulator()),
				Zone:             zoneName + "-" + t.TestID(),
				Tenant:           tenantName,
				Name:             authpolName,
				Description:      "This auth policy allows the oidc authenticated user belonging to a group as a viewer",
				AuthPolicyClaims: [][]string{{"@auth:realm=oidc", "@auth:groups:8d094467-1e1d-4135-b90f-83425a67d44d=true"}},
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

			apocheck.Step(t, "Given I setup an auth policy for tenant to access", func() (err error) {

				err = ap.Create(ctx, t.PublicManipulator())
				if err != nil {
					logger.LogError(fmt.Sprintln(err))
				}
				return err
			})

			return nil, func() {

				// Delete the auth policy
				if err := ap.Delete(ctx, t.PublicManipulator()); err != nil {
					logger.LogError(fmt.Sprintln(err))
				}

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

			op := oidc.OIDC{
				Account:      maniphttp.ExtractNamespace(t.PublicManipulator()),
				Zone:         zoneName + "-" + t.TestID(),
				Tenant:       tenantName,
				Name:         oidcproviderName,
				Endpoint:     "https://sts.windows.net/91d22d90-334e-4dc7-854c-9afd1da4fe21/",
				ClientID:     "e65d717f-fb81-4cb9-8b39-27e3a39bbc1a",
				ClientSecret: "h1KQX]z.9p0Ctk+DFJ:5E0qi?8CA/n8a",
				Scopes:       []string{},
				Default:      true,
				Subjects:     []string{},
			}

			apocheck.Step(t, "Given I setup a OIDC provider for accessing Azure Active Directory", func() (err error) {

				err = op.Create(ctx, t.PublicManipulator())
				if err != nil {
					logger.LogError(fmt.Sprintln(err))
				}
				return err
			})

			apocheck.Step(t, "Then I get the OIDC provider configuration", func() (err error) {

				e, err := op.Get(ctx, t.PublicManipulator())
				if err != nil {
					logger.LogError(fmt.Sprintln(err))
					return err
				}
				if e == nil {
					logger.LogError(fmt.Sprintln("OIDC provider not found"))
					return errors.New("OIDC provider not found")
				}
				return nil
			})

			// FIX THIS: Perform login - user and password is args. search for viper.GetString
			// FIX THIS: Login as root

			apocheck.Step(t, "Then I delete the OIDC provider", func() (err error) {

				err = op.Delete(ctx, t.PublicManipulator())
				if err != nil {
					logger.LogError(fmt.Sprintln(err))
				}
				return err
			})

			apocheck.Step(t, "Then I Get the deleted OIDC provider", func() (err error) {

				_, err = op.Get(ctx, t.PublicManipulator())
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
