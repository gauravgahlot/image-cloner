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

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/gauravgahlot/image-cloner/internal/docker"
	"github.com/stretchr/testify/assert"
)

var uid types.UID = "4584308f-b307-455b-ab11-5765b4548b71"

const (
	alpine       = "alpine:3.12"
	registryUser = "gauravgahlot"
	registry     = "quay.io"
)

type args struct {
	dc         mockDockerClient
	mods       []serverModifier
	containers []v1.Container
}

func TestCreateResponse(t *testing.T) {
	type want struct {
		err bool
		res reviewResponse
	}

	cases := map[string]struct {
		args
		want
	}{
		"image-with-backup-registry": {
			args: args{
				dc:   mockDockerClient{},
				mods: []serverModifier{withRegistryUser(registryUser), withRegistry(registry)},
				containers: []v1.Container{
					{
						Image: strings.Join([]string{registry, registryUser, alpine}, "/"),
					},
				},
			},
			want: want{res: reviewResponse{uid: uid, allowed: true, patch: nil}},
		},
		"image-with-backup-registry-user": {
			args: args{
				dc:   mockDockerClient{},
				mods: []serverModifier{withRegistryUser(registryUser)},
				containers: []v1.Container{
					{
						Image: strings.Join([]string{registryUser, alpine}, "/"),
					},
				},
			},
			want: want{res: reviewResponse{uid: uid, allowed: true, patch: nil}},
		},
		"error-image-pull": {
			args: args{
				dc: mockDockerClient{
					ImagePullFunc: func(ctx context.Context, image string) error {
						return errors.New("error image pull")
					},
				},
				mods:       []serverModifier{withRegistryUser(registryUser)},
				containers: []v1.Container{{Image: alpine}},
			},
			want: want{
				err: true,
				res: errorResponse(errCreatingPatch),
			},
		},
		"error-image-tag": {
			args: args{
				dc: mockDockerClient{
					ImagePullFunc: func(ctx context.Context, image string) error { return nil },
					ImageTagFunc: func(ctx context.Context, src, dst string) error {
						return errors.New("error image tag")
					},
				},
				mods:       []serverModifier{withRegistryUser(registryUser)},
				containers: []v1.Container{{Image: alpine}},
			},
			want: want{
				err: true,
				res: errorResponse(errCreatingPatch),
			},
		},
		"error-image-push": {
			args: args{
				dc: mockDockerClient{
					ImagePullFunc: func(ctx context.Context, image string) error { return nil },
					ImageTagFunc:  func(ctx context.Context, src, dst string) error { return nil },
					ImagePushFunc: func(ctx context.Context, image string) error {
						return errors.New("error image push")
					},
				},
				mods:       []serverModifier{withRegistryUser(registryUser)},
				containers: []v1.Container{{Image: alpine}},
			},
			want: want{
				err: true,
				res: errorResponse(errCreatingPatch),
			},
		},
		"success-image-pull-tag-push-with-user": {
			args: args{
				dc: mockDockerClient{
					ImagePullFunc: func(ctx context.Context, image string) error { return nil },
					ImageTagFunc:  func(ctx context.Context, src, dst string) error { return nil },
					ImagePushFunc: func(ctx context.Context, image string) error { return nil },
				},
				mods:       []serverModifier{withRegistryUser(registryUser)},
				containers: []v1.Container{{Image: alpine}},
			},
			want: want{res: reviewResponse{uid: uid, allowed: true, patch: getPatch(alpine, "", registryUser)}},
		},
		"success-image-pull-tag-push-with-registry": {
			args: args{
				dc: mockDockerClient{
					ImagePullFunc: func(ctx context.Context, image string) error { return nil },
					ImageTagFunc:  func(ctx context.Context, src, dst string) error { return nil },
					ImagePushFunc: func(ctx context.Context, image string) error { return nil },
				},
				mods:       []serverModifier{withRegistryUser(registryUser), withRegistry(registry)},
				containers: []v1.Container{{Image: alpine}},
			},
			want: want{res: reviewResponse{uid: uid, allowed: true, patch: getPatch(alpine, registry, registryUser)}},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), maxWebhookTimeout*time.Second)
	defer cancel()

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			s := testServer(t, tc.args.dc, tc.args.mods...)
			res, err := s.createResponse(ctx, tc.args.containers, uid)
			if err != nil {
				assert.True(t, tc.want.err)
				assert.Error(t, err)
			}
			assert.NotNil(t, res)
			assert.Equal(t, tc.want.res, res)
		})
	}
}

func testServer(t *testing.T, d docker.Client, modifiers ...serverModifier) *server {
	s := &server{
		client: d,
	}

	for _, fn := range modifiers {
		fn(s)
	}
	return s
}

func errorResponse(msg string) reviewResponse {
	return reviewResponse{
		uid:     uid,
		allowed: false,
		status: status{
			code:    500,
			reason:  metav1.StatusReasonInternalError,
			message: msg,
		},
	}
}

func getPatch(src, reg, user string) []byte {
	list := []patch{
		{
			Op:    "replace",
			Path:  fmt.Sprintf("/spec/template/spec/containers/%d/image", 0),
			Value: newImage(src, reg, user),
		},
	}
	patch, _ := json.Marshal(list)
	return patch
}
