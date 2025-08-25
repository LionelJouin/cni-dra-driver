/*
Copyright 2024 The Kubernetes Authors.

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
	resourcev1 "k8s.io/api/resource/v1"
)

const (
	// DeviceAttributePrefix is the prefix used for cni-dra-driver device attributes.
	DeviceAttributePrefix = "cni.dra.networking.x-k8s.io/"

	// InterfaceName is used to set the interface name attribute of a network device.
	InterfaceName resourcev1.QualifiedName = DeviceAttributePrefix + "interfaceName"
	// MacAddress is used to set the MAC address attribute of a network device.
	MacAddress resourcev1.QualifiedName = DeviceAttributePrefix + "macAddress"
	// MaxVirtualDevices is used to set the maximum number of virtual devices in a device capacity.
	MaxVirtualDevices resourcev1.QualifiedName = DeviceAttributePrefix + "maxVirtualDevices"
)
