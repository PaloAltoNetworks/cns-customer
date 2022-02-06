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

func xfrmNetPols(file string, netpols []map[string]interface{}, extnetList gaia.ExternalNetworksList) (xnetrulesetpolicies []map[string]interface{}) {

	xnetrulesetpolicies = make([]map[string]interface{}, 0)

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
			zap.String("file", file),
			zap.String("ns", netpol.Namespace),
			zap.String("name", netpol.Name),
			zap.Reflect("mode", netpol.ApplyPolicyMode),
			zap.Strings("ports", netpol.Ports),
			zap.Bool("propagate", netpol.Propagate),
			zap.Int("num-subjects", len(netpol.Subject)),
			zap.Int("num-objects", len(netpol.Object)),
		)

		xnetrulesetpolicies = append(xnetrulesetpolicies, transformations...)
	}

	return
}

func xfrmExtNets(file string, extnets, extraextnets []map[string]interface{}) (extnetList gaia.ExternalNetworksList, xextnets []map[string]interface{}) {

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
			zap.String("file", file),
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
	importData.Data.Label = exportedData.Label + utils.MigrationSuffix
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

	utils.ExtnetPrefix = *prefix

	process("configs", "root.yaml", []string{})
	process("configs", "zone.yaml", []string{"root.yaml"})
	process("configs", "tenant-a.yaml", []string{"root.yaml", "zone.yaml"})
	process("configs", "tenant-b.yaml", []string{"root.yaml", "zone.yaml"})
	process("configs", "tenant-c.yaml", []string{"root.yaml", "zone.yaml"})
}
