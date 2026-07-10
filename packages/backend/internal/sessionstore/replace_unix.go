//go:build !windows

package sessionstore

import "os"

func replaceFile(source, target string) error {
	return os.Rename(source, target)
}
