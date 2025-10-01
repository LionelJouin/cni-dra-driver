/*
Copyright 2025 The Kubernetes Authors.

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

// Package status implements the UpdateStatus function
// to update the status of ResourceClaims via Kubernetes API.
package status

import (
	"context"
	"fmt"

	resourcev1 "k8s.io/api/resource/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
)

// Handler implements the UpdateStatus function
// to update the status of ResourceClaims via Kubernetes API.
type Handler struct {
	ClientSet clientset.Interface
}

// UpdateStatus updates the status of the given ResourceClaim via Kubernetes API.
func (sh *Handler) UpdateStatus(ctx context.Context, claim *resourcev1.ResourceClaim) error {
	_, err := sh.ClientSet.ResourceV1().ResourceClaims(claim.GetNamespace()).UpdateStatus(ctx, claim, v1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update resource claim status: %v", err)
	}

	return nil
}
