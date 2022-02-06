package networkpolicies

import (
	"reflect"
	"testing"
)

func Test_match(t *testing.T) {
	type args struct {
		tags    []string
		objTags []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "same set",
			args: args{
				tags:    []string{"hello"},
				objTags: []string{"hello"},
			},
			want: true,
		},
		{
			name: "sub set 1",
			args: args{
				tags:    []string{"hello"},
				objTags: []string{"hello", "world"},
			},
			want: true,
		},
		{
			name: "sub set 2",
			args: args{
				tags:    []string{"world"},
				objTags: []string{"hello", "world"},
			},
			want: true,
		},
		{
			name: "super set",
			args: args{
				tags:    []string{"hello", "world"},
				objTags: []string{"hello"},
			},
			want: false,
		},
		{
			name: "mismatch set",
			args: args{
				tags:    []string{"hello"},
				objTags: []string{"world"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := match(tt.args.tags, tt.args.objTags); got != tt.want {
				t.Errorf("match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_refHasExtNextworks(t *testing.T) {
	type args struct {
		ref []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "none",
			args: args{
				ref: []string{"hello"},
			},
			want: false,
		},
		{
			name: "one 1",
			args: args{
				ref: []string{"customer:ext:network=x"},
			},
			want: true,
		},
		{
			name: "one 2",
			args: args{
				ref: []string{"hello", "customer:ext:network=x"},
			},
			want: true,
		},
		{
			name: "all",
			args: args{
				ref: []string{"$name=x", "customer:ext:network=x"},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		e := extnetPrefix
		extnetPrefix = "customer:ext:network=x"
		t.Run(tt.name, func(t *testing.T) {
			if got := refHasExtNextworks(tt.args.ref); got != tt.want {
				t.Errorf("refHasExtNextworks() = %v, want %v", got, tt.want)
			}
		})
		extnetPrefix = e
	}
}

func Test_intersect(t *testing.T) {
	type args struct {
		a []string
		b []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "icmp 1",
			args: args{
				a: []string{"icmp/11/0"},
				b: []string{"icmp"},
			},
			want: []string{"icmp/11/0"},
		},
		{
			name: "icmp 2",
			args: args{
				a: []string{"icmp"},
				b: []string{"icmp/11/0"},
			},
			want: []string{"icmp/11/0"},
		},
		{
			name: "no first arg",
			args: args{
				a: []string{},
				b: []string{"tcp/9126"},
			},
			want: []string{"TCP/9126"},
		},
		{
			name: "no second arg",
			args: args{
				a: []string{"tcp/9126"},
				b: []string{},
			},
			want: []string{"TCP/9126"},
		},
		{
			name: "intersection 1",
			args: args{
				a: []string{"tcp/9126", "tcp/10000"},
				b: []string{"tcp/1:65535"},
			},
			want: []string{"TCP/9126", "TCP/10000"},
		},
		{
			name: "intersection 2",
			args: args{
				a: []string{"tcp/9126"},
				b: []string{"tcp/1:65535", "tcp/9126"},
			},
			want: []string{"TCP/9126"},
		},
		{
			name: "intersection 3",
			args: args{
				a: []string{"tcp/9126:9130"},
				b: []string{"tcp/1:65535", "tcp/9126"},
			},
			want: []string{"TCP/9126:9130"},
		},
		{
			name: "intersection 4",
			args: args{
				a: []string{"tcp/9126:9130", "tcp/9300:9500"},
				b: []string{"tcp/1:65535", "tcp/9126"},
			},
			want: []string{"TCP/9126:9130", "TCP/9300:9500"},
		},
		{
			name: "multiprotocol 1",
			args: args{
				a: []string{"tcp/9126:9130"},
				b: []string{"tcp/1:65535", "udp/1:65535"},
			},
			want: []string{"TCP/9126:9130"},
		},
		{
			name: "multiprotocol 21",
			args: args{
				a: []string{"tcp/9126:9130", "udp/2330"},
				b: []string{"tcp/1:65535", "udp/1:65535", "udp/100"},
			},
			want: []string{"TCP/9126:9130", "UDP/2330"},
		},
		{
			name: "multiprotocol 22",
			args: args{
				a: []string{"tcp/1:65535", "udp/1:65535"},
				b: []string{"tcp/9126:9130", "udp/2330"},
			},
			want: []string{"TCP/9126:9130", "UDP/2330"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := intersect(tt.args.a, tt.args.b); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("intersect() = %v, want %v", got, tt.want)
			}
		})
	}
}
