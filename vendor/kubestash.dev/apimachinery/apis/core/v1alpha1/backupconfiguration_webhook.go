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
	"fmt"

	"kubestash.dev/apimachinery/apis"
	storageapi "kubestash.dev/apimachinery/apis/storage/v1alpha1"

	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	restclient "k8s.io/client-go/rest"
	kmapi "kmodules.xyz/client-go/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var backupconfigurationlog = logf.Log.WithName("backupconfiguration-resource")

func (b *BackupConfiguration) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(b).
		Complete()
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-core-kubestash-com-v1alpha1-backupconfiguration,mutating=false,failurePolicy=fail,sideEffects=None,groups=core.kubestash.com,resources=backupconfigurations,verbs=create;update,versions=v1alpha1,name=vbackupconfiguration.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &BackupConfiguration{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (b *BackupConfiguration) ValidateCreate() error {
	backupconfigurationlog.Info("validate create", apis.KeyName, b.Name)

	c, err := getNewRuntimeClient()
	if err != nil {
		return fmt.Errorf("failed to set Kubernetes client, Reason: %w", err)
	}

	if err := b.validateRepositories(context.Background(), c); err != nil {
		return err
	}
	if err := b.validateBackendsAgainstUsagePolicy(context.Background(), c); err != nil {
		return err
	}

	return b.validateHookTemplatesAgainstUsagePolicy(context.Background(), c)
}

func getNewRuntimeClient() (client.Client, error) {
	config, err := restclient.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get Kubernetes config. Reason: %w", err)
	}
	scheme := runtime.NewScheme()
	utilruntime.Must(storageapi.AddToScheme(scheme))
	utilruntime.Must(core.AddToScheme(scheme))
	utilruntime.Must(AddToScheme(scheme))

	mapper, err := apiutil.NewDynamicRESTMapper(config)
	if err != nil {
		return nil, err
	}

	return client.New(config, client.Options{
		Scheme: scheme,
		Mapper: mapper,
		Opts: client.WarningHandlerOptions{
			SuppressWarnings:   false,
			AllowDuplicateLogs: false,
		},
	})
}

func (b *BackupConfiguration) validateRepositories(ctx context.Context, c client.Client) error {
	if err := b.validateRepositoryNameUnique(); err != nil {
		return err
	}
	return b.validateRepositoryReferences(ctx, c)
}

func (b *BackupConfiguration) validateRepositoryReferences(ctx context.Context, c client.Client) error {
	for _, session := range b.Spec.Sessions {
		for _, repo := range session.Repositories {
			if !b.backendMatched(repo) {
				return fmt.Errorf("backend %q for repository %q doesn't match with any of the given backends", repo.Backend, repo.Name)
			}

			existingRepo, err := b.getRepository(ctx, c, repo.Name)
			if err != nil {
				if kerr.IsNotFound(err) {
					continue
				}
				return err
			}

			if !targetMatched(&existingRepo.Spec.AppRef, b.GetTargetRef()) {
				return fmt.Errorf("repository '%q' already exists in the cluster with a different target reference. Please, choose a different repository name", repo.Name)
			}

			if !storageRefMatched(b.GetStorageRef(repo.Backend), &existingRepo.Spec.StorageRef) {
				return fmt.Errorf("repository '%q' already exists in the cluster with a different storage reference. Please, choose a different repository name", repo.Name)
			}

		}
	}
	return nil
}

func storageRefMatched(b1, b2 *kmapi.ObjectReference) bool {
	return b1.Name == b2.Name && b1.Namespace == b2.Namespace
}

func targetMatched(t1, t2 *kmapi.TypedObjectReference) bool {
	return t1.APIGroup == t2.APIGroup &&
		t1.Kind == t2.Kind &&
		t1.Namespace == t2.Namespace &&
		t1.Name == t2.Name
}

func (b *BackupConfiguration) validateRepositoryNameUnique() error {
	repoMap := make(map[string]struct{})

	for _, session := range b.Spec.Sessions {
		for _, repo := range session.Repositories {
			if _, ok := repoMap[repo.Name]; ok {
				return fmt.Errorf("duplicate repository name found: %q. Please choose a different repository name", repo.Name)
			}
			repoMap[repo.Name] = struct{}{}
		}
	}
	return nil
}

func (b *BackupConfiguration) backendMatched(repo RepositoryInfo) bool {
	for _, b := range b.Spec.Backends {
		if b.Name == repo.Backend {
			return true
		}
	}
	return false
}

