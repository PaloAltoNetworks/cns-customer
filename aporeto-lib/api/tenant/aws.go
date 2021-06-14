package tenant

import (
	"context"
	"fmt"

	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/constants"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/internal/utils"
	"go.aporeto.io/gaia"
	"go.aporeto.io/manipulate"
)

// CreateAWSAutoRegistrationAuth creates an Api Authorization Policy specifically to allow for enforcers to register via AWSAutoRegistration.
func CreateAWSAutoRegistrationAuth(ctx context.Context, m manipulate.Manipulator, namespace, authorizedNamespace, description string, claims []string) error {

	// Ensure Namespaces are correctly formatted.
	namespace = utils.SetupNamespaceString(namespace)
	authorizedNamespace = utils.SetupNamespaceString(authorizedNamespace)

	// Setup a new authorization policy to allow enforcers to register with AWS security tokens
	ap := gaia.NewAPIAuthorizationPolicy()
	ap.Name = fmt.Sprintf("aws enforcer auto registration - rail '%s'", authorizedNamespace)
	ap.Description = description
	ap.Subject = [][]string{
		{
			"@auth:realm=awssecuritytoken",
		},
	}
	ap.Subject[0] = append(ap.Subject[0], claims...)
	ap.AuthorizedIdentities = []string{constants.AuthEnforcerd}
	ap.AuthorizedNamespace = authorizedNamespace
	ap.PropagationHidden = false
	ap.Metadata = utils.MakeTenantMetadata(authorizedNamespace)

	// Create a sub context so we dont retry too long.
	subctx, cancel := context.WithTimeout(ctx, constants.APIDefaultContextTimeout)
	defer cancel()

	// Create a namespace context where we are creating an object.
	mctx := manipulate.NewContext(
		subctx,
		manipulate.ContextOptionNamespace(namespace),
	)

	// Try creating multiple times in case of connection errors.
	return m.Create(mctx, ap)
}
