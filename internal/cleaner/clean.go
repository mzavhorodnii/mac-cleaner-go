package cleaner

import "os"

func Clean(path string) error {
	return os.RemoveAll(path)
}
