package networkpolicies

import (
	"fmt"
	"strings"

	"github.com/PaloAltoNetworks/cns-customer/apoxfrm/libs/portranges"
	"github.com/PaloAltoNetworks/cns-customer/apoxfrm/libs/portspec"
	"github.com/PaloAltoNetworks/cns-customer/apoxfrm/libs/utils"
	"go.aporeto.io/gaia"
	"go.aporeto.io/gaia/protocols"
	"go.uber.org/zap"
)

func match(tags, objTags []string) bool {

	for _, t := range tags {
		found := false
		for _, n := range objTags {
			if n == t {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func refHasExtNextworks(ref []string) bool {

	for _, tag := range ref {
		if strings.HasPrefix(tag, utils.ExtnetPrefix) {
			return true
		} else if strings.HasPrefix(tag, utils.ExtnetNamePrefix) {
			return true
		}
	}
	return false
}

func refConvertTags(ref [][]string) [][]string {
	r := [][]string{}
	for _, obj := range ref {
		conv := refConvertExtNetworks(obj)
		r = append(r, conv)
	}
	return r
}

func refConvertExtNetworks(ref []string) []string {

	r := []string{}
	for _, tag := range ref {
		if strings.HasPrefix(tag, utils.ExtnetPrefix) {
			r = append(r, tag+utils.MigrationSuffix)
		} else if strings.HasPrefix(tag, utils.ExtnetNamePrefix) {
			r = append(r, tag+utils.MigrationSuffix)
		} else {
			r = append(r, tag)
		}
	}
	return r
}

// parseServicePort returns protocol and ports from servicePort.
func parseServicePort(servicePort string) (string, string, error) {

	if err := gaia.ValidateServicePort("servicePort", servicePort); err != nil {
		return "", "", err
	}

	parts := strings.SplitN(servicePort, "/", 2)
	protocol := parts[0]

	ports := ""

	if len(parts) == 2 {
		ports = parts[1]
	}

	return protocol, ports, nil
}

// extractProtocolsPorts is a helper function to extract ports for a given protocol from servicePorts.
func extractProtocolsPorts(protocol string, servicePorts []string, restrictedPortList []string) []string {

	ports := []string{}
	restrictedPortsMap := make(map[int]struct{})

	for _, restrictedPort := range restrictedPortList {
		rprotocol, rports, err := parseServicePort(restrictedPort)
		if err != nil {
			zap.L().Error("unable to parse restrictedPort", zap.Error(err))
			continue
		}

		if !strings.EqualFold(protocol, rprotocol) {
			continue
		}

		portSpec, err := portspec.NewPortSpecFromString(rports, nil)
		if err != nil {
			continue
		}

		for port := uint32(portSpec.Min); port <= uint32(portSpec.Max); port++ {
			restrictedPortsMap[int(port)] = struct{}{}
		}
	}

	for _, servicePort := range servicePorts {

		sprotocol, sports, err := parseServicePort(servicePort)
		if err != nil {
			zap.L().Error("unable to parse servicePort", zap.Error(err))
			continue
		}

		if !strings.EqualFold(sprotocol, protocols.L4ProtocolTCP) && !strings.EqualFold(sprotocol, protocols.L4ProtocolUDP) {
			continue
		}

		if !strings.EqualFold(sprotocol, protocol) {
			continue
		}

		lports, err := portranges.TrimPortRange(sports, restrictedPortsMap)
		if err != nil {
			continue
		}

		ports = append(ports, lports...)
	}

	return ports
}

func extractNonTCPAndUDPProtocols(a []string, b []string) []string {

	protoMap := make(map[string]string)

	for _, a1 := range a {
		rprotocol, rports, err := parseServicePort(a1)
		if err != nil {
			zap.L().Error("unable to parse restrictedPort", zap.Error(err))
			continue
		}

		if strings.EqualFold(protocols.L4ProtocolTCP, rprotocol) || strings.EqualFold(protocols.L4ProtocolUDP, rprotocol) {
			continue
		}

		protoMap[rprotocol] = rports
	}

	for _, b1 := range b {

		sprotocol, sports, err := parseServicePort(b1)
		if err != nil {
			zap.L().Error("unable to parse servicePort", zap.Error(err))
			continue
		}

		if strings.EqualFold(sprotocol, protocols.L4ProtocolTCP) && strings.EqualFold(sprotocol, protocols.L4ProtocolUDP) {
			continue
		}

		if v, ok := protoMap[sprotocol]; ok && len(v) < len(sports) {
			protoMap[sprotocol] = sports
		}
	}

	protos := []string{}
	for k, v := range protoMap {
		if len(v) > 0 {
			protos = append(protos, fmt.Sprintf("%s/%s", k, v))
		} else {
			protos = append(protos, k)
		}
	}

	return protos
}

func intersect(a []string, b []string) []string {

	set1 := a
	set2 := b

	// Not really intersection now is it?
	if len(a) == 0 {
		set1 = b
		set2 = a
	}

	i := []string{}
	y := extractProtocolsPorts(protocols.L4ProtocolTCP, set1, set2)
	for _, y1 := range y {

		i = append(i, fmt.Sprintf("%s/%s", protocols.L4ProtocolTCP, y1))
	}

	y = extractProtocolsPorts(protocols.L4ProtocolUDP, set1, set2)
	for _, y1 := range y {

		i = append(i, fmt.Sprintf("%s/%s", protocols.L4ProtocolUDP, y1))
	}

	y = extractNonTCPAndUDPProtocols(set1, set2)
	i = append(i, y...)
	return i
}

// equalSlices tells whether a and b contain the same elements.
func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for _, v := range a {
		found := false
		for _, w := range b {
			if v == w {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

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
