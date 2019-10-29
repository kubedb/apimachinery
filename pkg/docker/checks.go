/*
Copyright The KubeDB Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package docker

import (
	"net/http"

	docker "github.com/appscode/docker-registry-client/registry"
	"github.com/appscode/go/ioutil"
	"github.com/pkg/errors"
)

const (
	registryUrl = "https://registry-1.docker.io/"
)

const dockerConfigPath = "/srv/docker/secrets/.dockercfg"

type RegistrySecret struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Auth     string `json:"auth"`
}

func CheckDockerImageVersion(repository, reference string) error {
	if ioutil.IsFileExists(dockerConfigPath) {
		registrySecret := make(map[string]RegistrySecret)
		if err := ioutil.ReadFileAs(dockerConfigPath, &registrySecret); err != nil {
			return err
		}
		for key, val := range registrySecret {
			dockerRegistry := &docker.Registry{
				URL: key,
				Client: &http.Client{
					Transport: docker.WrapTransport(http.DefaultTransport, key, val.Username, val.Password),
				},
				Logf: docker.Quiet,
			}
			_, err := dockerRegistry.Manifest(repository, reference)
			if err == nil {
				return nil
			}
		}
	}

	dockerRegistry := &docker.Registry{
		URL: registryUrl,
		Client: &http.Client{
			Transport: docker.WrapTransport(http.DefaultTransport, registryUrl, "", ""),
		},
		Logf: docker.Quiet,
	}

	if _, err := dockerRegistry.Manifest(repository, reference); err != nil {
		return errors.New("failed to verify docker image")
	}
	return nil
}
