package constants

import "time"

const (
	// APIDefaultContextTimeout is the default time allowed for a context.
	APIDefaultContextTimeout = 10 * time.Second
)

// NamespaceKeys
const (
	NamespaceKey = "$namespace="
)

// Child Namespace in a Tenant
const (
	NamespacePublic    = "public"
	NamespaceProtected = "protected"
	NamespacePrivate   = "private"
)

// Authorization Roles
const (
	AuthNamespaceViewer  = "@auth:role=namespace.viewer"
	AuthEnforcerd        = "@auth:role=enforcer"
	AuthEnforcerdRuntime = "@auth:role=enforcer.runtime"
)

// DefaultTenantROAuthPolicy is the name of default policy
const DefaultTenantROAuthPolicy = "default"

// MetadataKeys
const (
	MetadataOwnerKeyVal  = "@cns-customer:owner=soc"
	MetadataTenantKey    = "@cns-customer:tenant="
	MetadataNamespaceKey = "@cns-customer:namespace="
)

// AssociatedTagKeys
const (
	AssociatedTagNamespaceKey       = "cns-customer:namespace="
	AssociatedTagExternalNetworkKey = "cns-customer:ext:network="
	AssociatedTagHostServiceKey     = "cns-customer:ext:hostservice="
)

// ExcludedManagementServices
const (
	ManagementServiceName = "ssh"
	ManagementServices    = "tcp/22"
)

// ExternalNetworkConstants
const (
	ExternalNetworkAllTCP     = "all-tcp"
	ExternalNetworkAllUDP     = "all-udp"
	ExternalNetworkAnyCIDR    = "0.0.0.0/0"
	ExternalNetworkAllPorts   = "1:65535"
	ExternalNetworkProtcolTCP = "tcp"
	ExternalNetworkProtcolUDP = "udp"
)
