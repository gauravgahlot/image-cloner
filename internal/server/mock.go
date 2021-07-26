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
