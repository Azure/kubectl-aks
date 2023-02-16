#!/bin/bash

REMOTE_PORT=8887
LOCAL_PORT=8080

nodes_str=$(kubectl get nodes -o jsonpath='{.items[*].metadata.name}')
nodes=($nodes_str)

echo "All nodes: ${nodes[*]}"

node=${nodes[0]}

echo "poc, selected first node: $node"

pods_str=$(kubectl get pods -n kube-system --field-selector spec.nodeName=$node -o jsonpath='{.items[*].metadata.name}')
echo "all pods on selected node: $pods_str"

target_pod=""
for pod in $pods_str
do
    if [[ $pod == *"cloud-node-manager-"* ]]; then
        target_pod=$pod
        break
    fi
done

if [[ -z "$target_pod" ]]; then
    echo "Didn't find targetpod"
    exit 1
fi
echo "Found targetpod: $target_pod"

kubectl port-forward $target_pod -n kube-system $LOCAL_PORT:$REMOTE_PORT
