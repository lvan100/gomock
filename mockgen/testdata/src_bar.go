/*
 * Copyright 2025 The Go-Spring Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package testdata

import (
	"context"

	"github.com/lvan100/gomock/mockgen/testdata/inner"
)

type RepositoryV2[T ~int | ~uint] interface {
	Save(item T) error
	FindByID(id string) (T, error)
}

type ServiceV2 interface {
	Get(ctx context.Context, req *inner.Request, params map[string]string) (*Response, error)
}
