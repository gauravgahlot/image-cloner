// Copyright 2021 The image-cloner Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import "context"

type serverModifier func(*server)

func withRegistry(reg string) serverModifier {
	return func(s *server) { s.registry = reg }
}

func withRegistryUser(user string) serverModifier {
	return func(s *server) { s.registryUser = user }
}

type mockDockerClient struct {
	ImagePullFunc func(ctx context.Context, image string) error
	ImagePushFunc func(ctx context.Context, image string) error
	ImageTagFunc  func(ctx context.Context, src, dst string) error
}

func (d mockDockerClient) ImagePull(ctx context.Context, image string) error {
	return d.ImagePullFunc(ctx, image)
}
func (d mockDockerClient) ImagePush(ctx context.Context, image string) error {
	return d.ImagePushFunc(ctx, image)
}

func (d mockDockerClient) ImageTag(ctx context.Context, src, dst string) error {
	return d.ImageTagFunc(ctx, src, dst)
}
