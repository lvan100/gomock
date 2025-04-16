package testdata

import (
	"context"
	"testing"

	"github.com/lvan100/gomock/gomock"
	"github.com/lvan100/gomock/internal/assert"
	"github.com/lvan100/gomock/mockgen/testdata/inner"
)

func runService(s Service) {
	_, _ = s.Get(context.Background(), nil, nil)
}

func TestMock(t *testing.T) {
	r, _ := gomock.Init(t.Context())
	impl := NewServiceMockImpl(r)
	count := 0
	impl.MockGet().Handle(func(ctx context.Context, req *inner.Request, m map[string]string) (*Response, error, bool) {
		count++
		return nil, nil, count > 1
	})
	assert.Panic(t, func() {
		runService(impl)
	}, "no mock code matched")
	runService(impl)
}
