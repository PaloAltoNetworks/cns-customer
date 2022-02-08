package externalnetwork

import (
	"testing"

	"github.com/PaloAltoNetworks/cns-customer/apoxfrm/libs/utils"
	"go.aporeto.io/gaia"
)

func TestTransform(t *testing.T) {
	type args struct {
		extnet *gaia.ExternalNetwork
	}
	tests := []struct {
		name string
		args args
		want *gaia.ExternalNetwork
	}{
		{
			name: "name",
			args: args{
				extnet: &gaia.ExternalNetwork{
					Name: "hello",
				},
			},
			want: &gaia.ExternalNetwork{
				Name: "hello" + utils.MigrationSuffix,
			},
		},
		{
			name: "service ports",
			args: args{
				extnet: &gaia.ExternalNetwork{
					Name: "service-port",
					ServicePorts: []string{
						"tcp/500",
					},
				},
			},
			want: &gaia.ExternalNetwork{
				Name: "service-port" + utils.MigrationSuffix,
			},
		},
		{
			name: "associated tags",
			args: args{
				extnet: &gaia.ExternalNetwork{
					Name:           "hello",
					AssociatedTags: []string{"$name=hello", "customer:ext:net=dhcp", "apple=pie"},
				},
			},
			want: &gaia.ExternalNetwork{
				Name:           "hello" + utils.MigrationSuffix,
				AssociatedTags: []string{"$name=hello", "customer:ext:net=dhcp" + utils.MigrationSuffix, "apple=pie"},
			},
		},
	}
	utils.ExtnetPrefix = "customer:ext:net="
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Transform(tt.args.extnet)
			if got.Name != tt.want.Name {
				t.Errorf("Transform() Name = %v, want %v", got.Name, tt.want.Name)
			}
			if got.Namespace != tt.want.Namespace {
				t.Errorf("Transform() Namespace = %v, want %v", got.Namespace, tt.want.Namespace)
			}
			if len(got.AssociatedTags) != len(tt.want.AssociatedTags) {
				t.Errorf("Transform() AssociatedTags = %v, want %v", got.AssociatedTags, tt.want.AssociatedTags)
			}
			for i := range got.AssociatedTags {
				if got.AssociatedTags[i] != tt.want.AssociatedTags[i] {
					t.Errorf("Transform() AssociatedTags = %v, want %v", got.AssociatedTags[i], tt.want.AssociatedTags[i])
				}
			}
			if len(got.ServicePorts) != len(tt.want.ServicePorts) {
				t.Errorf("Transform() ServicePorts = %v, want %v", got.ServicePorts, tt.want.ServicePorts)
			}
			for i := range got.ServicePorts {
				if got.ServicePorts[i] != tt.want.ServicePorts[i] {
					t.Errorf("Transform() ServicePorts = %v, want %v", got.ServicePorts[i], tt.want.ServicePorts[i])
				}
			}
		})
	}
}
