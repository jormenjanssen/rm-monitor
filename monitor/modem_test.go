package main

import (
	"reflect"
	"testing"
)

func TestTranslateModemDBM(t *testing.T) {
	type args struct {
		rawValue int
		berValue int
	}
	tests := []struct {
		name string
		args args
		want SignalStrength
	}{
		{name: "No Signal test", args: args{rawValue: 0, berValue: 99}, want: NoSignal},
		{name: "Weak Signal test", args: args{rawValue: 4, berValue: 99}, want: WeakSignal},
		{name: "Fair Signal test", args: args{rawValue: 8, berValue: 99}, want: FairSignal},
		{name: "Good Signal test", args: args{rawValue: 13, berValue: 99}, want: GoodSignal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TranslateModemDBM(tt.args.rawValue, tt.args.berValue); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TranslateModemDBM() = %v, want %v", got, tt.want)
			}
		})
	}
}
