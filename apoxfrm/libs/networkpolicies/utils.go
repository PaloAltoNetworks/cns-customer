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

func match(tags []string, e *gaia.ExternalNetwork) bool {

	for _, t := range tags {

		if t == "$identity=externalnetwork" {
			continue
		} else if strings.HasPrefix(t, "$namespace=") {
			continue
		} else if strings.HasPrefix(t, "$name=") {
			name := strings.SplitN(t, "=", 2)
			if name[1] != e.Name {
				return false
			}
		} else {
			found := false
			for _, n := range e.AssociatedTags {
				if n == t {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}
	return true
}

func refHasBadNames(ref [][]string) bool {

	for _, tags := range ref {
		namePrefix := false
		identityPrefix := false
		for _, tag := range tags {
			if strings.HasPrefix(tag, utils.ExtnetNamePrefix) {
				namePrefix = true
			}
			if strings.HasPrefix(tag, "$identity=") {
				identityPrefix = true
			}
		}

		if namePrefix {
			if !identityPrefix {
				return true
			}
		}
	}
	return false
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

	protoMap := make(map[string]map[string]interface{})

	for _, a1 := range a {
		rprotocol, rports, err := parseServicePort(a1)
		if err != nil {
			zap.L().Error("unable to parse restrictedPort", zap.Error(err))
			continue
		}

		if strings.EqualFold(protocols.L4ProtocolTCP, rprotocol) || strings.EqualFold(protocols.L4ProtocolUDP, rprotocol) {
			continue
		}

		_, ok := protoMap[rprotocol]
		if !ok {
			protoMap[rprotocol] = make(map[string]interface{})
		}
		if !strings.Contains(rports, "/") {
			if len(protoMap[rprotocol]) > 0 {
				continue
			}
		} else {
			code := strings.Split(rports, "/")[0]
			delete(protoMap[rprotocol], code)
		}
		if rports != "" {
			protoMap[rprotocol][rports] = nil
		}
	}

	for _, b1 := range b {

		sprotocol, sports, err := parseServicePort(b1)
		if err != nil {
			zap.L().Error("unable to parse servicePort", zap.Error(err))
			continue
		}

		if strings.EqualFold(sprotocol, protocols.L4ProtocolTCP) || strings.EqualFold(sprotocol, protocols.L4ProtocolUDP) {
			continue
		}

		_, ok := protoMap[sprotocol]
		if !ok {
			protoMap[sprotocol] = make(map[string]interface{})
		}
		if !strings.Contains(sports, "/") {
			if len(protoMap[sprotocol]) > 0 {
				continue
			}
		} else {
			code := strings.Split(sports, "/")[0]
			delete(protoMap[sprotocol], code)
		}
		if sports != "" {
			protoMap[sprotocol][sports] = nil
		}
	}

	protos := []string{}
	for k, v := range protoMap {
		if len(v) > 0 {
			for k1 := range v {
				protos = append(protos, fmt.Sprintf("%s/%s", k, k1))
			}
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

	// Condition tcp and udp protocols
	for i, pp := range set1 {
		if strings.EqualFold(pp, protocols.L4ProtocolTCP) || strings.EqualFold(pp, protocols.L4ProtocolUDP) {
			set1[i] = pp + "/1:65535"
		}
	}
	for i, pp := range set2 {
		if strings.EqualFold(pp, protocols.L4ProtocolTCP) || strings.EqualFold(pp, protocols.L4ProtocolUDP) {
			set2[i] = pp + "/1:65535"
		}
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
		if match(tags, e) {
			extnetList = append(extnetList, e)
		}
	}
	return
}
