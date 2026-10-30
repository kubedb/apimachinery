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

package v1alpha1

import (
	"context"
	"errors"
	"fmt"

	secret_lib "kubedb.dev/apimachinery/pkg/secret"

	vsecretapi "go.virtual-secrets.dev/apimachinery/apis/virtual/v1alpha1"
	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	appcat "kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// validateAuthSecretRef validates that the auth secret referenced by a RotateAuth
// ops request exists. It is read as a virtual-secrets.dev Secret when
// ref.APIGroup == "virtual-secrets.dev", otherwise as a core/v1 Secret.
func validateAuthSecretRef(ctx context.Context, c client.Client, namespace string, ref *appcat.TypedLocalObjectReference) error {
	if ref == nil {
		return nil
	}
	if ref.Name == "" {
		return errors.New("spec.authentication.secretRef.name can not be empty")
	}
	isVirtual := ref.APIGroup == vsecretapi.GroupName
	if _, err := secret_lib.GetData(ctx, c, namespace, ref.Name, isVirtual); err != nil {
		if apierrors.IsNotFound(err) {
			return fmt.Errorf("referenced secret %s/%s not found", namespace, ref.Name)
		}
		return err
	}
	return nil
}

// authSecretUsername reads the username from an auth secret (virtual or core).
func authSecretUsername(ctx context.Context, c client.Client, namespace, name string, isVirtual bool) (string, error) {
	data, err := secret_lib.GetData(ctx, c, namespace, name, isVirtual)
	if err != nil {
		return "", err
	}
	return string(data[core.BasicAuthUsernameKey]), nil
}

// validateRotateAuthSecretRef validates the RotateAuth SecretRef: it checks the
// referenced secret exists (virtual-aware) and, when oldSecretName is non-empty,
// ensures the username matches the database's current auth secret so that a
// rotation cannot silently change the username.
func validateRotateAuthSecretRef(ctx context.Context, c client.Client, namespace string, ref *appcat.TypedLocalObjectReference, oldSecretName string, oldIsVirtual bool) error {
	if err := validateAuthSecretRef(ctx, c, namespace, ref); err != nil {
		return err
	}
	if oldSecretName == "" || ref == nil {
		return nil
	}
	newIsVirtual := ref.APIGroup == vsecretapi.GroupName
	newUser, err := authSecretUsername(ctx, c, namespace, ref.Name, newIsVirtual)
	if err != nil {
		return err
	}
	oldUser, err := authSecretUsername(ctx, c, namespace, oldSecretName, oldIsVirtual)
	if err != nil {
		return err
	}
	if oldUser != newUser {
		return errors.New("database username cannot be changed")
	}
	return nil
}