func (b *BackupConfiguration) getRepository(ctx context.Context, c client.Client, name string) (*storageapi.Repository, error) {
	repo := &storageapi.Repository{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: b.Namespace,
		},
	}

	if err := c.Get(ctx, client.ObjectKeyFromObject(repo), repo); err != nil {
		return nil, err
	}
	return repo, nil
}

func (b *BackupConfiguration) validateBackendsAgainstUsagePolicy(ctx context.Context, c client.Client) error {
	for _, backend := range b.Spec.Backends {
		bs, err := b.getBackupStorage(ctx, c, backend.StorageRef)
		if err != nil {
			if kerr.IsNotFound(err) {
				continue
			}
			return err
		}

		ns := &core.Namespace{ObjectMeta: v1.ObjectMeta{Name: b.Namespace}}
		if err := c.Get(ctx, client.ObjectKeyFromObject(ns), ns); err != nil {
			return err
		}

		if !bs.UsageAllowed(ns) {
			return fmt.Errorf("namespace %q is not allowed to refer BackupStorage %s/%s. Please, check the `usagePolicy` of the BackupStorage", b.Namespace, bs.Name, bs.Namespace)
		}
	}
	return nil
}

func (b *BackupConfiguration) getBackupStorage(ctx context.Context, c client.Client, ref kmapi.ObjectReference) (*storageapi.BackupStorage, error) {
	bs := &storageapi.BackupStorage{
		ObjectMeta: v1.ObjectMeta{
			Name:      ref.Name,
			Namespace: ref.Namespace,
		},
	}

	if bs.Namespace == "" {
		bs.Namespace = b.Namespace
	}

	if err := c.Get(ctx, client.ObjectKeyFromObject(bs), bs); err != nil {
		return nil, err
	}
	return bs, nil
}

func (b *BackupConfiguration) validateHookTemplatesAgainstUsagePolicy(ctx context.Context, c client.Client) error {
	hookTemplates := b.getHookTemplates()
	for _, ht := range hookTemplates {
		err := c.Get(ctx, client.ObjectKeyFromObject(&ht), &ht)
		if err != nil {
			if kerr.IsNotFound(err) {
				continue
			}
			return err
		}

		ns := &core.Namespace{ObjectMeta: v1.ObjectMeta{Name: b.Namespace}}
		if err := c.Get(ctx, client.ObjectKeyFromObject(ns), ns); err != nil {
			return err
		}

		if !ht.UsageAllowed(ns) {
			return fmt.Errorf("namespace %q is not allowed to refer HookTemplate %s/%s. Please, check the `usagePolicy` of the HookTemplate", b.Namespace, ht.Name, ht.Namespace)
		}
	}
	return nil
}

func (b *BackupConfiguration) getHookTemplates() []HookTemplate {
	var hookTemplates []HookTemplate
	for _, session := range b.Spec.Sessions {
		if session.Hooks != nil {
			hookTemplates = append(hookTemplates, b.getHookTemplatesFromHookInfo(session.Hooks.PreBackup)...)
			hookTemplates = append(hookTemplates, b.getHookTemplatesFromHookInfo(session.Hooks.PostBackup)...)
		}
	}
	return hookTemplates
}

func (b *BackupConfiguration) getHookTemplatesFromHookInfo(hooks []HookInfo) []HookTemplate {
	var hookTemplates []HookTemplate
	for _, hook := range hooks {
		if hook.HookTemplate != nil {
			hookTemplates = append(hookTemplates, HookTemplate{
				ObjectMeta: v1.ObjectMeta{
					Name:      hook.HookTemplate.Name,
					Namespace: hook.HookTemplate.Namespace,
				},
			})
		}
	}
	return hookTemplates
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (b *BackupConfiguration) ValidateUpdate(old runtime.Object) error {
	backupconfigurationlog.Info("validate update", apis.KeyName, b.Name)
	c, err := getNewRuntimeClient()
	if err != nil {
		return fmt.Errorf("failed to set Kubernetes client. Reason: %w", err)
	}

	if err := b.validateRepositories(context.Background(), c); err != nil {
		return err
	}
	if err := b.validateBackendsAgainstUsagePolicy(context.Background(), c); err != nil {
		return err
	}
	return b.validateHookTemplatesAgainstUsagePolicy(context.Background(), c)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (b *BackupConfiguration) ValidateDelete() error {
	backupconfigurationlog.Info("validate delete", apis.KeyName, b.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
