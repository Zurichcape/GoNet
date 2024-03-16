package timer

import (
	"fmt"
	"testing"
)

func SayHello(message ...interface{}) {
	fmt.Println(message[0].(string), " ", message[1].(string))
}

func TestDelayFunc_Call(t *testing.T) {
	type fields struct {
		f    func(...interface{})
		args []interface{}
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			"zinx",
			fields{
				f:    SayHello,
				args: []interface{}{"hello", "zinx"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			df := &DelayFunc{
				f:    tt.fields.f,
				args: tt.fields.args,
			}
			df.Call()
		})
	}
}
