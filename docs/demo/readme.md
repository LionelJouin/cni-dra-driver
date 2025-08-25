# Demo

This demo shows a simple macvlan device class using the CNI DRA Driver, create a DeviceClass, and use it in a simple workload.

Build and push the image (the image must also be set in `deployments/cni-dra-driver.yaml`):
```sh
make push-image REGISTRY=localhost:5000/kubernetes-sigs VERSION=latest
```

Create a Kind cluster with 2 worker nodes and `DynamicResourceAllocation`, `DRAResourceClaimDeviceStatus` and `DRAConsumableCapacity` feature gates enabled:
```sh
kind create cluster --config docs/demo/kind-config.yaml
```

Deploy the CNI-DRA-Driver:
```sh
kubectl apply -f deployments/cni-dra-driver.yaml
kubectl set image daemonset/cni-dra-driver cni-dra-driver=localhost:5000/kubernetes-sigs/cni-dra-driver:latest
```

Deploy a generic device class for network interfaces:
```sh
kubectl apply -f docs/demo/deviceclasses.yaml
```

Deploy a deployment with 3 replicas and a resourceclaimtemplate requesting a macvlan based on eth0:
```sh
kubectl apply -f docs/demo/deployment.yaml
```

## Example

Check the available DeviceClasses:
```sh
$ kubectl get deviceclass
NAME           AGE
network-interface   10m
```

Inspect the details of the DeviceClass:
```yaml
$ kubectl get deviceclass network-interface -o yaml
apiVersion: resource.k8s.io/v1
kind: DeviceClass
metadata:
  creationTimestamp: "2025-09-15T16:21:29Z"
  generation: 1
  name: network-interface
  resourceVersion: "684"
  uid: a671855f-a57f-4ae5-829c-f2b32ea12b44
spec:
  selectors:
  - cel:
      expression: device.driver == "cni.dra.networking.x-k8s.io"
```

List the ResourceSlice objects, which represent the available network resources on each node:
```sh
$ kubectl get resourceslices
NAME                                                   NODE                 DRIVER                        POOL                 AGE
kind-control-plane-cni.dra.networking.x-k8s.io-4zln8   kind-control-plane   cni.dra.networking.x-k8s.io   kind-control-plane   11m
kind-worker-cni.dra.networking.x-k8s.io-jxc5l          kind-worker          cni.dra.networking.x-k8s.io   kind-worker          11m
kind-worker2-cni.dra.networking.x-k8s.io-whlg5         kind-worker2         cni.dra.networking.x-k8s.io   kind-worker2         11m
```

Inspect a specific ResourceSlice to see the available network devices and their attributes/capacity:
```yaml
$ kubectl get resourceslices kind-worker-cni.dra.networking.x-k8s.io-jxc5l -o yaml
apiVersion: resource.k8s.io/v1
kind: ResourceSlice
metadata:
  creationTimestamp: "2025-09-15T16:21:10Z"
  generateName: kind-worker-cni.dra.networking.x-k8s.io-
  generation: 71
  name: kind-worker-cni.dra.networking.x-k8s.io-jxc5l
  ownerReferences:
  - apiVersion: v1
    controller: true
    kind: Node
    name: kind-worker
    uid: 43b1a1ba-4765-4c73-9465-c0450b563b4a
  resourceVersion: "2111"
  uid: f5b1425d-f5fb-4e5e-8fde-d3438282e537
spec:
  devices:
  - allowMultipleAllocations: true
    attributes:
      cni.dra.networking.x-k8s.io/interfaceName:
        string: eth0
    capacity:
      cni.dra.networking.x-k8s.io/maxVirtualDevices:
        requestPolicy:
          default: "1"
          validValues:
          - "1"
        value: "65535"
    name: eth0
  driver: cni.dra.networking.x-k8s.io
  nodeName: kind-worker
  pool:
    generation: 1
    name: kind-worker
    resourceSliceCount: 1
```

Check the CNI-DRA driver pods and the demo application pods are running:
```sh
$ kubectl get pods -o wide
NAME                               READY   STATUS    RESTARTS   AGE   IP           NODE                 NOMINATED NODE   READINESS GATES
cni-dra-driver-d5rl5               1/1     Running   0          12m   172.18.0.4   kind-worker2         <none>           <none>
cni-dra-driver-ldrt5               1/1     Running   0          12m   172.18.0.2   kind-control-plane   <none>           <none>
cni-dra-driver-npzn4               1/1     Running   0          12m   172.18.0.3   kind-worker          <none>           <none>
demo-deployment-78bd85684c-cp62r   1/1     Running   0          11m   10.244.1.2   kind-worker          <none>           <none>
demo-deployment-78bd85684c-xqtnb   1/1     Running   0          11m   10.244.2.3   kind-worker2         <none>           <none>
demo-deployment-78bd85684c-zp9wq   1/1     Running   0          11m   10.244.2.2   kind-worker2         <none>           <none>
```

List the ResourceClaims representing the allocated network resources for each pod:
```sh
$ kubectl get resourceclaim
NAME                                                  STATE                AGE
demo-deployment-78bd85684c-cp62r-macvlan-eth0-2zrrn   allocated,reserved   12m
demo-deployment-78bd85684c-xqtnb-macvlan-eth0-qdxcp   allocated,reserved   12m
demo-deployment-78bd85684c-zp9wq-macvlan-eth0-lrbjv   allocated,reserved   12m
```

