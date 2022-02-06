package main

import (
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"go.aporeto.io/gaia"
)

type netPolInfo struct {
	// Network Policy
	netpol *gaia.NetworkAccessPolicy

	// External References
	subjectHasExternalNetworks           bool
	allSubjectsReferenceExternalNetworks bool
	subjectExternalNetworks              []gaia.ExternalNetworksList
	subjectProtocolPorts                 [][]string
	objectHasExternalNetworks            bool
	allObjectsReferenceExternalNetworks  bool
	objectExternalNetworks               []gaia.ExternalNetworksList
	objectProtocolPorts                  [][]string

	outgoing     *gaia.NetworkRuleSetPolicy
	outgoingRule *gaia.NetworkRule
	incoming     *gaia.NetworkRuleSetPolicy
	incomingRule *gaia.NetworkRule

	// Transformations
	transformations []map[string]interface{}

	// Warnings
	candidateForUnidirectionalPolicy                bool
	ineffectivePolicy                               bool
	subjectExternalNetworksNoNameRefButExtNetsFound []bool
	subjectExternalNetworksPortMigrationNotPossible []bool
	objectExternalNetworksNoNameRefButExtNetsFound  []bool
	objectExternalNetworksPortMigrationNotPossible  []bool

	// Exceptions detected
	negationsNotSupported bool
	exceptions            bool
}

func newNetPolInfo(netpol *gaia.NetworkAccessPolicy) *netPolInfo {

	return &netPolInfo{
		netpol:                               netpol,
		subjectExternalNetworks:              make([]gaia.ExternalNetworksList, len(netpol.Subject)),
		subjectProtocolPorts:                 make([][]string, len(netpol.Subject)),
		objectExternalNetworks:               make([]gaia.ExternalNetworksList, len(netpol.Object)),
		objectProtocolPorts:                  make([][]string, len(netpol.Object)),
		subjectHasExternalNetworks:           false,
		allSubjectsReferenceExternalNetworks: true,
		objectHasExternalNetworks:            false,
		allObjectsReferenceExternalNetworks:  true,

		// Transformations
		transformations: make([]map[string]interface{}, 0),

		// Warnings
		subjectExternalNetworksNoNameRefButExtNetsFound: make([]bool, len(netpol.Subject)),
		subjectExternalNetworksPortMigrationNotPossible: make([]bool, len(netpol.Subject)),
		objectExternalNetworksNoNameRefButExtNetsFound:  make([]bool, len(netpol.Object)),
		objectExternalNetworksPortMigrationNotPossible:  make([]bool, len(netpol.Object)),
	}
}

func (n *netPolInfo) resolveExternalNetworks(extnetList gaia.ExternalNetworksList) {

	bidir := n.netpol.ApplyPolicyMode == gaia.NetworkAccessPolicyApplyPolicyModeBidirectional

	// Check for external network name references
	for i := range n.netpol.Subject {
		ref := refHasExtNextworks(n.netpol.Subject[i])
		n.subjectHasExternalNetworks = n.subjectHasExternalNetworks || ref
		n.allSubjectsReferenceExternalNetworks = n.allSubjectsReferenceExternalNetworks && ref
		n.subjectExternalNetworks[i] = extnetsFromTags(n.netpol.Namespace, n.netpol.Subject[i], extnetList)
	}
	for i := range n.netpol.Object {
		ref := refHasExtNextworks(n.netpol.Object[i])
		n.objectHasExternalNetworks = n.objectHasExternalNetworks || ref
		n.allObjectsReferenceExternalNetworks = n.allObjectsReferenceExternalNetworks && ref
		n.objectExternalNetworks[i] = extnetsFromTags(n.netpol.Namespace, n.netpol.Object[i], extnetList)
	}

	// Get actual matching external networks
	for i := range n.netpol.Subject {
		ref := refHasExtNextworks(n.netpol.Subject[i])
		if ref != (len(n.subjectExternalNetworks[i]) != 0) {
			n.subjectExternalNetworksNoNameRefButExtNetsFound[i] = true
			if !n.allObjectsReferenceExternalNetworks {
				n.exceptions = true
			}
		}
	}
	for i := range n.netpol.Object {
		ref := refHasExtNextworks(n.netpol.Object[i])
		if ref != (len(n.objectExternalNetworks[i]) != 0) {
			n.objectExternalNetworksNoNameRefButExtNetsFound[i] = true
			if !n.allSubjectsReferenceExternalNetworks {
				n.exceptions = true
				n.objectExternalNetworks[i] = extnetsFromTags(n.netpol.Namespace, n.netpol.Object[i], extnetList)
			}
		}
	}

	// Port/Protocol checks on external networks
	// All subject external networks should have the same service protos
	for i, eList := range n.subjectExternalNetworks {
		if ok := n.subjectExternalNetworksNoNameRefButExtNetsFound[i]; ok {
			continue
		}
		for j, e := range eList {
			portProtos := intersect(n.netpol.Ports, e.ServicePorts)
			if j == 0 {
				n.subjectProtocolPorts[i] = portProtos
			} else if !equalSlices(n.subjectProtocolPorts[i], portProtos) {
				n.exceptions = true
				n.subjectExternalNetworksPortMigrationNotPossible[i] = true
			}
		}
	}
	// All object external networks should have the same service protos
	for i, eList := range n.objectExternalNetworks {
		if ok := n.objectExternalNetworksNoNameRefButExtNetsFound[i]; ok {
			continue
		}
		for j, e := range eList {
			portProtos := intersect(n.netpol.Ports, e.ServicePorts)
			if j == 0 {
				n.objectProtocolPorts[i] = portProtos
			} else if !equalSlices(n.objectProtocolPorts[i], portProtos) {
				n.exceptions = true
				n.objectExternalNetworksPortMigrationNotPossible[i] = true
			}
		}
	}

	if n.netpol.NegateObject || n.netpol.NegateSubject {
		n.exceptions = true
		n.negationsNotSupported = true
	}

	n.candidateForUnidirectionalPolicy = bidir && (n.allObjectsReferenceExternalNetworks || n.allSubjectsReferenceExternalNetworks)
	n.ineffectivePolicy = (n.allObjectsReferenceExternalNetworks && n.allSubjectsReferenceExternalNetworks)
}

