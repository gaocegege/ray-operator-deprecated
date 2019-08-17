# ray-operator

This operator is still in early stage, do not use it in prod.

## Clone

```
git clone https://github.com/gaocegege/ray-operator
mkdir -p $GOPATH/src/github.com/kubeflow/
mv ./ray-operator $GOPATH/src/github.com/kubeflow/
```

## Build

```sh
cd $GOPATH/src/github.com/kubeflow/ray-operator
build
```

## HOWTO

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
            image: rayproject/examples:1
            command: [ "/bin/bash", "-c", "--" ]
            args: ["ray start --head --redis-port=6379 --redis-shard-ports=6380,6381 --object-manager-port=12345 --node-manager-port=12346 --node-ip-address=$RAY_NODE_IP --block"]
            ports:
              - containerPort: 6379
              - containerPort: 6380
              - containerPort: 6381
              - containerPort: 12345
              - containerPort: 12346
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
            image: rayproject/examples:1
            command: ["/bin/bash", "-c", "--"]
            args: ["ray start --node-ip-address=$MY_POD_IP --redis-address=$(python -c 'import socket;import sys;import os; sys.stdout.write(socket.gethostbyname(os.environ[\"RAY_HEAD_SERVICE\"]));sys.stdout.flush()'):6379 --object-manager-port=12345 --node-manager-port=12346 --block"]
            ports:
              - containerPort: 12345
              - containerPort: 12346
            resources:
              requests:
                cpu: 500m
              limit:
                cpu: 1
```

Then you could run `kubectl describe ray`. 

```
$ kubectl describe ray
...
Status:
  Conditions:
    Last Transition Time:  2019-08-17T10:41:54Z
    Last Update Time:      2019-08-17T10:42:28Z
    Status:                Unknown
    Type:                  RayWorkerDeploymentReplicaFailure
    Last Transition Time:  2019-08-17T10:41:54Z
    Last Update Time:      2019-08-17T10:42:28Z
    Status:                Unknown
    Type:                  RayHeadDeploymentReplicaFailure
    Last Transition Time:  2019-08-17T10:42:28Z
    Last Update Time:      2019-08-17T10:42:28Z
    Status:                True
    Type:                  Health
    Last Transition Time:  2019-08-17T10:41:54Z
    Last Update Time:      2019-08-17T10:42:28Z
    Message:               ReplicaSet "sample-cluster-head-8684888b99" has successfully progressed.
    Reason:                NewReplicaSetAvailable
    Status:                True
    Type:                  RayHeadDeploymentProgressing
    Last Transition Time:  2019-08-17T10:41:54Z
    Last Update Time:      2019-08-17T10:42:28Z
    Message:               ReplicaSet "sample-cluster-worker-7447bd5cbc" has successfully progressed.
    Reason:                NewReplicaSetAvailable
    Status:                True
    Type:                  RayWorkerDeploymentProgressing
    Last Transition Time:  2019-08-17T10:41:57Z
    Last Update Time:      2019-08-17T10:42:28Z
    Status:                True
    Type:                  RayHeadDeploymentAvailable
    Last Transition Time:  2019-08-17T10:42:28Z
    Last Update Time:      2019-08-17T10:42:28Z
    Status:                True
    Type:                  RayWorkerDeploymentAvailable
  Head:
    Available Replicas:  1
    Ready Replicas:      1
    Replicas:            1
    Updated Replicas:    1
  Last Reconcile Time:   2019-08-17T10:42:28Z
  Observed Generation:   1
  Start Time:            2019-08-17T10:41:54Z
  Worker:
    Available Replicas:  3
    Ready Replicas:      3
    Replicas:            3
    Updated Replicas:    3
Events:
  Type     Reason                            Age                     From                    Message
  ----     ------                            ----                    ----                    -------
  Normal   SuccessfullyCreate                3m30s                   ray-operator            Successfully create the service sample-cluster-head
  Normal   SuccessfullyCreate                3m30s                   ray-operator            Successfully create the deployment sample-cluster-head
  Normal   SuccessfullyCreate                3m30s                   ray-operator            Successfully create the deployment sample-cluster-worker
```

There is a condition `Health`, which shows if all components are ready in the Ray cluster. Besides this, you can get the status of the Head and Worker in `status.head` and `status.worker`:

```
Status:
  Head:
    Available Replicas:  1
    Ready Replicas:      1
    Replicas:            1
    Updated Replicas:    1
  Worker:
    Available Replicas:  3
    Ready Replicas:      3
    Replicas:            3
    Updated Replicas:    3
```

It shows the the most recently observed status of the Head and Worker. In this case, we get 3 available workers and 1 available head.

## Design

[Design Document](./docs/design.md)

## LoC

```sh
cloc . --exclude-dir=vendor,docs,config,hack --exclude-lang=Markdown,make,Dockerfile
21 text files.
21 unique files.                              
8 files ignored.

github.com/AlDanial/cloc v 1.74  T=0.04 s (324.6 files/s, 31442.0 lines/s)
-------------------------------------------------------------------------------
Language                     files          blank        comment           code
-------------------------------------------------------------------------------
Go                              14            188            196            972
-------------------------------------------------------------------------------
SUM:                            14            188            196            972
-------------------------------------------------------------------------------
```
