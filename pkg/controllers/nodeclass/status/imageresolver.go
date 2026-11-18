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

	"github.com/absaoss/karpenter-provider-vsphere/pkg/apis/v1alpha1"
	"github.com/vmware/govmomi/object"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type imageProvider interface {
	ResolveImage(context.Context, v1alpha1.ImageSelectorTerm) (*object.VirtualMachine, error)
}

type VsphereImageReconciler struct {
	vsphereImageProvider imageProvider
}

func NewVsphereImageReconciler(provider imageProvider) *VsphereImageReconciler {
	return &VsphereImageReconciler{
		vsphereImageProvider: provider,
	}
}

func (r *VsphereImageReconciler) Reconcile(ctx context.Context, nodeClass *v1alpha1.VsphereNodeClass) (reconcile.Result, error) {
	image, err := r.vsphereImageProvider.ResolveImage(ctx, nodeClass.Spec.ImageSelector)
	if err == nil && (image == nil || image.InventoryPath == "") {
		err = fmt.Errorf("resolved image has no inventory path")
	}
	if err != nil {
		nodeClass.Status.ImageTemplateID = ""
		nodeClass.StatusConditions().SetFalse(v1alpha1.ConditionTypeImageReady, "ImageResolutionFailed", err.Error())
		return reconcile.Result{}, fmt.Errorf("resolving image template, %w", err)
	}
	if nodeClass.Status.ImageTemplateID != image.InventoryPath {
		log.FromContext(ctx).V(1).Info("discovered image template", "path", image.InventoryPath)
	}
	nodeClass.Status.ImageTemplateID = image.InventoryPath
	nodeClass.StatusConditions().SetTrue(v1alpha1.ConditionTypeImageReady)
	return reconcile.Result{}, nil
}
