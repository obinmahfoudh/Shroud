package main

import (
	"fmt"

	"shroud/internal/manager"
)

func main() {
	mngr := manager.NewManager()
	info, err := mngr.DetectActiveClient()
	if err == nil {
		fmt.Printf("%+v\n", *info)
	}
}
