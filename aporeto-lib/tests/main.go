package main

import (
	"log"

	"go.aporeto.io/apocheck"

	// Import all the test suites
	_ "github.com/PaloAltoNetworks/cns-customer/aporeto-lib/tests/authpolicy"
	_ "github.com/PaloAltoNetworks/cns-customer/aporeto-lib/tests/extnetwork"
	_ "github.com/PaloAltoNetworks/cns-customer/aporeto-lib/tests/hostservice"
	_ "github.com/PaloAltoNetworks/cns-customer/aporeto-lib/tests/networkpolicy"
	_ "github.com/PaloAltoNetworks/cns-customer/aporeto-lib/tests/oidc"
	_ "github.com/PaloAltoNetworks/cns-customer/aporeto-lib/tests/tenant"
	_ "github.com/PaloAltoNetworks/cns-customer/aporeto-lib/tests/zone"
)

func main() {

	// Run the command.
	cmd := apocheck.NewCommand("test", "integration tests for CNS customer APIs", "1.0")

	if err := cmd.Execute(); err != nil {
		log.Fatalf(err.Error())
	}
}
