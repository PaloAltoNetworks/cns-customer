package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/hostservice"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/networkpolicy"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/tenant"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/zone"
	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/manipctx"
)

var scenarios []string

func init() {
	scenarios = []string{
		"zone-create",
		"zone-delete",
		"tenant-create",
		"tenant-disable",
		"tenant-delete",
		"service-create",
		"service-delete",
		"exception-create",
		"exception-delete",
	}
}

func usage() {
	fmt.Printf("Usage:\n  ac [-config <config-path>] -scenario <%s>\n", strings.Join(scenarios, "|"))
}

// Service definition.
type Service struct {
	Name       string   `json:"name"`
	Rail       string   `json:"rail"`
	Definition []string `json:"definition"`

	description string
}

// Policy defintion.
type Policy struct {
	Name          string   `json:"name"`
	SubjectTenant string   `json:"subject-tenant"`
	SubjectTags   []string `json:"subject-tags"`
	ObjectTenant  string   `json:"object-tenant"`
	ObjectTags    []string `json:"object-tags"`

	description string
}

// Aporeto is the configuration script.
type Aporeto struct {
	AppCredPath            string     `json:"app-cred-path"`
	Account                string     `json:"account"`
	Zone                   string     `json:"zone"`
	Tenant                 string     `json:"tenant"`
	TenantAuthPolicyClaims [][]string `json:"tenant-auth-policy-claims"`
	EnforcerAppCredPath    string     `json:"enforcer-app-cred-path"`
	Services               []Service  `json:"services"`
	ExceptionPolicies      []Policy   `json:"exception-policies"`

	zoneDescription             string
	tenantDescription           string
	tenantAuthPolicyDescription string
}

// Setup sets up computed fields.
func (a *Aporeto) Setup() {

	a.zoneDescription = "zone: " + a.Zone
	a.tenantDescription = a.zoneDescription + " tenant: " + a.Tenant
	a.tenantAuthPolicyDescription = a.tenantDescription + " read-only access using oidc claims"
	for i := range a.Services {
		a.Services[i].description = a.tenantDescription + " rail: " + a.Services[i].Rail + " service: " + a.Services[i].Name
	}
	for i := range a.ExceptionPolicies {
		a.ExceptionPolicies[i].description = a.tenantDescription + " purpose: " + a.ExceptionPolicies[i].Name
	}
}

func args() (*Aporeto, string) {

	configPtr := flag.String("config", "../config/tenant-a.json", "<config-path>")
	scenarioPtr := flag.String("scenario", "", strings.Join(scenarios, "|"))
	flag.Parse()

	if *configPtr == "" {
		usage()
		os.Exit(1)
	}

	jsonFile, err := os.Open(*configPtr)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Printf("Error: %s", err)
		os.Exit(1)
	}
	// read our opened xmlFile as a byte array.
	config, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		fmt.Printf("Error: %s", err)
		os.Exit(1)
	}

	if *scenarioPtr == "" {
		usage()
		os.Exit(1)
	} else {
		valid := false
		for _, s := range scenarios {
			if s == *scenarioPtr {
				valid = true
			}
		}
		if !valid {
			usage()
			os.Exit(1)
		}
	}

	var aporeto Aporeto
	json.Unmarshal(config, &aporeto)

	return &aporeto, *scenarioPtr
}

func main() {

	cfg, scenario := args()

	// Create Context and Install Signal Handlers
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	manipctx.InstallSIGINTHandler(cancel)

	// Utilize the application credential to get access to a manipulator.
	m, err := manipctx.Manipulator(ctx, cfg.AppCredPath)
	if err != nil {
		log.Printf("unable to prepare manipulator: %s\n", err.Error())
		os.Exit(1)
	}

	// Setup descriptions etc.
	cfg.Setup()

	switch scenario {
	case "zone-create":
		zone := zone.New(cfg.Account, cfg.Zone, cfg.zoneDescription)
		if err := zone.Create(ctx, m); err != nil {
			log.Printf("error: %s\n", err)
			os.Exit(1)
		}
	case "zone-delete":
		zone := zone.New(cfg.Account, cfg.Zone, cfg.zoneDescription)
		if err := zone.Delete(ctx, m); err != nil {
			log.Printf("error: %s\n", err)
			os.Exit(1)
		}
	case "tenant-create":
		tenant := tenant.Tenant{
			Account:               cfg.Account,
			Zone:                  cfg.Zone,
			Name:                  cfg.Tenant,
			Description:           cfg.tenantDescription,
			AuthPolicyClaims:      cfg.TenantAuthPolicyClaims,
			AuthPolicyDescription: cfg.tenantAuthPolicyDescription,
		}
		if err := tenant.Create(ctx, m); err != nil {
			log.Printf("error: %s\n", err)
			os.Exit(1)
		}
	case "tenant-disable":
		tenant := tenant.Tenant{
			Account: cfg.Account,
			Zone:    cfg.Zone,
			Name:    cfg.Tenant,
		}
		if err := tenant.Disable(ctx, m); err != nil {
			log.Printf("error: %s\n", err)
			os.Exit(1)
		}
	case "tenant-delete":
		tenant := tenant.Tenant{
			Account: cfg.Account,
			Zone:    cfg.Zone,
			Name:    cfg.Tenant,
		}
		if err := tenant.Delete(ctx, m); err != nil {
			log.Printf("error: %s\n", err)
			os.Exit(1)
		}
	case "service-create":
		for _, s := range cfg.Services {
			svc := hostservice.Service{
				Account:     cfg.Account,
				Zone:        cfg.Zone,
				Name:        cfg.Tenant,
				Rail:        s.Rail,
				Definition:  s.Definition,
				Description: s.description,
			}
			if err := svc.Create(ctx, m); err != nil {
				log.Printf("error: %s\n", err)
				os.Exit(1)
			}
		}
	case "service-delete":
		for _, s := range cfg.Services {
			svc := hostservice.Service{
				Account: cfg.Account,
				Zone:    cfg.Zone,
				Name:    cfg.Tenant,
				Rail:    s.Rail,
			}
			if err := svc.Delete(ctx, m); err != nil {
				log.Printf("error: %s\n", err)
				os.Exit(1)
			}
		}
	case "exception-create":
		for _, e := range cfg.ExceptionPolicies {
			np := networkpolicy.NetworkPolicy{
				Name:                   e.Name,
				Description:            e.description,
				SubjectTenantNamespace: e.SubjectTenant,
				SubjectTags:            e.SubjectTags,
				ObjectTenantNamespace:  e.ObjectTenant,
				ObjectTags:             e.ObjectTags,
			}
			if err := np.Create(ctx, m); err != nil {
				log.Printf("error: %s\n", err)
				os.Exit(1)
			}
		}
	case "exception-delete":
		for _, e := range cfg.ExceptionPolicies {
			np := networkpolicy.NetworkPolicy{
				Name:                   e.Name,
				Description:            e.description,
				SubjectTenantNamespace: e.SubjectTenant,
				SubjectTags:            e.SubjectTags,
				ObjectTenantNamespace:  e.ObjectTenant,
				ObjectTags:             e.ObjectTags,
			}
			if err := np.Delete(ctx, m); err != nil {
				log.Printf("error: %s\n", err)
				os.Exit(1)
			}
		}
	default:
		usage()
		panic("invalid scenario")
	}
}
