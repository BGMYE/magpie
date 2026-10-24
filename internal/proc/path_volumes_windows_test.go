package proc

import (
	"fmt"
	"reflect"
	"testing"

	"golang.org/x/sys/windows"
)

func TestUnavailablePathVolumes(t *testing.T) {
	dirs := []string{`C:\Tools`, `E:\Python`, `e:\Other`, `F:\Tools`, `G:\Tools`, `H:\Tools`, `relative`, `\\server\tools`}
	calls := map[string]int{}
	got := unavailablePathVolumes(dirs, func(root string) error {
		calls[root]++
		switch root {
		case `E:\`:
			return fmt.Errorf("volume: %w", windows.ERROR_UNRECOGNIZED_VOLUME)
		case `F:\`:
			return windows.ERROR_NOT_READY
		case `G:\`:
			return windows.ERROR_ACCESS_DENIED
		case `H:\`:
			return windows.ERROR_PATH_NOT_FOUND
		}
		return nil
	})
	want := map[string]bool{"E:": true, "F:": true}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unavailable volumes = %v, want %v", got, want)
	}
	if len(calls) != 5 || calls[`E:\`] != 1 {
		t.Fatalf("volume probes = %v; want each local drive once, no relative or network probes", calls)
	}
}
