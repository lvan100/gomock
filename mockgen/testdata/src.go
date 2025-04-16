package testdata

import (
	"context"

	"github.com/lvan100/gomock/mockgen/testdata/inner"
)

//go:generate /Users/didi/gomock/mockgen/mockgen -o src_mock.go -t '!ServiceV2'

type Response struct{}

type Service interface {
	Get(ctx context.Context, req *inner.Request, params map[string]string) (*Response, error)
}
