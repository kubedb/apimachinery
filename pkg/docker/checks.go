package docker

import (
	"net/http"

	docker "github.com/heroku/docker-registry-client/registry"
)

const (
	registryUrl = "https://registry-1.docker.io/"
)

func CheckDockerImageVersion(repository, reference, username, password string) error {
	registry := &docker.Registry{
		URL: registryUrl,
		Client: &http.Client{
			Transport: docker.WrapTransport(http.DefaultTransport, registryUrl, username, password),
		},
		Logf: docker.Quiet,
	}

	_, err := registry.Manifest(repository, reference)
	return err
}
