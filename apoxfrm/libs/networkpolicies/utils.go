package networkpolicies

import (
	"fmt"
	"sort"
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

func appendUnique(a []string, b string) []string {

	found := false
	for _, val := range a {
		if val == b {
			found = true
			break
		}
	}

	if !found {
		a = append(a, b)
	}

	return a
}

func extractNonTCPAndUDPProtocols(a []string, b []string) []string {

	sort.Strings(a)
	sort.Strings(b)

	protoMap := make(map[string]map[string][]string)

	for _, a1 := range a {
		rprotocol, rports, err := parseServicePort(a1)
		if err != nil {
			zap.L().Error("unable to parse restrictedPort", zap.Error(err))
			continue
		}

		if strings.EqualFold(protocols.L4ProtocolTCP, rprotocol) || strings.EqualFold(protocols.L4ProtocolUDP, rprotocol) {
			continue
		}

		codeMap, ok := protoMap[rprotocol]
		if !ok {
			codeMap = make(map[string][]string)
		}

		if rports == "" {
			codeMap = nil
		}

		if codeMap == nil {
			protoMap[rprotocol] = codeMap
			continue
		}

		if !strings.Contains(rports, "/") {
			codeMap[rports] = nil
			protoMap[rprotocol] = codeMap
			continue
		}

		parts := strings.Split(rports, "/")
		code := parts[0]
		codeTypeList, ok := codeMap[code]
		if !ok {
			codeTypeList = []string{}
		}

		if codeTypeList == nil {
			protoMap[rprotocol] = codeMap
			continue
		}

		codeMap[code] = appendUnique(codeTypeList, parts[1])
		protoMap[rprotocol] = codeMap
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

		codeMap, ok := protoMap[sprotocol]
		if !ok {
			codeMap = make(map[string][]string)
		}

		if sports == "" {
			codeMap = nil
		}

		if codeMap == nil {
			protoMap[sprotocol] = codeMap
			continue
		}

		if !strings.Contains(sports, "/") {
			codeMap[sports] = nil
			protoMap[sprotocol] = codeMap
			continue
		}

		parts := strings.Split(sports, "/")
		code := parts[0]
		codeTypeList, ok := codeMap[code]
		if !ok {
			codeTypeList = []string{}
		}

		if codeTypeList == nil {
			protoMap[sprotocol] = codeMap
			continue
		}

		codeMap[code] = appendUnique(codeTypeList, parts[1])
		protoMap[sprotocol] = codeMap
	}

	protos := []string{}
	for k, v := range protoMap {
		if v == nil {
			protos = append(protos, k)
			continue
		}

		for k1, v1 := range v {
			if v1 == nil {
				protos = append(protos, fmt.Sprintf("%s/%s", k, k1))
				continue
			}

			for _, v2 := range v1 {
				protos = append(protos, fmt.Sprintf("%s/%s/%s", k, k1, v2))
				continue
			}
		}
	}

	return protos
}

func intersect(a []string, b []string) []string {

	// Condition tcp and udp protocols and remove any
	set1 := make([]string, 0)
	for _, pp := range a {
		if strings.EqualFold(pp, protocols.L4ProtocolTCP) || strings.EqualFold(pp, protocols.L4ProtocolUDP) {
			set1 = append(set1, pp+"/1:65535")
		} else if !strings.EqualFold(pp, protocols.ANY) {
			set1 = append(set1, pp)
		} else {
			// If we find any, just reset the set
			set1 = make([]string, 0)
			break
		}
	}
	set2 := make([]string, 0)
	for _, pp := range b {
		if strings.EqualFold(pp, protocols.L4ProtocolTCP) || strings.EqualFold(pp, protocols.L4ProtocolUDP) {
			set2 = append(set2, pp+"/1:65535")
		} else if !strings.EqualFold(pp, protocols.ANY) {
			set2 = append(set2, pp)
		} else {
			// If we find any, just reset the set
			set2 = make([]string, 0)
			break
		}
	}

	// Setup sets to process
	cset1 := set1
	cset2 := set2

	// Not really intersection now is it?
	if len(set1) == 0 {
		cset1 = set2
		cset2 = set1
	}

	i := []string{}
	y := extractProtocolsPorts(protocols.L4ProtocolTCP, cset1, cset2)
	for _, y1 := range y {

		i = append(i, fmt.Sprintf("%s/%s", protocols.L4ProtocolTCP, y1))
	}

	y = extractProtocolsPorts(protocols.L4ProtocolUDP, cset1, cset2)
	for _, y1 := range y {

		i = append(i, fmt.Sprintf("%s/%s", protocols.L4ProtocolUDP, y1))
	}

	y = extractNonTCPAndUDPProtocols(cset1, cset2)
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
