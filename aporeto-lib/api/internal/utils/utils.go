package utils

import (
	"strings"

	"github.com/PaloAltoNetworks/cns-customer/aporeto-lib/api/constants"
)

// SetupNamespaceString returns a namespace string. It
// accounts for any leading/trailing '/'
func SetupNamespaceString(namespaces ...string) string {

	absoluteNamespace := ""
	for _, namespace := range namespaces {
		ns := strings.TrimLeft(namespace, "/")
		ns = strings.TrimRight(ns, "/")
		absoluteNamespace = absoluteNamespace + "/" + ns
	}
	return absoluteNamespace
}

// MetadataNamespaceKeyVal provides the key value pair for a Namespace
func MetadataNamespaceKeyVal(namespace string) string {
	return constants.MetadataNamespaceKey + SetupNamespaceString(namespace)
}

// MakeNamespaceMetadata provides Namespace metadata thats populated on all objects for a given Namespace.
func MakeNamespaceMetadata(namespace string) []string {
	return []string{constants.MetadataOwnerKeyVal, MetadataNamespaceKeyVal(namespace)}
}

// AssociatedTagExternalNetworkKeyVal provides the key value pair for an External Network
func AssociatedTagExternalNetworkKeyVal(name string) string {
	return constants.AssociatedTagExternalNetworkKey + name
}

// MakeExternalNetworkAssociatedTags provides ExternalNetwork associated tags for a given External Network.
func MakeExternalNetworkAssociatedTags(name string) []string {
	return []string{AssociatedTagExternalNetworkKeyVal(name)}
}

// AssociatedTagHostServiceKeyVal provides the key value pair for a Host Service
func AssociatedTagHostServiceKeyVal(name string) string {
	return constants.AssociatedTagHostServiceKey + name
}

// MakeHostServiceAssociatedTags provides HostService associated tags for a given Host Service.
func MakeHostServiceAssociatedTags(name string) []string {
	return []string{AssociatedTagHostServiceKeyVal(name)}
}

// AssociatedTagNamespaceKeyVal provides the key value pair for a Namespace
func AssociatedTagNamespaceKeyVal(namespace string) string {
	return constants.AssociatedTagNamespaceKey + SetupNamespaceString(namespace)
}

// MakeNamespaceAssociatedTags provides Namespace associated tags for a given Namespace.
func MakeNamespaceAssociatedTags(namespace string) []string {
	return []string{AssociatedTagNamespaceKeyVal(namespace)}
}

// MaketenantNamespaceMetadata provides Namespace metadata thats populated on all objects for a given Namespace and Tenant.
func MaketenantNamespaceMetadata(tenant, namespace string) []string {
	return []string{constants.MetadataOwnerKeyVal, MetadataTenantKeyVal(tenant), MetadataNamespaceKeyVal(namespace)}
}

// MetadataTenantKeyVal provides the key value pair for a tenant
func MetadataTenantKeyVal(namespace string) string {
	return constants.MetadataTenantKey + SetupNamespaceString(namespace)
}

// MakeTenantMetadata provides tenant metadata thats populated on all objects for a given tenant.
func MakeTenantMetadata(namespace string) []string {
	return []string{constants.MetadataOwnerKeyVal, MetadataTenantKeyVal(namespace)}
}

// MakeTenantPairMetadata provides tenant metadata thats populated on all objects for a given src/dst tenant.
func MakeTenantPairMetadata(srcNamespace, dstNamespace string) []string {
	if srcNamespace == dstNamespace {
		return []string{constants.MetadataOwnerKeyVal, MetadataTenantKeyVal(srcNamespace)}
	}
	return []string{constants.MetadataOwnerKeyVal, MetadataTenantKeyVal(srcNamespace), MetadataTenantKeyVal(dstNamespace)}
}

// MakeOwnerMetadata provides owner metadata populated on all objects created by soc
func MakeOwnerMetadata() []string {
	return []string{constants.MetadataOwnerKeyVal}
}

// MakeNamespaceKeyVal returns a namespace tag
func MakeNamespaceKeyVal(namespace string) []string {
	return []string{constants.NamespaceKey + namespace}
}