func (n *netPolInfo) checkAndPrintWarnings(verbose bool) bool {

	warning := ""

	if n.negationsNotSupported {
		warning += fmt.Sprintf("      - negationsNotSupported:            %v\n", n.negationsNotSupported)
	}
	if verbose || n.candidateForUnidirectionalPolicy {
		warning += fmt.Sprintf("      - candidateForUnidirectionalPolicy: %v\n", n.candidateForUnidirectionalPolicy)
	}
	if verbose || n.ineffectivePolicy {
		warning += fmt.Sprintf("      - ineffectivePolicy:                %v\n", n.ineffectivePolicy)
	}
	if len(n.subjectExternalNetworksNoNameRefButExtNetsFound) > 0 {
		buf := ""
		for i := 0; i < len(n.subjectExternalNetworksNoNameRefButExtNetsFound); i++ {
			if verbose || n.subjectExternalNetworksPortMigrationNotPossible[i] {
				buf += fmt.Sprintf("          - [%d] subjectExternalNetworksPortMigrationNotPossible: %v\n", i, n.subjectExternalNetworksPortMigrationNotPossible[i])
			}
			if verbose || n.subjectExternalNetworksNoNameRefButExtNetsFound[i] {
				buf += fmt.Sprintf("          - [%d] subjectExternalNetworksNoNameRefButExtNetsFound: %v\n", i, n.subjectExternalNetworksNoNameRefButExtNetsFound[i])
			}
		}
		if buf != "" {
			warning += "      - subjects:\n"
			warning += buf
		}
	}
	if len(n.objectExternalNetworksNoNameRefButExtNetsFound) > 0 {
		buf := ""
		for i := 0; i < len(n.objectExternalNetworksNoNameRefButExtNetsFound); i++ {
			if verbose || n.objectExternalNetworksPortMigrationNotPossible[i] {
				buf += fmt.Sprintf("          - [%d] objectExternalNetworksPortMigrationNotPossible: %v\n", i, n.objectExternalNetworksPortMigrationNotPossible[i])
			}
			if verbose || n.objectExternalNetworksNoNameRefButExtNetsFound[i] {
				buf += fmt.Sprintf("          - [%d] objectExternalNetworksNoNameRefButExtNetsFound: %v\n", i, n.objectExternalNetworksNoNameRefButExtNetsFound[i])
			}
		}
		if buf != "" {
			warning += "      - objects:\n"
			warning += buf
		}
	}

	if verbose || warning != "" {

		if n.exceptions {
			warning = fmt.Sprintf("\n    Policy=%s (Exception)\n", n.netpol.Name) + warning
		} else {
			warning = fmt.Sprintf("\n    Policy=%s\n", n.netpol.Name) + warning
		}

		fmt.Println(warning)
	}

	return warning == ""
}

