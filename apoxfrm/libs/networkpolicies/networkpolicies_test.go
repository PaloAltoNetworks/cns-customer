package networkpolicies

import (
	"reflect"
	"testing"

	"github.com/PaloAltoNetworks/cns-customer/apoxfrm/libs/utils"
	"go.aporeto.io/gaia"
)

func compareNetworks(a, b []*gaia.NetworkRuleNet) bool {

	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if v.Namespace != b[i].Namespace {
			return false
		}
		if v.ModelVersion != b[i].ModelVersion {
			return false
		}
		if !equalSlices(v.Entries, b[i].Entries) {
			return false
		}
	}

	return true
}

func compareNetworkRules(t *testing.T, a, b []*gaia.NetworkRule) bool {

	if len(a) != len(b) {
		t.Errorf("Get() len %v, want %v", len(a), len(b))
		return false
	}

	for i := range a {
		if a[i].Action != b[i].Action {
			t.Errorf("Get() Action %v, want %v", a[i].Action, b[i].Action)
			return false
		}
		if a[i].LogsDisabled != b[i].LogsDisabled {
			t.Errorf("Get() LogsDisabled %v, want %v", a[i].LogsDisabled, b[i].LogsDisabled)
			return false
		}
		if a[i].Name != b[i].Name {
			t.Errorf("Get() Name %v, want %v", a[i].Name, b[i].Name)
			return false
		}
		if !compareNetworks(a[i].Networks, b[i].Networks) {
			t.Errorf("Get() Networks %v, want %v", a[i].Networks, b[i].Networks)
			return false
		}
		if a[i].ObservationEnabled != b[i].ObservationEnabled {
			t.Errorf("Get() ObservationEnabled %v, want %v", a[i].ObservationEnabled, b[i].ObservationEnabled)
			return false
		}
		if a[i].ModelVersion != b[i].ModelVersion {
			t.Errorf("Get() ModelVersion %v, want %v", a[i].ModelVersion, b[i].ModelVersion)
			return false
		}
		if !equalSlices(a[i].ProtocolPorts, b[i].ProtocolPorts) {
			t.Errorf("Get() ProtocolPorts %v, want %v", a[i].ProtocolPorts, b[i].ProtocolPorts)
			return false
		}
	}
	return true
}

