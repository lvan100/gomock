package testdata

import (
	"context"

	"github.com/lvan100/gomock/mockgen/testdata/inner"
)

type ServiceV2 interface {
	Get(ctx context.Context, req *inner.Request, params map[string]string) (*Response, error)
}
