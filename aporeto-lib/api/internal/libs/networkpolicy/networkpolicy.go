package networkpolicy

import (
	"context"
	"fmt"

	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/constants"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/internal/utils"
	"go.aporeto.io/elemental"
	"go.aporeto.io/gaia"
	"go.aporeto.io/manipulate"
)

// Create creates a network access policy
// Params:
//   - namespace: namespace where policy will be created.
//   - name: name of policy.
//   - description: description of policy.
//   - srctenantNamespace: namespace of src tenant using this policy (used for metadata only).
//   - dsttenantNamespace: namespace of dst tenant using this policy (used for metadata only).
//   - subject: tags for subject as 2d array. Every row is tags to match with AND clause. OR clause across rows.
//   - object: tags for object as 2d array. Every row is tags to match with AND clause. OR clause across rows.
//   - encrypt: should encryption be turned on.
func Create(
	ctx context.Context,
	m manipulate.Manipulator,
	namespace, name, description, srctenantNamespace, dsttenantNamespace string,
	subject, object [][]string,
	mode gaia.NetworkAccessPolicyApplyPolicyModeValue,
	action gaia.NetworkAccessPolicyActionValue,
	encrypt bool,
) error {

	// Ensure namespaces are correctly formatted.
	namespace = utils.SetupNamespaceString(namespace)

	// Setup a new network access policy.
	np := gaia.NewNetworkAccessPolicy()
	np.Name = name
	np.Description = description
	np.Subject = subject
	np.Object = object
	np.ApplyPolicyMode = mode
	np.Action = action
	np.EncryptionEnabled = encrypt
	np.LogsEnabled = true
	np.Propagate = true
	np.Metadata = utils.MakeTenantPairMetadata(srctenantNamespace, dsttenantNamespace)

	// Create a sub context so we dont retry too long.
	subctx, cancel := context.WithTimeout(ctx, constants.APIDefaultContextTimeout)
	defer cancel()

	// Create a namespace context where we are creating an object.
	mctx := manipulate.NewContext(
		subctx,
		manipulate.ContextOptionNamespace(namespace),
	)

	// Try creating multiple times in case of connection errors.
	return m.Create(mctx, np)
}

// Delete deletes a network access policy.
func Delete(ctx context.Context, m manipulate.Manipulator, namespace, name string) error {

	// Ensure namespaces are correctly formatted.
	namespace = utils.SetupNamespaceString(namespace)

	// Get matching network access policies.
	np, err := Get(ctx, m, namespace, name)
	if err != nil {
		return err
	}

	// Create a sub context so we dont retry too long.
	subctx, cancel := context.WithTimeout(ctx, constants.APIDefaultContextTimeout)
	defer cancel()

	// Create a namespace context where we are creating an object.
	mctx := manipulate.NewContext(
		subctx,
		manipulate.ContextOptionNamespace(utils.SetupNamespaceString(namespace)),
	)
	return m.Delete(mctx, np)
}

// DeleteManyWithMetadata deletes multiple network access policies
func DeleteManyWithMetadata(ctx context.Context, m manipulate.Manipulator, namespace, metadata string) error {

	// Ensure namespaces are correctly formatted.
	namespace = utils.SetupNamespaceString(namespace)

	nps := gaia.NetworkAccessPoliciesList{}

	subctx, cancel := context.WithTimeout(ctx, constants.APIDefaultContextTimeout)
	defer cancel()

	mctx := manipulate.NewContext(
		subctx,
		manipulate.ContextOptionFilter(
			elemental.NewFilterComposer().
				WithKey("namespace").Equals(utils.SetupNamespaceString(namespace)).
				WithKey("metadata").Contains(metadata).
				Done(),
		),
	)

	if err := m.RetrieveMany(mctx, &nps); err != nil {
		return err
	}

	var ret error
	for _, np := range nps {
		err := m.Delete(mctx, np)
		if err != nil {
			ret = err
		}
	}
	return ret
}

// Get fetches a list of network access policies matching the criteria.
func Get(ctx context.Context, m manipulate.Manipulator, namespace, name string) (*gaia.NetworkAccessPolicy, error) {

	nps := gaia.NetworkAccessPoliciesList{}

	subctx, cancel := context.WithTimeout(ctx, constants.APIDefaultContextTimeout)
	defer cancel()

	mctx := manipulate.NewContext(
		subctx,
		manipulate.ContextOptionNamespace(utils.SetupNamespaceString(namespace)),
		manipulate.ContextOptionFilter(
			elemental.NewFilterComposer().
				WithKey("name").Equals(name).
				Done(),
		),
	)

	if err := m.RetrieveMany(mctx, &nps); err != nil {
		return nil, err
	}
	if len(nps) == 0 {
		return nil, fmt.Errorf("no network access policy '%s' found in namespace '%s'", name, namespace)
	}
	if len(nps) > 1 {
		return nil, fmt.Errorf("multiple (%d) network access policies found with name '%s' in namespace '%s'", len(nps), name, namespace)
	}

	return nps[0], nil
}
