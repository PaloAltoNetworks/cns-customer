package zone

import (
	"context"
	"fmt"
	"os"

	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/zone"
	"go.aporeto.io/apocheck"
	"go.aporeto.io/apotests-lib/logging"
	"go.aporeto.io/manipulate/maniphttp"
)

const (
	zoneName = "test-zone"
)

func init() {

	apocheck.RegisterTest(apocheck.Test{
		Name:        "Zone Creation/Deletion Test",
		Description: "Test to ensure zone creation/deletion is working correctly",
		Author:      "Satyam",
		Tags:        []string{"suite:zone", "zone"},
		Setup: func(ctx context.Context, t apocheck.TestInfo) (interface{}, apocheck.TearDownFunction, error) {
			return nil, func() {}, nil
		},

		Function: func(ctx context.Context, t apocheck.TestInfo) error {

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

			apocheck.Step(t, "Then I delete a zone on aporeto", func() (err error) {

				err = zn.Delete(ctx, t.PublicManipulator())
				if err != nil {
					logger.LogError(fmt.Sprintln(err))
				}
				return err
			})

			return nil
		},
	})
}
