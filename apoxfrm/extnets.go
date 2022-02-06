package main

import (
	"go.uber.org/zap"

	"github.com/PaloAltoNetworks/cns-customer/apoxfrm/libs/externalnetwork"
	"go.aporeto.io/gaia"
)

func xfrmExtNets(file string, extnets, extraextnets []map[string]interface{}) (extnetList gaia.ExternalNetworksList, xextnets []map[string]interface{}) {

	for i, e := range append(extraextnets, extnets...) {

		extnet, err := externalnetwork.Decode(e)
		if err != nil {
			panic("error in external network: " + err.Error())
		}

		// create a global list that can be used in network policies
		extnetList = append(extnetList, extnet)

		// Process the external network - Create a v2 copy, add suffix to name, remove protocol and ports
		v2extnet := externalnetwork.Transform(extnet, migrationSuffix, extnetPrefix)

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
