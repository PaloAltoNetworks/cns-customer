package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/PaloAltoNetworks/cns-customer/apoxfrm/libs/externalnetwork"
	"github.com/PaloAltoNetworks/cns-customer/apoxfrm/libs/networkpolicies"
	"github.com/PaloAltoNetworks/cns-customer/apoxfrm/libs/utils"
	"github.com/ghodss/yaml"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"

	"go.aporeto.io/gaia"
)

// xfrmNetPols transforms the network access policies into new policies called network ruleset policies.
//
// Arguments:
// - netpols: network policies
// - extnetList: external networks (from complete ns hierarcy that may be needed to resolve these policies)
//
// Returns:
// - netrulesetpolicies: network rule set policies.
//
// The key things to observe w.r.t. new policies are:
// - Bidirectional policy mode not supported.
// - One network rule set policy is applied to a set of processing units and contains both ingress and egress rules for this set of processing units.
// - Multiple network rule set policies can still be applied to the same set of processing units.
// - Port matching is a part of incoming and outgoing rules.
//
func xfrmNetPols(netpols []map[string]interface{}, extnetList gaia.ExternalNetworksList) (netrulesetpolicies []map[string]interface{}) {

	netrulesetpolicies = make([]map[string]interface{}, 0)

	for _, n := range netpols {

		netpol := gaia.NewNetworkAccessPolicy()
		if err := mapstructure.Decode(n, netpol); err != nil {
			panic(err)
		}

		transformations, err := networkpolicies.Get(netpol, extnetList)
		if err != nil {
			fmt.Println("    Error: " + err.Error())
		}

		zap.L().Info(
			"Network Policy",
			zap.String("ns", netpol.Namespace),
			zap.String("name", netpol.Name),
			zap.Reflect("mode", netpol.ApplyPolicyMode),
			zap.Strings("ports", netpol.Ports),
			zap.Bool("propagate", netpol.Propagate),
			zap.Int("num-subjects", len(netpol.Subject)),
			zap.Int("num-objects", len(netpol.Object)),
		)

		netrulesetpolicies = append(netrulesetpolicies, transformations...)
	}

	return
}

// xfrmExtNets transforms the external networks to a v2 model.
// The key thing here is an external network can not define protocols and ports in the external network definition.
//
// Arguments:
// - extnets: external networks that are in this ns level and will need to be reimported.
// - extraextnets: external networks that are in the higher ns level if any. these will not be generated in file to import.
//
// Returns:
// - extnetList: list of external networks which will be used to resolve policies.
// - xextnets: transformed external networks that will need to be added to import files.
//
func xfrmExtNets(extnets, extraextnets []map[string]interface{}) (extnetList gaia.ExternalNetworksList, xextnets []map[string]interface{}) {

	for i, e := range append(extraextnets, extnets...) {

		extnet, err := externalnetwork.Decode(e)
		if err != nil {
			panic("error in external network: " + err.Error())
		}

		// create a global list that can be used in network policies
		extnetList = append(extnetList, extnet)

		// Process the external network - Create a v2 copy, add suffix to name, remove protocol and ports
		v2extnet := externalnetwork.Transform(extnet)

		zap.L().Info(
			"External Network",
			zap.String("ns", extnet.Namespace),
			zap.String("name", extnet.Name),
			zap.Strings("ports", extnet.ServicePorts),
		)

		// Dont export extra external networks
		if i >= len(extraextnets) {
			xe, err := externalnetwork.Encode(v2extnet)
			if err != nil {
				panic("error in external network: " + err.Error())
			}
			xextnets = append(xextnets, xe)
		}
	}
	return
}

func process(dir, file string, extraFiles []string) {

	location := filepath.Join(dir, file)

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

	gextnets, xextnets := xfrmExtNets(extnets, extraextnets)
	netrulesetpolicies := xfrmNetPols(netpols, gextnets)

	importData := gaia.NewImport()
	importData.Data.Label = exportedData.Label + utils.MigrationSuffix
	importData.Data.APIVersion = 1
	importData.Data.Data["externalnetworks"] = xextnets
	importData.Data.Data["networkrulesetpolicies"] = netrulesetpolicies

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
	fmt.Println("apoxfrm -extnet-prefix customer:ext:name= -config-dir <directory> -config-file <yaml-file> [-extra-files <yaml-file1> <yaml-file2> ...]")
	fmt.Println("examples:")
	fmt.Println("  apoxfrm -extnet-prefix customer:ext:name= -config-dir configs -config-file zone.yaml -extra-files root.yaml")
	fmt.Println("  apoxfrm -extnet-prefix customer:ext:name= -config-dir configs -config-file tenant.yaml -extra-files root.yaml zone.yaml")
}

func main() {

	var extraFiles arrayFlags
	directory := flag.String("config-dir", "configs", "configuation directory for yaml files")
	file := flag.String("config-file", "root.yaml", "yaml configuation file")
	flag.Var(&extraFiles, "extra-files", "additional files needed to resolve extra external networks.")
	prefix := flag.String("extnet-prefix", "", "prefix used in the tag to reference external networks")
	flag.Parse()

	if *prefix == "" {
		usage()
		os.Exit(1)
	}

	utils.ExtnetPrefix = *prefix

	location := filepath.Join(*directory, *file)
	fmt.Println("External network prefix: " + *prefix)
	fmt.Println("Processing file:         " + location)
	fmt.Println("Additional files:        " + extraFiles.String())

	process(*directory, *file, extraFiles)
}
