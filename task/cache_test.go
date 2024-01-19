package task

import (
	"reflect"
	"testing"
)

func Test_dedupe(t *testing.T) {
	type args struct {
		in0 []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "dedupe",
			args: args{
				in0: []string{"a", "b", "a", "c", "b"},
			},
			want: []string{"a", "b", "c"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dedupe(tt.args.in0); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("dedupe() = %v, want %v", got, tt.want)
			}
		})
	}
}
