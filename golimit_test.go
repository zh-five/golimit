// 协程并发数限制库
// https://github.com/zh-five/golimit

package golimit

import (
	"testing"
	"time"
)

func TestAddDone(t *testing.T) {
	type args struct {
		max   uint64
		total int
		sleep int
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "TestNewGoLimit",
			args: args{
				max:   2,
				total: 12,
				sleep: 1,
			},
			want: 6,
		},
		{
			name: "TestNewGoLimit",
			args: args{
				max:   3,
				total: 12,
				sleep: 1,
			},
			want: 4,
		},
		{
			name: "TestNewGoLimit",
			args: args{
				max:   6,
				total: 12,
				sleep: 2,
			},
			want: 4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gl := NewGoLimit(tt.args.max)
			t1 := time.Now().Unix()

			run(gl, tt.args.total, tt.args.sleep)

			got := time.Now().Unix() - t1

			if got != tt.want {
				t.Errorf("run time = %v, want %v", got, tt.want)
			}
		})
	}
}

func run(gl *GoLimit, total int, sleep int) {
	for i := 0; i < total; i++ {
		gl.Add()
		go func() {
			defer gl.Done()
			time.Sleep(time.Second * time.Duration(sleep))
		}()
	}
	gl.WaitZero()
}
