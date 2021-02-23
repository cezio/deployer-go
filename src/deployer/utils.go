package deployer

import (
	"os"
)

// IsDirectory returns true if given path is a directory
func IsDirectory(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()

}
