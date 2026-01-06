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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	resourceapply "k8s.io/client-go/applyconfigurations/resource/v1"
	clientset "k8s.io/client-go/kubernetes"
)

// Handler implements the UpdateStatus function
// to update the status of ResourceClaims via Kubernetes API.
type Handler struct {
	ClientSet  clientset.Interface
	DriverName string
}

// UpdateStatus updates the status of the given ResourceClaim via Kubernetes API.
func (sh *Handler) UpdateStatus(ctx context.Context, claim *resourcev1.ResourceClaim) error {
	statusUpdates := &resourceapply.ResourceClaimStatusApplyConfiguration{Devices: []resourceapply.AllocatedDeviceStatusApplyConfiguration{}}

	for _, device := range claim.Status.Devices {
		if device.Driver != sh.DriverName {
			continue
		}

		resourceClaimStatusDevice := resourceapply.
			AllocatedDeviceStatus().
			WithDevice(device.Device).
			WithDriver(device.Driver).
			WithPool(device.Pool)

		if device.ShareID != nil {
			resourceClaimStatusDevice.WithShareID(*device.ShareID)
		}
		if device.Data != nil {
			resourceClaimStatusDevice.WithData(*device.Data)
		}

		if device.NetworkData != nil {
			networkDataApply := resourceapply.NetworkDeviceData().
				WithHardwareAddress(device.NetworkData.HardwareAddress).
				WithIPs(device.NetworkData.IPs...).
				WithInterfaceName(device.NetworkData.InterfaceName)

			resourceClaimStatusDevice.WithNetworkData(networkDataApply)
		}

		// todo: condition

		statusUpdates.WithDevices(resourceClaimStatusDevice)
	}

	resourceClaimApply := resourceapply.ResourceClaim(claim.GetName(), claim.GetNamespace()).WithStatus(statusUpdates)
	_, err := sh.ClientSet.ResourceV1().ResourceClaims(claim.GetNamespace()).ApplyStatus(ctx,
		resourceClaimApply,
		metav1.ApplyOptions{FieldManager: sh.DriverName, Force: true},
	)
	if err != nil {
		return fmt.Errorf("failed to update resource claim status: %v", err)
	}

	return nil
}
