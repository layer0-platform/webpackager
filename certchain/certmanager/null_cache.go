// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package certmanager

import (
	"github.com/layer0-platform/webpackager/certchain"
)

// NullCache is a dummy Cache that does nothing.
var NullCache Cache = &nullCache{}

type nullCache struct{}

func (*nullCache) Read(digest string) (*certchain.AugmentedChain, error) {
	return nil, ErrNotFound
}

func (*nullCache) ReadLatest() (*certchain.AugmentedChain, error) {
	return nil, ErrNotFound
}

func (*nullCache) Write(ac *certchain.AugmentedChain) error {
	return nil
}
