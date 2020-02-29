package apollo

import (
	"testing"

	"github.com/astaxie/beego/validation"
)

type T1 struct {
	A string
	B bool
	C map[string]bool
}

func (t *T1) Valid(v *validation.Validation) {

}

func TestJsonSerializer(t *testing.T) {
	type args struct {
		buf []byte
		c   validation.ValidFormer
	}
	t1 := &T1{}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test-1",
			args: args{
				buf: []byte("{\"a\": \"a\", \"b\":true, \"c\":{\"1\":true}}"),
				c:   t1,
			},
			wantErr: true,
		},
		{
			name: "test-2",
			args: args{
				buf: []byte("{\"a\": \"aaaa\", \"b\":true, \"c\":{\"2\":true}}"),
				c:   t1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := JsonSerializer(tt.args.buf, tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("JsonSerializer() error = %v, wantErr %v", err, tt.wantErr)
			}
			t.Log(tt.args.c)
		})
	}
}
