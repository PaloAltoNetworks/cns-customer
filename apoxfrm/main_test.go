package main

import (
	"reflect"
	"testing"

	"go.aporeto.io/gaia"
)

func Test_extnetsFromTags(t *testing.T) {
	type args struct {
		policyNamespace string
		tags            []string
		eList           gaia.ExternalNetworksList
	}

	extnetSameNs := gaia.NewExternalNetwork()
	extnetSameNs.Name = "match-same-ns"
	extnetSameNs.Namespace = "/policy/child"
	extnetSameNs.AssociatedTags = []string{"ext:network=match-same-ns", "tag=match"}

	extnetSameNsNoMatch := gaia.NewExternalNetwork()
	extnetSameNsNoMatch.Name = "match-same-ns-no-match"
	extnetSameNsNoMatch.Namespace = "/policy/child"
	extnetSameNsNoMatch.AssociatedTags = []string{"ext:network=match-same-ns-no-match", "tag=no-match"}

	extnetParentNs := gaia.NewExternalNetwork()
	extnetParentNs.Name = "match-parent-ns"
	extnetParentNs.Namespace = "/policy"
	extnetParentNs.Propagate = true
	extnetParentNs.AssociatedTags = []string{"ext:network=match-parent-ns", "tag=match"}

	extnetParentNsNoPropagate := gaia.NewExternalNetwork()
	extnetParentNsNoPropagate.Name = "parent-ns-no-propagate"
	extnetParentNsNoPropagate.Namespace = "/policy"
	extnetParentNsNoPropagate.AssociatedTags = []string{"ext:network=parent-ns-no-propagate", "tag=match"}

	extnetNonParentNs := gaia.NewExternalNetwork()
	extnetNonParentNs.Name = "non-parent-ns"
	extnetNonParentNs.Namespace = "/pol"
	extnetNonParentNs.Propagate = true
	extnetNonParentNs.AssociatedTags = []string{"ext:network=non-parent-ns", "tag=match"}

	tests := []struct {
		name           string
		args           args
		wantExtnetList gaia.ExternalNetworksList
	}{
		{
			name: "one test for complex matches",
			args: args{
				policyNamespace: "/policy/child",
				tags:            []string{"tag=match"},
				eList: gaia.ExternalNetworksList{
					extnetSameNs,
					extnetSameNsNoMatch, // Tags dont match
					extnetParentNs,
					extnetParentNsNoPropagate, // Not propagated
					extnetNonParentNs,         // Not in parent ns
				},
			},
			wantExtnetList: gaia.ExternalNetworksList{
				extnetSameNs,
				extnetParentNs,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotExtnetList := extnetsFromTags(tt.args.policyNamespace, tt.args.tags, tt.args.eList); !reflect.DeepEqual(gotExtnetList, tt.wantExtnetList) {
				t.Errorf("extnetsFromTags() = %v, want %v", gotExtnetList, tt.wantExtnetList)
			}
		})
	}
}
