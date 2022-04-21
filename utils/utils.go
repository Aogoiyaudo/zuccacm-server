package utils

import (
	"path"
	"time"
)

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

func SubDays(begin, end time.Time) int {
	return int(end.Sub(begin).Hours()) / 24
}

func Max(x, y int) int {
	if x < y {
		return y
	} else {
		return x
	}
}
