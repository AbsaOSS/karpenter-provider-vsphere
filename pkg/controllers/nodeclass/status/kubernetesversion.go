/*
Portions Copyright (c) Microsoft Corporation.

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

package status

import (
	"context"
	"fmt"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/awslabs/operatorpkg/reasonable"
	semver "github.com/blang/semver/v4"

	"github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
	"github.com/absaoss/karpenter-provider-vsphere/pkg/providers/kubernetesversion"

	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/karpenter/pkg/utils/pretty"
)

const (
	kubernetesVersionReconcilerName = "nodeclass.kubernetesversion"
)

type KubernetesVersionReconciler struct {
	kubernetesVersionProvider kubernetesversion.KubernetesVersionProvider
	cm                        *pretty.ChangeMonitor
}

func NewKubernetesVersionReconciler(provider kubernetesversion.KubernetesVersionProvider) *KubernetesVersionReconciler {
	return &KubernetesVersionReconciler{
		kubernetesVersionProvider: provider,
		cm:                        pretty.NewChangeMonitor(),
	}
}

func (r *KubernetesVersionReconciler) Register(_ context.Context, m manager.Manager) error {
	return controllerruntime.NewControllerManagedBy(m).
		Named(kubernetesVersionReconcilerName).
		For(&v1alpha1.VsphereNodeClass{}).
		WithOptions(controller.Options{
			RateLimiter:             reasonable.RateLimiter(),
			MaxConcurrentReconciles: 10,
		}).
		Complete(reconcile.AsReconciler(m.GetClient(), r))
}

func (r *KubernetesVersionReconciler) Reconcile(ctx context.Context, nodeClass *v1alpha1.VsphereNodeClass) (reconcile.Result, error) {
	ctx = log.IntoContext(ctx, log.FromContext(ctx).WithName(kubernetesVersionReconcilerName))
	logger := log.FromContext(ctx)
	logger.V(1).Info("starting reconcile")

	goalK8sVersion, err := r.kubernetesVersionProvider.KubeServerVersion(ctx)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("getting kubernetes version, %w", err)
	}

	// Handles case 1: init, update kubernetes status to API server version found
	if !nodeClass.StatusConditions().Get(v1alpha1.ConditionTypeKubernetesVersionReady).IsTrue() || nodeClass.Status.KubernetesVersion == "" {
		logger.Info(fmt.Sprintf("init kubernetes version: %s", goalK8sVersion))
	} else {
		// Check if there is an upgrade
		newK8sVersion, err := semver.Parse(goalK8sVersion)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("parsing discovered kubernetes version, %w", err)
		}
		currentK8sVersion, err := semver.Parse(nodeClass.Status.KubernetesVersion)
		if err != nil {
			return reconcile.Result{}, fmt.Errorf("parsing current kubernetes version, %w", err)
		}
		// Handles case 2: Upgrade kubernetes version [Note: we set node image to not ready, since we upgrade node image when there is a kubernetes upgrade]
		if newK8sVersion.GT(currentK8sVersion) {
			logger.Info(fmt.Sprintf("kubernetes upgrade detected: from %s (current), to %s (discovered)", currentK8sVersion.String(), newK8sVersion.String()))
		} else if newK8sVersion.LT(currentK8sVersion) {
			logger.Info(fmt.Sprintf("detected potential kubernetes downgrade: from %s (current), to %s (discovered)", currentK8sVersion.String(), newK8sVersion.String()))
			// We do not currently support downgrading, so keep the kubernetes version the same
			goalK8sVersion = nodeClass.Status.KubernetesVersion
		}
	}
	nodeClass.Status.KubernetesVersion = goalK8sVersion
	nodeClass.StatusConditions().SetTrue(v1alpha1.ConditionTypeKubernetesVersionReady)
	logger.V(1).Info("successful reconcile")
	return reconcile.Result{RequeueAfter: 15 * time.Minute}, nil
}
