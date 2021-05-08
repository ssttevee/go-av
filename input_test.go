package av_test

import (
	"net/http"
	"runtime"
	"testing"

	"github.com/ssttevee/go-av"
)

func TestOpenInputReader(t *testing.T) {
	res, err := http.Get("https://archive.org/download/BigBuckBunny_124/Content/big_buck_bunny_720p_surround.mp4")
	if err != nil {
		t.Fatal(err)
	}

	defer res.Body.Close()

	if _, err = av.OpenInputReader(res.Body); err != nil {
		t.Fatal(err)
	}

	// invoke gc to test finalizer
	runtime.GC()
}
