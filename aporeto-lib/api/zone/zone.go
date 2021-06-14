package zone

import (
	"context"

	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/internal/libs/namespace"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/internal/utils"
	"go.aporeto.io/manipulate"
)

// Zone definition.
type Zone struct {
	Account     string `json:"account"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// New sets up a new zone.
func New(account, name, description string) *Zone {
	return &Zone{
		Account:     account,
		Name:        name,
		Description: description,
	}
}

// Create is an implementation of how to create a zone namespace.
func (z *Zone) Create(ctx context.Context, m manipulate.Manipulator) error {

	accountNamespace := utils.SetupNamespaceString(z.Account)
	return namespace.Create(ctx, m, accountNamespace, z.Name, z.Description)
}

// Delete is an implementation of how to delete a zone namespace.
func (z *Zone) Delete(ctx context.Context, m manipulate.Manipulator) error {

	accountNamespace := utils.SetupNamespaceString(z.Account)
	return namespace.Delete(ctx, m, accountNamespace, z.Name)
}
