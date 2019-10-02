package k8sStatusWatcher

import (
	"fmt"
	"os/exec"
)

func PodJson(namespace, pod string) (string, error) {
	args := []string{"get", "pod", pod, "--output", "json"}

	if namespace != "" {
		args = append(args, "--namespace", namespace)
	}

	out, err := exec.Command("kubectl", args...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Error %s: %s", err, string(out))
	}

	return string(out), nil
}