func (n *netPolInfo) xfrm() {

	n.outgoing = gaia.NewNetworkRuleSetPolicy()
	n.outgoing.Annotations = n.netpol.Annotations
	n.outgoing.AssociatedTags = n.netpol.AssociatedTags
	n.outgoing.CreateIdempotencyKey = n.netpol.CreateIdempotencyKey
	n.outgoing.Description = n.netpol.Description
	n.outgoing.Disabled = n.netpol.Disabled
	n.outgoing.Fallback = n.netpol.Fallback
	n.outgoing.Metadata = n.netpol.Metadata
	n.outgoing.Name = n.netpol.Name + migrationSuffix
	n.outgoing.Namespace = n.netpol.Namespace
	n.outgoing.NormalizedTags = n.netpol.NormalizedTags
	n.outgoing.Propagate = n.netpol.Propagate
	n.outgoing.Protected = n.netpol.Protected
	n.outgoing.UpdateIdempotencyKey = n.netpol.UpdateIdempotencyKey

	n.outgoingRule = gaia.NewNetworkRule()
	if n.netpol.Action == gaia.NetworkAccessPolicyActionAllow {
		n.outgoingRule.Action = gaia.NetworkRuleActionAllow
	} else if n.netpol.Action == gaia.NetworkAccessPolicyActionReject {
		n.outgoingRule.Action = gaia.NetworkRuleActionReject
	}
	n.outgoingRule.LogsDisabled = !n.netpol.LogsEnabled
	n.outgoingRule.ObservationEnabled = n.netpol.ObservationEnabled
	n.outgoingRule.ProtocolPorts = n.netpol.Ports // TODO

	// Make copies for incoming ruleset and rule
	n.incoming = n.outgoing.DeepCopy()
	n.incomingRule = n.outgoingRule.DeepCopy()

	// Setup subjects/objects
	subjects := refConvertTags(n.netpol.Subject)
	objects := refConvertTags(n.netpol.Object)
	n.incoming.Subject = objects
	n.outgoing.Subject = subjects

	// Setup incoming rules
	n.incoming.IncomingRules = []*gaia.NetworkRule{}
	for index, s := range subjects {
		i := n.incomingRule.DeepCopy()
		i.Object = [][]string{s}
		i.ProtocolPorts = n.subjectProtocolPorts[index]
		n.incoming.IncomingRules = append(n.incoming.IncomingRules, i)
	}

	// Setup outgoing rules
	n.outgoing.OutgoingRules = []*gaia.NetworkRule{}
	for index, s := range objects {
		o := n.outgoingRule.DeepCopy()
		o.Object = [][]string{s}
		o.ProtocolPorts = n.objectProtocolPorts[index]
		n.outgoing.OutgoingRules = append(n.outgoing.OutgoingRules, o)
	}

	// Add reflexive rules for bidirectional policies
	if n.netpol.ApplyPolicyMode == gaia.NetworkAccessPolicyApplyPolicyModeBidirectional {
		n.outgoing.IncomingRules = n.incoming.IncomingRules
		n.incoming.OutgoingRules = n.outgoing.OutgoingRules
	}

	// Create transformations
	if n.netpol.ApplyPolicyMode == gaia.NetworkAccessPolicyApplyPolicyModeBidirectional ||
		n.netpol.ApplyPolicyMode == gaia.NetworkAccessPolicyApplyPolicyModeIncomingTraffic {
		xn := map[string]interface{}{}
		if err := mapstructure.Decode(n.incoming, &xn); err != nil {
			panic(err)
		}

		for k, v := range xn {
			keySpec := n.incoming.SpecificationForAttribute(strings.ToLower(k))
			if !keySpec.Exposed || keySpec.ReadOnly || keySpec.Autogenerated || v == keySpec.DefaultValue {
				delete(xn, k)
			}
		}

		n.transformations = append(n.transformations, xn)
	}

	if n.netpol.ApplyPolicyMode == gaia.NetworkAccessPolicyApplyPolicyModeBidirectional ||
		n.netpol.ApplyPolicyMode == gaia.NetworkAccessPolicyApplyPolicyModeOutgoingTraffic {
		xn := map[string]interface{}{}
		if err := mapstructure.Decode(n.outgoing, &xn); err != nil {
			panic(err)
		}

		for k, v := range xn {
			keySpec := n.outgoing.SpecificationForAttribute(strings.ToLower(k))
			if !keySpec.Exposed || keySpec.ReadOnly || keySpec.Autogenerated || v == keySpec.DefaultValue {
				delete(xn, k)
			}
		}

		n.transformations = append(n.transformations, xn)
	}
}

func getNetPolInfo(netpol *gaia.NetworkAccessPolicy, extnetList gaia.ExternalNetworksList) (*netPolInfo, error) {

	var err error
	n := newNetPolInfo(netpol)
	n.resolveExternalNetworks(extnetList)
	if !n.checkAndPrintWarnings(false) {
		err = fmt.Errorf("policy: %s warnings/errors found", netpol.Name)
	}
	n.xfrm()
	return n, err
}
