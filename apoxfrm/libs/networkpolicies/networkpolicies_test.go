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

	// Network Policies
	netpol := gaia.NewNetworkAccessPolicy()
	netpol.Name = "ports"
	netpol.Namespace = "/customer/root/zone/tenant"
	netpol.ApplyPolicyMode = gaia.NetworkAccessPolicyApplyPolicyModeIncomingTraffic
	netpol.Object = [][]string{{"$identity=processingunit"}}
	netpol.Subject = [][]string{
		{"$name=ssh", "customer:ext:net=ssh"},
		{"customer:ext:net=tenant"},
	}
	netpol.Ports = []string{"tcp/22", "udp/52"}
	netpol.Action = gaia.NetworkAccessPolicyActionAllow

	netpolNoPorts := gaia.NewNetworkAccessPolicy()
	netpolNoPorts.Name = "no-ports"
	netpolNoPorts.Namespace = "/customer/root/zone/tenant"
	netpolNoPorts.ApplyPolicyMode = gaia.NetworkAccessPolicyApplyPolicyModeIncomingTraffic
	netpolNoPorts.Object = [][]string{{"$identity=processingunit"}}
	netpolNoPorts.Subject = [][]string{
		{"$name=ssh", "customer:ext:net=ssh"},
		{"customer:ext:net=tenant"},
	}
	netpolNoPorts.Action = gaia.NetworkAccessPolicyActionAllow

	// External Networks
	extnetList := gaia.ExternalNetworksList{
		&gaia.ExternalNetwork{
			Name:           "ssh",
			Namespace:      "/customer/root",
			AssociatedTags: []string{"customer:namespace=/customer/root", "customer:ext:net=ssh"},
			NormalizedTags: []string{"$name=ssh", "$namespace=/customer/root", "customer:namespace=/customer/root", "customer:ext:net=ssh"},
			ServicePorts:   []string{"tcp/22"},
			Propagate:      true,
		},
		&gaia.ExternalNetwork{
			Name:           "traceroute",
			Namespace:      "/root",
			AssociatedTags: []string{"customer:namespace=/customer/root", "customer:ext:net=traceroute"},
			NormalizedTags: []string{"$name=traceroute", "$namespace=/customer/root", "customer:namespace=/customer/root", "customer:ext:net=traceroute"},
			ServicePorts:   []string{"icmp"},
			Propagate:      true,
		},
		&gaia.ExternalNetwork{
			Name:           "dhcp",
			Namespace:      "/root",
			AssociatedTags: []string{"customer:namespace=/customer/root", "customer:ext:net=dhcp"},
			NormalizedTags: []string{"$name=dhcp", "$namespace=/customer/root", "customer:namespace=/customer/root", "customer:ext:net=dhcp"},
			ServicePorts:   []string{"udp"},
			Propagate:      true,
		},

		// Tenant level
		&gaia.ExternalNetwork{
			Name:           "tenant",
			Namespace:      "/customer/root/zone/tenant",
			AssociatedTags: []string{"customer:namespace=/customer/root", "customer:ext:net=tenant"},
			NormalizedTags: []string{"$name=tenant", "$namespace=/customer/root/zone/tenant", "customer:namespace=/customer/root/zone/tenant", "customer:ext:net=tenant"},
			ServicePorts:   []string{"tcp/443"},
			Propagate:      true,
		},
	}
	// Network Rules
	netrule := gaia.NetworkRule{
		Action:       gaia.NetworkRuleActionAllow,
		LogsDisabled: true,
		Object: [][]string{
			{"$name=ssh" + utils.MigrationSuffix, "customer:ext:net=ssh" + utils.MigrationSuffix},
		},
		ProtocolPorts: []string{"TCP/22", "UDP/52"},
		ModelVersion:  1,
	}
	tenantnetrule := gaia.NetworkRule{
		Action:       gaia.NetworkRuleActionAllow,
		LogsDisabled: true,
		Object: [][]string{
			{"$name=tenant" + utils.MigrationSuffix, "customer:ext:net=tenant" + utils.MigrationSuffix},
		},
		ProtocolPorts: []string{"UDP/52"},
		ModelVersion:  1,
	}
	netruleNoPorts := gaia.NetworkRule{
		Action:       gaia.NetworkRuleActionAllow,
		LogsDisabled: true,
		Object: [][]string{
			{"$name=ssh" + utils.MigrationSuffix, "customer:ext:net=ssh" + utils.MigrationSuffix},
		},
		ProtocolPorts: []string{"TCP/22"},
		ModelVersion:  1,
	}
	tenantnetruleNoPorts := gaia.NetworkRule{
		Action:       gaia.NetworkRuleActionAllow,
		LogsDisabled: true,
		Object: [][]string{
			{"$name=tenant" + utils.MigrationSuffix, "customer:ext:net=tenant" + utils.MigrationSuffix},
		},
		ProtocolPorts: []string{"TCP/443"},
		ModelVersion:  1,
	}
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
			name: "ports intersection",
			args: args{
				netpol:     netpol,
				extnetList: extnetList,
			},
			prefix: "customer:ext:net=",
			want: []map[string]interface{}{
				{
					"incomingRules": []*gaia.NetworkRule{&netrule, &tenantnetrule},
					"name":          "ports-v2",
					"subject":       [][]string{{"$identity=processingunit"}},
				},
			},
			wantErr: false,
		},
		{
			name: "no ports",
			args: args{
				netpol:     netpolNoPorts,
				extnetList: extnetList,
			},
			prefix: "customer:ext:net=",
			want: []map[string]interface{}{
				{
					"incomingRules": []*gaia.NetworkRule{&netruleNoPorts, &tenantnetruleNoPorts},
					"name":          "no-ports-v2",
					"subject":       [][]string{{"$identity=processingunit"}},
				},
			},
			wantErr: false,
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
					if mk == "incomingRules" {
						if !compareNetworkRules(t, mv.([]*gaia.NetworkRule), wv.([]*gaia.NetworkRule)) {
							t.Errorf("Get() mv=%v, want wv=%v", mv, wv)
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