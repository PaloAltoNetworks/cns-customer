package appcred

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"

	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/constants"
	"go.aporeto.io/elemental"
	"go.aporeto.io/gaia"
	"go.aporeto.io/manipulate"
	"go.aporeto.io/tg/tglib"
)

// Create creates a new *gaia.AppCredential.
func Create(ctx context.Context, m manipulate.Manipulator, parentNamespace, name, description string, roles []string) ([]byte, error) {

	creds := gaia.NewAppCredential()
	creds.Name = name
	creds.Description = description
	creds.Roles = roles

	if err := m.Create(
		manipulate.NewContext(
			ctx,
			manipulate.ContextOptionNamespace(parentNamespace),
		),
		creds,
	); err != nil {
		return nil, err
	}

	creds, err := Renew(ctx, m, creds)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(creds.Credentials, "", "    ")
}

// Delete deletes an application credential
func Delete(ctx context.Context, m manipulate.Manipulator, parentNamespace, name string) error {

	// Get matching application credentials.
	ac, err := Get(ctx, m, parentNamespace, name)
	if err != nil {
		return err
	}

	// Create a sub context so we dont retry too long.
	subctx, cancel := context.WithTimeout(ctx, constants.APIDefaultContextTimeout)
	defer cancel()

	// Create a namespace context where we are creating an object.
	mctx := manipulate.NewContext(
		subctx,
		manipulate.ContextOptionNamespace(parentNamespace),
	)
	return m.Delete(mctx, ac)
}

// Get fetches a list of application credentials matching the criteria.
func Get(ctx context.Context, m manipulate.Manipulator, parentNamespace, name string) (*gaia.AppCredential, error) {

	acs := gaia.AppCredentialsList{}

	subctx, cancel := context.WithTimeout(ctx, constants.APIDefaultContextTimeout)
	defer cancel()

	mctx := manipulate.NewContext(
		subctx,
		manipulate.ContextOptionNamespace(parentNamespace),
		manipulate.ContextOptionFilter(
			elemental.NewFilter().
				WithKey("name").Equals(name).
				Done(),
		),
	)

	if err := m.RetrieveMany(mctx, &acs); err != nil {
		return nil, err
	}
	if len(acs) == 0 {
		return nil, fmt.Errorf("no application credential '%s' found in namespace '%s'", name, parentNamespace)
	}
	if len(acs) > 1 {
		return nil, fmt.Errorf("multiple (%d) application credentials found with name '%s' in namespace '%s'", len(acs), name, parentNamespace)
	}

	return acs[0], nil
}

// Renew renews the given application credential.
func Renew(ctx context.Context, m manipulate.Manipulator, creds *gaia.AppCredential) (*gaia.AppCredential, error) {

	// Generate a private key and a CSR from the application credential info.
	csr, pk, err := makeCSR(creds.Name, creds.ID, creds.Namespace)
	if err != nil {
		return nil, err
	}

	// Update the application credential with the csr
	creds.CSR = string(csr)

	if err = m.Update(
		manipulate.NewContext(
			ctx,
			manipulate.ContextOptionNamespace(creds.Namespace),
		),
		creds,
	); err != nil {
		return nil, err
	}

	// Write the private key in the application credential.
	creds.Credentials.CertificateKey = base64.StdEncoding.EncodeToString(pk)

	return creds, nil
}

func makeCSR(name string, id string, namespace string) (csr []byte, key []byte, err error) {

	pk, err := tglib.ECPrivateKeyGenerator()
	if err != nil {
		return nil, nil, err
	}

	csr, err = tglib.GenerateSimpleCSR([]string{namespace}, nil, "app:credential:"+id+":"+name, nil, pk)
	if err != nil {
		return nil, nil, err
	}

	keyBlock, err := tglib.KeyToPEM(pk)
	if err != nil {
		return nil, nil, err
	}

	return csr, pem.EncodeToMemory(keyBlock), nil
}
