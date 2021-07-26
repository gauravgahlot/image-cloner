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
	"encoding/base64"
	"encoding/json"
	"io/ioutil"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"k8s.io/klog/v2"
)

// Client defines the operations that can be performed with a Docker client.
type Client interface {
	ImagePull(ctx context.Context, image string) error
	ImagePush(ctx context.Context, image string) error
	ImageTag(ctx context.Context, src, dst string) error
}

type docker struct {
	client       *client.Client
	registryAuth string
}

// volume path for registry username and password
const (
	usrPath  = "/auth/username"
	pswdPath = "/auth/password"
)

// CreateClient returns a DockerClient
func CreateClient() (Client, error) {
	c, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}

	auth := types.AuthConfig{
		Username: readData(usrPath),
		Password: readData(pswdPath),
	}
	encodedJSON, err := json.Marshal(auth)
	if err != nil {
		panic(err)
	}

	authStr := base64.StdEncoding.EncodeToString(encodedJSON)
	return &docker{client: c, registryAuth: authStr}, nil
}

// RegistryUser returns the username for provided container registry
func RegistryUser() string {
	return readData("/auth/username")
}

func readData(src string) string {
	data, err := ioutil.ReadFile(src)
	if err != nil {
		klog.Errorf("[error]: error reading data from source: %s\n%v", src, err)
		return ""
	}
	return string(data)
}