Inspect a specific ResourceClaim to see the allocated device and its configuration:
```yaml
$ kubectl get resourceclaim demo-deployment-78bd85684c-cp62r-macvlan-eth0-2zrrn -o yaml
apiVersion: resource.k8s.io/v1
kind: ResourceClaim
metadata:
  annotations:
    resource.kubernetes.io/pod-claim-name: macvlan-eth0
  creationTimestamp: "2025-09-15T16:22:28Z"
  finalizers:
  - resource.kubernetes.io/delete-protection
  generateName: demo-deployment-78bd85684c-cp62r-macvlan-eth0-
  name: demo-deployment-78bd85684c-cp62r-macvlan-eth0-2zrrn
  namespace: default
  ownerReferences:
  - apiVersion: v1
    blockOwnerDeletion: true
    controller: true
    kind: Pod
    name: demo-deployment-78bd85684c-cp62r
    uid: f9715e15-4159-478c-95bf-d3529ed45322
  resourceVersion: "916"
  uid: 928a22a6-aedd-4ff3-a2c1-c4dca29ad8d8
spec:
  devices:
    config:
    - opaque:
        driver: cni.dra.networking.x-k8s.io
        parameters:
          apiVersion: cni.networking.x-k8s.io/v1alpha1
          config:
            cniVersion: 1.0.0
            name: macvlan-eth0
            plugins:
            - ipam:
                ranges:
                - - subnet: 10.10.1.0/24
                type: host-local
              master: eth0
              mode: bridge
              type: macvlan
          ifName: net1
          kind: CNI
      requests:
      - macvlan-eth0
    requests:
    - exactly:
        allocationMode: ExactCount
        count: 1
        deviceClassName: network-interface
        selectors:
        - cel:
            expression: device.attributes["cni.dra.networking.x-k8s.io"].interfaceName
              == "eth0"
      name: macvlan-eth0
status:
  allocation:
    devices:
      config:
      - opaque:
          driver: cni.dra.networking.x-k8s.io
          parameters:
            apiVersion: cni.networking.x-k8s.io/v1alpha1
            config:
              cniVersion: 1.0.0
              name: macvlan-eth0
              plugins:
              - ipam:
                  ranges:
                  - - subnet: 10.10.1.0/24
                  type: host-local
                master: eth0
                mode: bridge
                type: macvlan
            ifName: net1
            kind: CNI
        requests:
        - macvlan-eth0
        source: FromClaim
      results:
      - consumedCapacity:
          cni.dra.networking.x-k8s.io/maxVirtualDevices: "1"
        device: eth0
        driver: cni.dra.networking.x-k8s.io
        pool: kind-worker
        request: macvlan-eth0
        shareID: 8e7acdf9-0290-4ecd-a801-a654b021d2b7
    nodeSelector:
      nodeSelectorTerms:
      - matchFields:
        - key: metadata.name
          operator: In
          values:
          - kind-worker
  devices:
  - conditions: null
    data:
      cniVersion: 1.0.0
      interfaces:
      - mac: 5a:9f:d8:84:fb:51
        name: net1
        sandbox: /var/run/netns/cni-4cf00f23-a64c-f12a-8318-816b30fac0fc
      ips:
      - address: 10.10.1.2/24
        gateway: 10.10.1.1
        interface: 0
    device: eth0
    driver: cni.dra.networking.x-k8s.io
    networkData:
      hardwareAddress: 5a:9f:d8:84:fb:51
      interfaceName: net1
      ips:
      - 10.10.1.2/24
    pool: kind-worker
    shareID: 8e7acdf9-0290-4ecd-a801-a654b021d2b7
  reservedFor:
  - name: demo-deployment-78bd85684c-cp62r
    resource: pods
    uid: f9715e15-4159-478c-95bf-d3529ed45322
```

Check the network interfaces inside the pod to confirm the interface (net1) is created with corresponding IP and Mac Address:
```sh
$ kubectl exec -it demo-deployment-78bd85684c-cp62r -- ip a
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
    inet6 ::1/128 scope host 
       valid_lft forever preferred_lft forever
2: eth0@net1: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP qlen 1000
    link/ether 16:25:c2:d7:02:01 brd ff:ff:ff:ff:ff:ff
    inet 10.244.1.2/24 brd 10.244.1.255 scope global eth0
       valid_lft forever preferred_lft forever
    inet6 fd00:10:244:1::2/64 scope global 
       valid_lft forever preferred_lft forever
    inet6 fe80::1425:c2ff:fed7:201/64 scope link 
       valid_lft forever preferred_lft forever
3: net1@eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP qlen 1000
    link/ether 5a:9f:d8:84:fb:51 brd ff:ff:ff:ff:ff:ff
    inet 10.10.1.2/24 brd 10.10.1.255 scope global net1
       valid_lft forever preferred_lft forever
    inet6 fe80::589f:d8ff:fe84:fb51/64 scope link 
       valid_lft forever preferred_lft forever
```
