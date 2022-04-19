package utils

import "path"

func SimplePath(file, root string) string {
	tmp := file
	for len(tmp) > 0 && path.Base(tmp) != root {
		tmp = tmp[:len(tmp)-len(path.Base(tmp))]
	}
	return file[len(tmp):]
}

func IsLocalFile(file, root string) bool {
	for len(file) > 0 {
		if path.Base(file) == root {
			return true
		}
		file = file[:len(file)-len(path.Base(file))]
	}
	return false
}
