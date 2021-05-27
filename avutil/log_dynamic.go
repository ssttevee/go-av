// +build av_dynamic

package avutil

func init() {
	initFuncs = append(initFuncs, InitLogging)
}

func InitLogging() {
	dynamicInit2()
	initLogging()
}
