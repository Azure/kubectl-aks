// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package kubectldebug

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/briandowns/spinner"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	pkgruntime "github.com/Azure/kubectl-aks/pkg/runtime"
)

const (
	defaultImage     = "busybox:latest"
	defaultNamespace = "default"
	podPrefix        = "kubectl-aks-debug-"
)

// Runtime executes commands on AKS nodes by creating a privileged debug pod.
type Runtime struct {
	Clientset kubernetes.Interface
	Config    *rest.Config
	Image     string
	Namespace string
}

func (r *Runtime) image() string {
	if r.Image != "" {
		return r.Image
	}
	return defaultImage
}

func (r *Runtime) namespace() string {
	if r.Namespace != "" {
		return r.Namespace
	}
	return defaultNamespace
}

func (r *Runtime) RunCommand(ctx context.Context, opts *pkgruntime.RunOptions) (*pkgruntime.RunResult, error) {
	if r.Clientset == nil {
		return nil, fmt.Errorf("kubernetes clientset is required for kube-api runtime")
	}
	if opts.NodeName == "" {
		return nil, fmt.Errorf("node name is required for kube-api runtime")
	}

	// Use nsenter to get host-level access, matching VMSS RunCommand behavior
	nsenterCmd := fmt.Sprintf("nsenter -t 1 -m -u -i -n -p -- sh -c '%s'", opts.Command)

	pod := r.buildDebugPod(opts.NodeName, nsenterCmd)

	s := spinner.New(spinner.CharSets[9], 200*time.Millisecond)
	s.Suffix = " Creating debug pod..."
	s.Start()

	// Create the debug pod
	createdPod, err := r.Clientset.CoreV1().Pods(r.namespace()).Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		s.Stop()
		return nil, fmt.Errorf("creating debug pod: %w", err)
	}

	// Always clean up the pod
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = r.Clientset.CoreV1().Pods(r.namespace()).Delete(cleanupCtx, createdPod.Name, metav1.DeleteOptions{})
	}()

	s.Suffix = " Running..."

	// Wait for pod to complete
	if err := r.waitForPodComplete(ctx, createdPod.Name, opts.Timeout); err != nil {
		s.Stop()
		return nil, fmt.Errorf("waiting for debug pod to complete: %w", err)
	}

	s.Stop()

	// Get logs (stdout/stderr combined in container logs)
	stdout, err := r.getPodLogs(ctx, createdPod.Name)
	if err != nil {
		return nil, fmt.Errorf("getting debug pod logs: %w", err)
	}

	return &pkgruntime.RunResult{
		Stdout: stdout,
		Stderr: "",
	}, nil
}

func (r *Runtime) buildDebugPod(nodeName, command string) *corev1.Pod {
	privileged := true
	hostPID := true
	zero := int64(0)

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: podPrefix,
			Namespace:    r.namespace(),
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "kubectl-aks",
			},
		},
		Spec: corev1.PodSpec{
			NodeName:      nodeName,
			HostPID:       hostPID,
			RestartPolicy: corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				{
					Name:    "debug",
					Image:   r.image(),
					Command: []string{"sh", "-c", command},
					SecurityContext: &corev1.SecurityContext{
						Privileged: &privileged,
					},
				},
			},
			TerminationGracePeriodSeconds: &zero,
		},
	}
}

func (r *Runtime) waitForPodComplete(ctx context.Context, podName string, timeoutSeconds int) error {
	timeout := time.Duration(timeoutSeconds) * time.Second

	return wait.PollImmediate(2*time.Second, timeout, func() (bool, error) {
		pod, err := r.Clientset.CoreV1().Pods(r.namespace()).Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		switch pod.Status.Phase {
		case corev1.PodSucceeded, corev1.PodFailed:
			return true, nil
		default:
			return false, nil
		}
	})
}

func (r *Runtime) getPodLogs(ctx context.Context, podName string) (string, error) {
	req := r.Clientset.CoreV1().Pods(r.namespace()).GetLogs(podName, &corev1.PodLogOptions{
		Container: "debug",
	})
	stream, err := req.Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("streaming pod logs: %w", err)
	}
	defer stream.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, stream); err != nil {
		return "", fmt.Errorf("reading pod logs: %w", err)
	}

	return buf.String(), nil
}
