// +build av_dynamic

package avformat

import "github.com/ssttevee/go-av/avutil"

func init() {
	initFuncs = append(initFuncs, avutil.InitLogging)
}
