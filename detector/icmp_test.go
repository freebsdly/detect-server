package detector

import (
	"context"
	"github.com/go-ping/ping"
	"reflect"
	"testing"
)

func TestIcmpDetector_Detect(t *testing.T) {
	type fields struct {
		options          IcmpDetectorOptions
		detectBuffer     chan DetectTarget[IcmpDetectOptions]
		resultQueue      chan DetectResult[IcmpDetectOptions, *ping.Statistics]
		parentCtx        context.Context
		parentCancelFunc context.CancelFunc
	}
	type args struct {
		target DetectTarget[IcmpDetectOptions]
	}

	var target = NewDetectTarget[IcmpDetectOptions](ICMPDetect, IcmpDetectOptions{
		Target:  "192.168.0.1",
		Count:   3,
		Timeout: 500,
	})

	tests := []struct {
		name   string
		fields fields
		args   args
		want   DetectResult[IcmpDetectOptions, *ping.Statistics]
	}{
		{
			name: "192.168.0.1",
			fields: fields{
				options:          IcmpDetectorOptions{},
				detectBuffer:     make(chan DetectTarget[IcmpDetectOptions], 10),
				resultQueue:      make(chan DetectResult[IcmpDetectOptions, *ping.Statistics], 10),
				parentCtx:        nil,
				parentCancelFunc: nil,
			},
			args: args{target: target},
			want: NewDetectResult[IcmpDetectOptions, *ping.Statistics](target, &ping.Statistics{}, nil),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := &IcmpDetector{
				options:          tt.fields.options,
				detectBuffer:     tt.fields.detectBuffer,
				resultQueue:      tt.fields.resultQueue,
				parentCtx:        tt.fields.parentCtx,
				parentCancelFunc: tt.fields.parentCancelFunc,
			}
			if got := detector.Detect(tt.args.target); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Detect() = %v, want %v", got, tt.want)
			}
		})
	}
}
