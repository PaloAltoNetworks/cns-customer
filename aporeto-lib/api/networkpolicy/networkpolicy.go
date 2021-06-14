package networkpolicy

import (
	"context"

	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/internal/libs/networkpolicy"
	"go.aporeto.io/gaia"
	"go.aporeto.io/manipulate"
)

// NetworkPolicy defintion.
type NetworkPolicy struct {
	Namespace              string   `json:"namespace"`
	Name                   string   `json:"name"`
	Description            string   `json:"description"`
	SubjectTenantNamespace string   `json:"subject-tenant-namespace"`
	SubjectTags            []string `json:"subject-tags"`
	ObjectTenantNamespace  string   `json:"object-tenant-namespace"`
	ObjectTags             []string `json:"object-tags"`
	PolicyMode             string   `json:"policy-mode"`
}

// Create is an implementation of how to create an network policy.
func (n *NetworkPolicy) Create(ctx context.Context, m manipulate.Manipulator) error {

	return networkpolicy.Create(
		ctx,
		m,
		n.Namespace,
		n.Name,
		n.Description,
		n.SubjectTenantNamespace,
		n.ObjectTenantNamespace,
		[][]string{n.SubjectTags},
		[][]string{n.ObjectTags},
		gaia.NetworkAccessPolicyApplyPolicyModeValue(n.PolicyMode),
		gaia.NetworkAccessPolicyActionAllow,
		false,
	)
}

// Delete is an implementation of how to delete an network policy.
func (n *NetworkPolicy) Delete(ctx context.Context, m manipulate.Manipulator) error {

	return networkpolicy.Delete(
		ctx,
		m,
		n.Namespace,
		n.Name,
	)
}

// Get fetches a list of external networks matching the criteria.
func (n *NetworkPolicy) Get(ctx context.Context, m manipulate.Manipulator) (*gaia.NetworkAccessPolicy, error) {

	return networkpolicy.Get(
		ctx,
		m,
		n.Namespace,
		n.Name,
	)
}