func TestGet(t *testing.T) {

	// External Networks
	extnetList := gaia.ExternalNetworksList{
		&gaia.ExternalNetwork{
			Name:           "ssh",
			Namespace:      "/customer/root",
			AssociatedTags: []string{"customer:namespace=/customer/root", "customer:ext:net=ssh"},
			NormalizedTags: []string{},
			ServicePorts:   []string{"tcp/22"},
			Propagate:      true,
		},
		&gaia.ExternalNetwork{
			Name:           "traceroute",
			Namespace:      "/root",
			AssociatedTags: []string{"customer:namespace=/customer/root", "customer:ext:net=traceroute"},
			ServicePorts:   []string{"icmp"},
			Propagate:      true,
		},
		&gaia.ExternalNetwork{
			Name:           "dhcp",
			Namespace:      "/root",
			AssociatedTags: []string{"customer:namespace=/customer/root", "customer:ext:net=dhcp"},
			ServicePorts:   []string{"udp"},
			Propagate:      true,
		},

		// Tenant level
		&gaia.ExternalNetwork{
			Name:           "tenant",
			Namespace:      "/customer/root/zone/tenant",
			AssociatedTags: []string{"customer:namespace=/customer/root", "customer:ext:net=tenant"},
			ServicePorts:   []string{"tcp/443"},
			Propagate:      true,
		},
	}

	// Incoming Network Policy with no ports
	incomingNetpolNoPorts := gaia.NewNetworkAccessPolicy()
	incomingNetpolNoPorts.Name = "no-ports-incoming"
	incomingNetpolNoPorts.Namespace = "/customer/root/zone/tenant"
	incomingNetpolNoPorts.Disabled = true
	incomingNetpolNoPorts.ApplyPolicyMode = gaia.NetworkAccessPolicyApplyPolicyModeIncomingTraffic
	incomingNetpolNoPorts.Object = [][]string{{"$identity=processingunit"}}
	incomingNetpolNoPorts.Subject = [][]string{
		{"$name=ssh", "$identity=externalnetwork", "customer:ext:net=ssh"},
		{"customer:ext:net=tenant"},
	}
	// Network rules to support no ports policy
	incomingNetpolNoPorts.Action = gaia.NetworkAccessPolicyActionAllow
	incomingNetruleNoPorts := gaia.NetworkRule{
		Action:       gaia.NetworkRuleActionAllow,
		LogsDisabled: true,
		Object: [][]string{
			{"$name=ssh" + utils.MigrationSuffix, "$identity=externalnetwork", "customer:ext:net=ssh" + utils.MigrationSuffix},
		},
		ProtocolPorts: []string{"TCP/22"},
		ModelVersion:  1,
	}
	incomingTenantNetruleNoPorts := gaia.NetworkRule{
		Action:       gaia.NetworkRuleActionAllow,
		LogsDisabled: true,
		Object: [][]string{
			{"$name=tenant" + utils.MigrationSuffix, "$identity=externalnetwork", "customer:ext:net=tenant" + utils.MigrationSuffix},
		},
		ProtocolPorts: []string{"TCP/443"},
		ModelVersion:  1,
	}
	incomingNoPortsWant := []map[string]interface{}{
		{
			"disabled":      true,
			"incomingRules": []*gaia.NetworkRule{&incomingNetruleNoPorts, &incomingTenantNetruleNoPorts},
			"name":          "no-ports-incoming-v2",
			"subject":       [][]string{{"$identity=processingunit"}},
		},
	}

	// Incoming Network Policy with ports
	incomingNetpol := gaia.NewNetworkAccessPolicy()
	incomingNetpol.Name = "ports-incoming"
	incomingNetpol.Namespace = "/customer/root/zone/tenant"
	incomingNetpol.ApplyPolicyMode = gaia.NetworkAccessPolicyApplyPolicyModeIncomingTraffic
	incomingNetpol.Object = [][]string{{"$identity=processingunit"}}
	incomingNetpol.Subject = [][]string{
		{"$name=ssh", "$identity=externalnetwork", "customer:ext:net=ssh"},
		{"customer:ext:net=tenant"},
	}
	incomingNetpol.Ports = []string{"tcp/22", "udp/52"}
	incomingNetpol.Action = gaia.NetworkAccessPolicyActionAllow
	// Network Rules with port intersection with policies
	incomingNetrule := gaia.NetworkRule{
		Action:       gaia.NetworkRuleActionAllow,
		LogsDisabled: true,
		Object: [][]string{
			{"$name=ssh" + utils.MigrationSuffix, "$identity=externalnetwork", "customer:ext:net=ssh" + utils.MigrationSuffix},
		},
		ProtocolPorts: []string{"TCP/22", "UDP/52"},
		ModelVersion:  1,
	}
	incomingTenantNetrule := gaia.NetworkRule{
		Action:       gaia.NetworkRuleActionAllow,
		LogsDisabled: true,
		Object: [][]string{
			{"$name=tenant" + utils.MigrationSuffix, "$identity=externalnetwork", "customer:ext:net=tenant" + utils.MigrationSuffix},
		},
		ProtocolPorts: []string{"UDP/52"},
		ModelVersion:  1,
	}
	incomingWant := []map[string]interface{}{
		{
			"incomingRules": []*gaia.NetworkRule{&incomingNetrule, &incomingTenantNetrule},
			"name":          "ports-incoming-v2",
			"subject":       [][]string{{"$identity=processingunit"}},
		},
	}

	// Outgoing Network Policy with no ports
	outgoingNetpolNoPorts := gaia.NewNetworkAccessPolicy()
	outgoingNetpolNoPorts.Name = "no-ports-outgoing"
	outgoingNetpolNoPorts.Namespace = "/customer/root/zone/tenant"
	outgoingNetpolNoPorts.ApplyPolicyMode = gaia.NetworkAccessPolicyApplyPolicyModeOutgoingTraffic
	outgoingNetpolNoPorts.Object = [][]string{
		{"$name=ssh", "$identity=externalnetwork", "customer:ext:net=ssh"},
		{"customer:ext:net=tenant"},
	}
	outgoingNetpolNoPorts.Subject = [][]string{{"$identity=processingunit"}}
	outgoingNetpolNoPorts.Action = gaia.NetworkAccessPolicyActionAllow
	// Network rules to support no ports policy
	outgoingNetruleNoPorts := gaia.NetworkRule{
		Action:       gaia.NetworkRuleActionAllow,
		LogsDisabled: true,
		Object: [][]string{
			{"$name=ssh" + utils.MigrationSuffix, "$identity=externalnetwork", "customer:ext:net=ssh" + utils.MigrationSuffix},
		},
		ProtocolPorts: []string{"TCP/22"},
		ModelVersion:  1,
	}
	outgoingTenantNetruleNoPorts := gaia.NetworkRule{
		Action:       gaia.NetworkRuleActionAllow,
		LogsDisabled: true,
		Object: [][]string{
			{"$name=tenant" + utils.MigrationSuffix, "$identity=externalnetwork", "customer:ext:net=tenant" + utils.MigrationSuffix},
		},
		ProtocolPorts: []string{"TCP/443"},
		ModelVersion:  1,
	}
	outgoingNoPortsWant := []map[string]interface{}{
		{
			"outgoingRules": []*gaia.NetworkRule{&outgoingNetruleNoPorts, &outgoingTenantNetruleNoPorts},
			"name":          "no-ports-outgoing-v2",
			"subject":       [][]string{{"$identity=processingunit"}},
		},
	}

	// Outgoing Network Policy with ports
	outgoingNetpol := gaia.NewNetworkAccessPolicy()
	outgoingNetpol.Name = "ports-outgoing"
	outgoingNetpol.Namespace = "/customer/root/zone/tenant"
	outgoingNetpol.ApplyPolicyMode = gaia.NetworkAccessPolicyApplyPolicyModeOutgoingTraffic
	outgoingNetpol.Object = [][]string{
		{"$name=ssh", "$identity=externalnetwork", "customer:ext:net=ssh"},
		{"customer:ext:net=tenant"},
	}
	outgoingNetpol.Subject = [][]string{{"$identity=processingunit"}}
	outgoingNetpol.Ports = []string{"tcp/22", "udp/52"}
	outgoingNetpol.Action = gaia.NetworkAccessPolicyActionAllow
	// Network Rules with port intersection with policies
	outgoingNetrule := gaia.NetworkRule{
		Action:       gaia.NetworkRuleActionAllow,
		LogsDisabled: true,
		Object: [][]string{
			{"$name=ssh" + utils.MigrationSuffix, "$identity=externalnetwork", "customer:ext:net=ssh" + utils.MigrationSuffix},
		},
		ProtocolPorts: []string{"TCP/22", "UDP/52"},
		ModelVersion:  1,
	}
	outgoingTenantNetrule := gaia.NetworkRule{
		Action:       gaia.NetworkRuleActionAllow,
		LogsDisabled: true,
		Object: [][]string{
			{"$name=tenant" + utils.MigrationSuffix, "$identity=externalnetwork", "customer:ext:net=tenant" + utils.MigrationSuffix},
		},
		ProtocolPorts: []string{"UDP/52"},
		ModelVersion:  1,
	}
	outgoingWant := []map[string]interface{}{
		{
			"outgoingRules": []*gaia.NetworkRule{&outgoingNetrule, &outgoingTenantNetrule},
			"name":          "ports-outgoing-v2",
			"subject":       [][]string{{"$identity=processingunit"}},
		},
	}

	// Bidirectional Network Policy with ports
	bidirectionalNetpol := gaia.NewNetworkAccessPolicy()
	bidirectionalNetpol.Name = "ports-bidirectional"
	bidirectionalNetpol.Namespace = "/customer/root/zone/tenant"
	bidirectionalNetpol.ApplyPolicyMode = gaia.NetworkAccessPolicyApplyPolicyModeBidirectional
	bidirectionalNetpol.Object = [][]string{
		{"$name=ssh", "$identity=externalnetwork", "customer:ext:net=ssh"},
		{"customer:ext:net=tenant"},
	}
	bidirectionalNetpol.Subject = [][]string{{"$identity=processingunit"}}
	bidirectionalNetpol.Ports = []string{"tcp/22", "udp/52"}
	bidirectionalNetpol.Action = gaia.NetworkAccessPolicyActionAllow
	// Network Rules with port intersection with policies
	bidirectionalNetrule := gaia.NetworkRule{
		Action:       gaia.NetworkRuleActionAllow,
		LogsDisabled: true,
		Object: [][]string{
			{"$name=ssh" + utils.MigrationSuffix, "$identity=externalnetwork", "customer:ext:net=ssh" + utils.MigrationSuffix},
		},
		ProtocolPorts: []string{"TCP/22", "UDP/52"},
		ModelVersion:  1,
	}
	bidirectionalTenantNetrule := gaia.NetworkRule{
		Action:       gaia.NetworkRuleActionAllow,
		LogsDisabled: true,
		Object: [][]string{
			{"$name=tenant" + utils.MigrationSuffix, "$identity=externalnetwork", "customer:ext:net=tenant" + utils.MigrationSuffix},
		},
		ProtocolPorts: []string{"UDP/52"},
		ModelVersion:  1,
	}
	bidirectionalReflexiveNetrule := gaia.NetworkRule{
		Action:       gaia.NetworkRuleActionAllow,
		LogsDisabled: true,
		Object: [][]string{
			{"$identity=processingunit"},
		},
		ProtocolPorts: []string{"UDP/52"},
		ModelVersion:  1,
	}
	bidirectionalWant := []map[string]interface{}{
		{
			"outgoingRules": []*gaia.NetworkRule{&bidirectionalReflexiveNetrule},
			"incomingRules": []*gaia.NetworkRule{&bidirectionalReflexiveNetrule},
			"name":          "ports-bidirectional-v2",
			"subject": [][]string{
				{"$name=ssh-v2", "$identity=externalnetwork", "customer:ext:net=ssh-v2"},
				{"customer:ext:net=tenant-v2"},
			},
		},
		{
			"outgoingRules": []*gaia.NetworkRule{&bidirectionalNetrule, &bidirectionalTenantNetrule},
			"incomingRules": []*gaia.NetworkRule{&bidirectionalNetrule, &bidirectionalTenantNetrule},
			"name":          "ports-bidirectional-v2",
			"subject":       [][]string{{"$identity=processingunit"}},
		},
	}

	// Tests
	type args struct {
		netpol     *gaia.NetworkAccessPolicy
		extnetList gaia.ExternalNetworksList
	}
	tests := []struct {
		name    string
		args    args
		prefix  string
		want    []map[string]interface{}
		wantErr bool
	}{
		{
			name: "incoming no ports",
			args: args{
				netpol:     incomingNetpolNoPorts,
				extnetList: extnetList,
			},
			prefix:  "customer:ext:net=",
			want:    incomingNoPortsWant,
			wantErr: false,
		},
		{
			name: "incoming ports intersection",
			args: args{
				netpol:     incomingNetpol,
				extnetList: extnetList,
			},
			prefix:  "customer:ext:net=",
			want:    incomingWant,
			wantErr: false,
		},
		{
			name: "outgoing no ports",
			args: args{
				netpol:     outgoingNetpolNoPorts,
				extnetList: extnetList,
			},
			prefix:  "customer:ext:net=",
			want:    outgoingNoPortsWant,
			wantErr: false,
		},
		{
			name: "outgoing ports intersection",
			args: args{
				netpol:     outgoingNetpol,
				extnetList: extnetList,
			},
			prefix:  "customer:ext:net=",
			want:    outgoingWant,
			wantErr: false,
		},
		{
			name: "bidirectional no ports error (could be unidirectional)",
			args: args{
				netpol:     bidirectionalNetpol,
				extnetList: extnetList,
			},
			prefix:  "customer:ext:net=",
			want:    bidirectionalWant,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utils.ExtnetPrefix = tt.prefix
			got, err := Get(tt.args.netpol, tt.args.extnetList)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("Get() len(got)=%v, want len(tt.want)=%v", len(got), len(tt.want))
			}
			for i, v := range got {
				if len(v) != len(tt.want[i]) {
					t.Errorf("Get() len(v)=%v, want len(tt.want[i])=%v", len(v), len(tt.want[i]))
				}
				for mk, mv := range v {
					wv, ok := tt.want[i][mk]
					if !ok {
						t.Errorf("Get() missing key %v in tt.want[i][mk]", mk)
					}
					if mk == "incomingRules" || mk == "outgoingRules" {
						mvx, mok := mv.([]*gaia.NetworkRule)
						wvx, wok := mv.([]*gaia.NetworkRule)
						if !(mok && wok) {
							t.Errorf("Get() mvx=%v, want wvx=%v", mvx, wvx)
						}
						if !compareNetworkRules(t, mvx, wvx) {
							t.Errorf("Get() mv=%v, want wv=%v", mvx, wvx)
						}
					} else {
						if !reflect.DeepEqual(mv, wv) {
							t.Errorf("Get() mv=%v, want wv=%v", mv, wv)
						}
					}
				}
			}
		})
	}
}
