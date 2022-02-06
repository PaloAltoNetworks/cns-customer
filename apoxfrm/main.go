package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"

	"go.aporeto.io/gaia"
)

const extnetNamePrefix = "$name="
const migrationSuffix = "-v2"

var extnetPrefix string

func extnetsFromTags(policyNamespace string, tags []string, eList gaia.ExternalNetworksList) (extnetList gaia.ExternalNetworksList) {

	for _, e := range eList {

		if policyNamespace != e.Namespace {
			// Use external network if its defined in a hierarchy higher than policy namespace
			if !strings.HasPrefix(policyNamespace, e.Namespace+"/") {
				continue
			}
			// And the external network is propagated.
			if !e.Propagate {
				continue
			}
		}

		// The external network in either in same namespace as policy or at a higher level namespace with propagate=true
		if match(tags, e.NormalizedTags) {
			extnetList = append(extnetList, e)
		}
	}
	return
}

func xfrmNetPols(file string, netpols []map[string]interface{}, extnetList gaia.ExternalNetworksList) (xnetrulesetpolicies []map[string]interface{}) {

	xnetrulesetpolicies = make([]map[string]interface{}, 0)

	for _, n := range netpols {

		netpol := gaia.NewNetworkAccessPolicy()
		if err := mapstructure.Decode(n, netpol); err != nil {
			panic(err)
		}

		n, err := getNetPolInfo(netpol, extnetList)
		if err != nil {
			fmt.Println("    Error: " + err.Error())
		}

		zap.L().Info(
			"Network Policy",
			zap.String("file", file),
			zap.String("ns", netpol.Namespace),
			zap.String("name", netpol.Name),
			zap.Reflect("mode", netpol.ApplyPolicyMode),
			zap.Strings("ports", netpol.Ports),
			zap.Bool("propagate", netpol.Propagate),
			zap.Int("num-subjects", len(netpol.Subject)),
			zap.Bool("subject-networks", n.subjectHasExternalNetworks),
			zap.Int("num-objects", len(netpol.Object)),
			zap.Bool("object-networks", n.objectHasExternalNetworks),
			zap.Bool("subject-object-networks", n.subjectHasExternalNetworks && n.objectHasExternalNetworks),
		)

		xnetrulesetpolicies = append(xnetrulesetpolicies, n.transformations...)
	}

	return
}

func process(dir, file string, extraFiles []string) {

	location := filepath.Join(dir, file)
	fmt.Println("Processing file: " + location)

	inputData, err := os.ReadFile(location)
	if err != nil {
		panic(err)
	}

	exportedData := gaia.NewExport()
	if err := yaml.Unmarshal([]byte(inputData), exportedData); err != nil {
		panic(err)
	}

	extnets := exportedData.Data["externalnetworks"]
	netpols := exportedData.Data["networkaccesspolicies"]

	// Process extra external networks from the parent hierarchy
	var extraextnets []map[string]interface{}
	for _, file := range extraFiles {
		location := filepath.Join(dir, file)
		inputData, err := os.ReadFile(location)
		if err != nil {
			panic(err)
		}

		exportedData := gaia.NewExport()
		if err := yaml.Unmarshal([]byte(inputData), exportedData); err != nil {
			panic(err)
		}

		extraextnets = append(extraextnets, exportedData.Data["externalnetworks"]...)
	}

	gextnets, xextnets := xfrmExtNets(file, extnets, extraextnets)
	xnetrulesetpolicies := xfrmNetPols(file, netpols, gextnets)

	importData := gaia.NewImport()
	importData.Data.Label = exportedData.Label + migrationSuffix
	importData.Data.APIVersion = 1
	importData.Data.Data["externalnetworks"] = xextnets
	importData.Data.Data["networkrulesetpolicies"] = xnetrulesetpolicies

	data, err := yaml.Marshal(importData)
	if err != nil {
		panic(err)
	}

	location = filepath.Join(dir, "out-"+file)
	err = os.WriteFile(location, data, 0644)
	if err != nil {
		panic(err)
	}
}

func usage() {
	fmt.Println("apoxfrm -extnet-prefix customer:ext:name=")
}

func main() {

	prefix := flag.String("extnet-prefix", "", "prefix used in the tag to reference external networks")
	flag.Parse()

	if *prefix == "" {
		usage()
		os.Exit(1)
	}

	extnetPrefix = *prefix

	process("configs", "root.yaml", []string{})
	process("configs", "zone.yaml", []string{"root.yaml"})
	process("configs", "tenant-a.yaml", []string{"root.yaml", "zone.yaml"})
	process("configs", "tenant-b.yaml", []string{"root.yaml", "zone.yaml"})
	process("configs", "tenant-c.yaml", []string{"root.yaml", "zone.yaml"})
}
