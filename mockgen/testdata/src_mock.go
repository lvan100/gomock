package testdata

import (
	"context"
	"reflect"

	"github.com/lvan100/gomock/gomock"
	"github.com/lvan100/gomock/mockgen/testdata/inner"
)

type ServiceMockImpl struct {
	r *gomock.Manager
}

func NewServiceMockImpl(r *gomock.Manager) *ServiceMockImpl {
	return &ServiceMockImpl{r}
}

func (impl *ServiceMockImpl) Get(ctx context.Context, req *inner.Request, params map[string]string) (*Response, error) {
	t := reflect.TypeFor[ServiceMockImpl]()
	if ret, ok := gomock.Invoke(impl.r, t, "Get", ctx, req, params); ok {
		return gomock.Unbox2[*Response, error](ret)
	}
	panic("no mock code matched")
}

func (impl *ServiceMockImpl) MockGet() *gomock.Mocker32[context.Context, *inner.Request, map[string]string, *Response, error] {
	t := reflect.TypeFor[ServiceMockImpl]()
	return gomock.NewMocker32[context.Context, *inner.Request, map[string]string, *Response, error](impl.r, t, "Get")
}
