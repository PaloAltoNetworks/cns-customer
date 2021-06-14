package utils

import "testing"

func TestSetupNamespaceString(t *testing.T) {
	type args struct {
		namespaces []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "single namespace",
			args: args{namespaces: []string{"hello"}},
			want: "/hello",
		},
		{
			name: "two namespace",
			args: args{namespaces: []string{"hello", "world"}},
			want: "/hello/world",
		},
		{
			name: "namespaces with / in prefix and suffix",
			args: args{namespaces: []string{"/hello/", "/world/"}},
			want: "/hello/world",
		},
		{
			name: "namespaces with multiple / in prefix and suffix",
			args: args{namespaces: []string{"//hello/", "/world//"}},
			want: "/hello/world",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SetupNamespaceString(tt.args.namespaces...); got != tt.want {
				t.Errorf("SetupNamespaceString() = %v, want %v", got, tt.want)
			}
		})
	}
}
