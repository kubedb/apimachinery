package docker

import (
	"errors"
	"net/http"

	"github.com/appscode/go/io"
	docker "github.com/heroku/docker-registry-client/registry"
)

const (
	registryUrl = "https://registry-1.docker.io/"
)

const registrySecretPath = "/srv/docker/secrets/.dockercfg"

type RegistrySecret struct {
	Secret map[string]struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
		Auth     string `json:"auth"`
	} `json:"secret"`
}

func CheckDockerImageVersion(repository, reference string) error {

	var registrySecret RegistrySecret
	if io.IsFileExists(registrySecretPath) {
		if err := io.ReadFileAs(registrySecretPath, &registrySecret); err != nil {
			return err
		}

		for key, val := range registrySecret.Secret {
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

	if _, err := dockerRegistry.Manifest(repository, reference); err == nil {
		return nil
	}

	return errors.New("failed to verify docker image")
}
