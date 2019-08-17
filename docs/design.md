# Design Documentation

This document is the design documentation for the Ray Operator.

## API Design

API of the CRD is designed to be extensible:

```go
// RaySpec defines the desired state of Ray
type RaySpec struct {
	Head   *ReplicaSpec `json:"head,omitempty"`
	Worker ReplicaSpec  `json:"worker"`
}

// ReplicaSpec is the replica specification for Head and Worker.
type ReplicaSpec struct {
	Replicas *int32 `json:"replicas,omitempty"`

	// Describes the pod that will be created for this replica.
	Template *corev1.PodTemplateSpec `json:"template,omitempty"`
}
```

## Implementation

The Ray owns two resources: Deployment and Service. When the users submits a Ray CR, then the operator will creates two deployments `{ray.name}-head` and `{ray.name}-worker` and one service `{ray.name}-head` in the same namespace.

The environment `RAY_NODE_IP` will be injected into all pods of the two deployments. `RAY_HEAD_SERVICE` will be injected into all pods of the `{ray.name}-worker` deployment.

The service `{ray.name}-head` automatically gets all ports defined in `{ray.name}-head` deployment and expose them.

## Example

### Simple Use Case

If the users just want to have a try on the Ray, we provide the simplest config yaml file as described here:

```yaml
apiVersion: ray.kubeflow.org/v1
kind: Ray
metadata:
  name: sample-cluster
spec:
  worker:
    replicas: 3
```

### Complicated Use Case

If the users want to define the specification for Head and Worker, we provide the full example yaml file as described here:

```yaml
apiVersion: ray.kubeflow.org/v1
kind: Ray
metadata:
  name: sample-cluster
spec:
  head:
    replicas: 1
    template:
      spec:
        containers:
          - name: ray-head
            image: rayproject/examples
            command: [ "/bin/bash", "-c", "--" ]
            args: ["ray start --head --redis-port=6379 --redis-shard-ports=6380,6381 --object-manager-port=12345 --node-manager-port=12346 --node-ip-address=$RAY_NODE_IP --block"]
            ports:
              - containerPort: 6379
              - containerPort: 6380
              - containerPort: 6381
              - containerPort: 12345
              - containerPort: 12346
            # The environment variables `RAY_NODE_IP` will be injected automatically by the operator.
            # env:
            #   - name: RAY_NODE_IP
            #     valueFrom:
            #       fieldRef:
            #         fieldPath: status.podIP
            resources:
              requests:
                cpu: 500m
              limit:
                cpu: 1
  worker:
    replicas: 3
    template:
      spec:
        containers:
          - name: ray-worker
            image: rayproject/examples
            command: ["/bin/bash", "-c", "--"]
            args: ["ray start --node-ip-address=$MY_POD_IP --redis-address=$(python -c 'import socket;import sys;import os; sys.stdout.write(socket.gethostbyname(os.environ[\"RAY_HEAD_SERVICE\"]));sys.stdout.flush()'):6379 --object-manager-port=12345 --node-manager-port=12346 --block"]
            ports:
              - containerPort: 12345
              - containerPort: 12346
            # The environment variables `RAY_HEAD_SERVICE` and `RAY_NODE_IP`
            # will be injected automatically by the operator.
            # env:
            #   - name: RAY_HEAD_SERVICE
            #     value: <ray-head-service-name>
            #   - name: RAY_NODE_IP
            #     valueFrom:
            #       fieldRef:
            #         fieldPath: status.podIP
            resources:
              requests:
                cpu: 500m
              limit:
                cpu: 1
```