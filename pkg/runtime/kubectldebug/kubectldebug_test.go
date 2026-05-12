// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package kubectldebug

import (
	"context"
	"testing"

	pkgruntime "github.com/Azure/kubectl-aks/pkg/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestRunCommand_MissingClientset(t *testing.T) {
	r := &Runtime{}
	_, err := r.RunCommand(context.Background(), &pkgruntime.RunOptions{
		NodeName: "node1",
		Command:  "echo hello",
		Timeout:  30,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "kubernetes clientset is required")
}

func TestRunCommand_MissingNodeName(t *testing.T) {
	r := &Runtime{
		Clientset: fake.NewSimpleClientset(),
	}
	_, err := r.RunCommand(context.Background(), &pkgruntime.RunOptions{
		Command: "echo hello",
		Timeout: 30,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "node name is required")
}

func TestBuildDebugPod(t *testing.T) {
	r := &Runtime{
		Clientset: fake.NewSimpleClientset(),
		Image:     "alpine:latest",
		Namespace: "kube-system",
	}

	pod := r.buildDebugPod("my-node", "nsenter -t 1 -m -u -i -n -p -- sh -c 'hostname'")

	assert.Equal(t, "kube-system", pod.Namespace)
	assert.Equal(t, podPrefix, pod.GenerateName)
	assert.Equal(t, "my-node", pod.Spec.NodeName)
	assert.True(t, pod.Spec.HostPID)
	assert.Equal(t, corev1.RestartPolicyNever, pod.Spec.RestartPolicy)
	require.Len(t, pod.Spec.Containers, 1)
	assert.Equal(t, "alpine:latest", pod.Spec.Containers[0].Image)
	assert.True(t, *pod.Spec.Containers[0].SecurityContext.Privileged)
	assert.Equal(t, "kubectl-aks", pod.Labels["app.kubernetes.io/managed-by"])
}

func TestDefaultImageAndNamespace(t *testing.T) {
	r := &Runtime{
		Clientset: fake.NewSimpleClientset(),
	}

	assert.Equal(t, defaultImage, r.image())
	assert.Equal(t, defaultNamespace, r.namespace())
}

func TestCustomImageAndNamespace(t *testing.T) {
	r := &Runtime{
		Clientset: fake.NewSimpleClientset(),
		Image:     "custom:v1",
		Namespace: "monitoring",
	}

	assert.Equal(t, "custom:v1", r.image())
	assert.Equal(t, "monitoring", r.namespace())
}

func TestRunCommand_PodCreatedAndCleaned(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	r := &Runtime{
		Clientset: clientset,
		Namespace: "default",
	}

	// The fake clientset won't actually run the pod, so it stays in Pending.
	// We test that the pod is created and then cleaned up (though the command
	// will timeout). We use a very short timeout.
	ctx := context.Background()
	_, err := r.RunCommand(ctx, &pkgruntime.RunOptions{
		NodeName: "node1",
		Command:  "hostname",
		Timeout:  1,
	})

	// Expect timeout error since fake pods don't transition to Succeeded
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "waiting for debug pod")

	// Verify pod was cleaned up
	pods, listErr := clientset.CoreV1().Pods("default").List(ctx, metav1.ListOptions{})
	require.NoError(t, listErr)
	assert.Empty(t, pods.Items)
}

func TestWaitForPodComplete(t *testing.T) {
	// Create a fake client with a pod already in Succeeded state
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodSucceeded,
		},
	}
	clientset := fake.NewSimpleClientset(pod)
	r := &Runtime{
		Clientset: clientset,
		Namespace: "default",
	}

	err := r.waitForPodComplete(context.Background(), "test-pod", 10)
	assert.NoError(t, err)
}
