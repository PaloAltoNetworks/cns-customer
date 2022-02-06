package portranges

import (
	"reflect"
	"sort"
	"testing"
)

func Test_buildRanges(t *testing.T) {
	type args struct {
		ports []int
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "0 length",
			args: args{
				ports: []int{},
			},
			want: []string{},
		},
		{
			name: "1 length",
			args: args{
				ports: []int{5},
			},
			want: []string{"5"},
		},

		{
			name: "1 continuous range",
			args: args{
				ports: []int{5, 6, 7, 8, 9, 10},
			},
			want: []string{"5:10"},
		},
		{
			name: "n groups close hops",
			args: args{
				ports: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 12, 13, 14, 15, 17, 19},
			},
			want: []string{"1:10", "12:15", "17", "19"},
		},
		{
			name: "n groups",
			args: args{
				ports: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 500, 65533, 65534, 65535},
			},
			want: []string{"1:10", "500", "65533:65535"},
		},
		{
			name: "n groups first not part of sequence",
			args: args{
				ports: []int{1, 4, 5, 6, 7, 8, 9, 10, 500, 65533, 65534, 65535},
			},
			want: []string{"1", "4:10", "500", "65533:65535"},
		},
		{
			name: "n groups last not part of sequence",
			args: args{
				ports: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 500, 65535},
			},
			want: []string{"1:10", "500", "65535"},
		},
		{
			name: "n groups first and last not part of sequence",
			args: args{
				ports: []int{1, 4, 5, 6, 7, 8, 9, 10, 500, 65533, 65535},
			},
			want: []string{"1", "4:10", "500", "65533", "65535"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildRanges(tt.args.ports); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildRanges() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrimPortRange(t *testing.T) {
	type args struct {
		sports          string
		filteredPortMap map[int]struct{}
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "Single entry Port map with invalid sport",
			args: args{
				filteredPortMap: map[int]struct{}{
					22: {},
				},
				sports: ":23",
			},
			want:    []string{},
			wantErr: true,
		},
		{
			name: "Single entry Port map with max sport",
			args: args{
				filteredPortMap: map[int]struct{}{
					22: {},
				},
				sports: "1:65535",
			},
			want: []string{"22"},
		},
		{
			name: "Single entry Port map with invalid entry",
			args: args{
				filteredPortMap: map[int]struct{}{
					0: {},
				},
				sports: "1:65535",
			},
			want: []string{},
		},
		{
			name: "Empty Port map with sport range",
			args: args{
				filteredPortMap: map[int]struct{}{},
				sports:          "1:23",
			},
			want: []string{"1:23"},
		},
		{
			name: "One element Port Map with sport range",
			args: args{
				filteredPortMap: map[int]struct{}{
					21: {},
				},
				sports: "1:23",
			},
			want: []string{"21"},
		},
		{
			name: "Contiguos elements Port Map with sport range",
			args: args{
				filteredPortMap: map[int]struct{}{
					2: {},
					3: {},
					4: {},
				},
				sports: "1:23",
			},
			want: []string{"2:4"},
		},
		{
			name: "Non contiguous Port Map with sport range",
			args: args{
				filteredPortMap: map[int]struct{}{
					2:  {},
					3:  {},
					4:  {},
					7:  {},
					8:  {},
					10: {},
				},
				sports: "1:23",
			},
			want: []string{"2:4", "7:8", "10"},
		},
		{
			name: "Null set intersection with Port Map with sport range",
			args: args{
				filteredPortMap: map[int]struct{}{
					50: {},
				},
				sports: "1:2",
			},
			want: []string{},
		},
		{
			name: "Empty Port map with single sport",
			args: args{
				filteredPortMap: map[int]struct{}{},
				sports:          "50",
			},
			want: []string{"50"},
		},
		{
			name: "One element Port Map with single sport",
			args: args{
				filteredPortMap: map[int]struct{}{
					50: {},
				},
				sports: "50",
			},
			want: []string{"50"},
		},
		{
			name: "Contiguos elements Port Map with single sport",
			args: args{
				filteredPortMap: map[int]struct{}{
					48: {},
					49: {},
					50: {},
				},
				sports: "50",
			},
			want: []string{"50"},
		},
		{
			name: "Non contiguous Port Map with single sport",
			args: args{
				filteredPortMap: map[int]struct{}{
					2:  {},
					3:  {},
					4:  {},
					7:  {},
					8:  {},
					10: {},
					50: {},
				},
				sports: "50",
			},
			want: []string{"50"},
		},
		{
			name: "Null set intersection with Port Map with single sport",
			args: args{
				filteredPortMap: map[int]struct{}{
					10: {},
				},
				sports: "50",
			},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TrimPortRange(tt.args.sports, tt.args.filteredPortMap)
			if (err != nil) != tt.wantErr {
				t.Errorf("TrimPortRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TrimPortRange() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreatePortList(t *testing.T) {
	type args struct {
		portMap map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "empty port map",
			args: args{},
			want: []string{},
		},
		{
			name: "one member port map",
			args: args{
				portMap: map[string]interface{}{
					"1": struct{}{},
				},
			},
			want: []string{"1"},
		},
		{
			name: "n member port map",
			args: args{
				portMap: map[string]interface{}{
					"100": struct{}{},
					"239": struct{}{},
				},
			},
			want: []string{"100", "239"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreatePortList(tt.args.portMap)
			sort.StringSlice(got).Sort()
			sort.StringSlice(tt.want).Sort()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreatePortList() = %v, want %v", got, tt.want)
			}
		})
	}
}
