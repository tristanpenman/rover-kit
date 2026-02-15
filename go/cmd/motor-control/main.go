package main

import (
	"context"

	"rover-kit/pkg/common"
)

func main() {
	ctx := context.Background()
	driver := common.DummyDriver{}
	err := driver.Stop(ctx)
	if err != nil {
		return
	}
}
