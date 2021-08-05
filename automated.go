package main

import (
	"fmt"
	"runtime"
)

func main() {
	if runtime.GOOS == "windows" {
		fmt.Println("Can't Execute this on a windows machine")
	} else {
		backup_mysql()
	}
	print("\nFinished Running Automated Tasks\n\n")
}
