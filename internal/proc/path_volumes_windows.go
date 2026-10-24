package proc

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/sys/windows"
)

var pathVolumes struct {
	sync.Once
	unavailable map[string]bool
}

// Probe drive roots once before discovery; probing every missing executable
// on an unrecognized volume can block every agent's initial state read.
func availablePathDirs(dirs []string) []string {
	pathVolumes.Do(func() {
		all := append(filepath.SplitList(os.Getenv("PATH")), loginPath()...)
		pathVolumes.unavailable = unavailablePathVolumes(all, volumeError)
	})
	var out []string
	for _, dir := range dirs {
		if !pathVolumes.unavailable[strings.ToUpper(filepath.VolumeName(dir))] {
			out = append(out, dir)
		}
	}
	return out
}

func unavailablePathVolumes(dirs []string, inspect func(string) error) map[string]bool {
	seen, unavailable := map[string]bool{}, map[string]bool{}
	for _, dir := range dirs {
		volume := strings.ToUpper(filepath.VolumeName(dir))
		if len(volume) != 2 || volume[1] != ':' || seen[volume] {
			continue
		}
		seen[volume] = true
		err := inspect(volume + `\`)
		if errors.Is(err, windows.ERROR_UNRECOGNIZED_VOLUME) || errors.Is(err, windows.ERROR_NOT_READY) {
			unavailable[volume] = true
		}
	}
	return unavailable
}

func volumeError(root string) error {
	path, err := windows.UTF16PtrFromString(root)
	if err != nil {
		return err
	}
	return windows.GetVolumeInformation(path, nil, 0, nil, nil, nil, nil, 0)
}
