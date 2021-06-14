package manipctx

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.aporeto.io/gaia"
	"go.aporeto.io/manipulate"
	"go.aporeto.io/manipulate/maniphttp"
	"go.aporeto.io/midgard-lib/client"
)

// Errors returned
var (
	ErrMissingCreds = errors.New("no creds path provided")
)

const (
	apiContextTimeout = 10 * time.Second
)

// InstallSIGINTHandler installs signal handlers for graceful shutdown.
func InstallSIGINTHandler(cancelFunc context.CancelFunc) {

	signalCh := make(chan os.Signal, 1)
	signal.Reset(syscall.SIGINT, syscall.SIGTERM)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signalCh
		cancelFunc()
		signal.Stop(signalCh)
		close(signalCh)
	}()
}

// APICACertPool prepares the API cert pool if not empty.
func APICACertPool(ca string) (*x509.CertPool, error) {

	caPool := x509.NewCertPool()
	if ca != "" {
		caPool.AppendCertsFromPEM([]byte(ca))
	}

	return caPool, nil
}

// Manipulator creates the manipulator used to process commands. Currently
// only HTTP manipulator is supported, and a constants.OptionTokenKey field is therefore required.
func Manipulator(ctx context.Context, credsPath string) (manipulate.Manipulator, error) {

	if credsPath == "" {
		return nil, ErrMissingCreds
	}

	data, err := ioutil.ReadFile(credsPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read credential file: %s", err)
	}

	var appCred *gaia.Credential
	appCred, tlsConfig, err := midgardclient.ParseCredentials(data)
	if err != nil {
		return nil, fmt.Errorf("unable to parse credential: %s", err)
	}

	token, err := midgardclient.
		NewClientWithTLS(appCred.APIURL, tlsConfig).
		IssueFromCertificate(context.Background(), 2*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("unable to get token from app creds: %s", err)
	}

	subctx, cancel := context.WithTimeout(ctx, apiContextTimeout)
	defer cancel()

	return maniphttp.New(
		subctx,
		appCred.APIURL,
		maniphttp.OptionNamespace(appCred.Namespace),
		maniphttp.OptionTLSConfig(tlsConfig),
		maniphttp.OptionToken(token),
	)
}
