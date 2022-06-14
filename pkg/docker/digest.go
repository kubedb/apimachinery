/*
Copyright AppsCode Inc. and Contributors

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
	"context"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/awslabs/amazon-ecr-credential-helper/ecr-login"
	"github.com/chrismellard/docker-credential-acr-env/pkg/credhelper"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/authn/github"
	"github.com/google/go-containerregistry/pkg/authn/k8schain"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/v1/google"
	"k8s.io/client-go/kubernetes"
)

var (
	amazonKeychain = authn.NewKeychainFromHelper(ecr.NewECRHelper(ecr.WithLogger(ioutil.Discard)))
	azureKeychain  = authn.NewKeychainFromHelper(credhelper.NewACRCredentialsHelper())
)

func ImageWithDigest(kc kubernetes.Interface, image string, k8sOpts *k8schain.Options) (string, error) {
	// Drop the "@sha256:hash_string" part, if any
	image, err := ImageWithoutDigest(image)
	if err != nil {
		return "", err
	}

	keyChain, err := CreateKeyChain(context.TODO(), kc, k8sOpts)
	if err != nil {
		return "", err
	}

	digest, err := crane.Digest(image, crane.WithAuthFromKeychain(keyChain))
	if err != nil {
		return "", err
	}

	return image + "@" + digest, nil
}

// CreateKeyChain a multi keychain based in input arguments
func CreateKeyChain(ctx context.Context, client kubernetes.Interface, k8sOpts *k8schain.Options) (authn.Keychain, error) {
	if k8sOpts != nil {
		kChain, err := k8schain.New(ctx, client, *k8sOpts)
		if err != nil {
			return nil, err
		}
		return authn.NewMultiKeychain(kChain, authn.DefaultKeychain), nil
	}
	return authn.NewMultiKeychain(
		google.Keychain,
		authn.DefaultKeychain,
		github.Keychain,
		amazonKeychain,
		azureKeychain,
	), nil
}

// ImageWithoutDigest takes image as input, return image without the digest value
func ImageWithoutDigest(image string) (string, error) {
	if before, _, found := strings.Cut(image, "@"); found {
		if len(before) > 0 {
			return before, nil
		}
		return "", fmt.Errorf("invalid image: %s", image)
	}
	return image, nil
}
