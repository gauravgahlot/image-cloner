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

package docker

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/jsonmessage"
	"k8s.io/klog/v2"
)

func (d *docker) ImagePull(ctx context.Context, image string) error {
	res, err := d.client.ImagePull(context.Background(), image, types.ImagePullOptions{})
	if res != nil {
		defer res.Close()
	}
	if err != nil {
		return err
	}

	if err = d.watch(res); err != nil {
		return err
	}
	return nil
}

func (d *docker) ImagePush(ctx context.Context, image string) error {
	res, err := d.client.ImagePush(context.Background(), image,
		types.ImagePushOptions{
			RegistryAuth: d.registryAuth,
		})
	if res != nil {
		defer res.Close()
	}
	if err != nil {
		return err
	}

	if err = d.watch(res); err != nil {
		return err
	}
	return nil
}

func (d *docker) ImageTag(ctx context.Context, src, dst string) error {
	err := d.client.ImageTag(context.Background(), src, dst)
	if err != nil {
		return err
	}

	klog.Infof("[info]: '%s' successfully tagged as '%s'\n", src, dst)
	return nil
}

func (d *docker) watch(in io.Reader) error {
	dec := json.NewDecoder(in)
	status := ""

	for {
		var jm jsonmessage.JSONMessage
		if err := dec.Decode(&jm); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if jm.Error != nil {
			return jm.Error
		}
		if len(jm.ErrorMessage) > 0 {
			return errors.New(jm.ErrorMessage)
		}

		if jm.Status != "" && !strings.EqualFold(status, jm.Status) {
			klog.Infof("[info]: %v\n", jm.Status)
			status = jm.Status
		}
	}
	return nil
}
