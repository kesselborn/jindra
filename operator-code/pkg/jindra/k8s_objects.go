package jindra

import (
	"fmt"
	"path"

	core "k8s.io/api/core/v1"
)

func watcherContainer(stageName, waitFor string, semaphoreMount core.VolumeMount) core.Container {
	return core.Container{
		Name:  "jindra-watcher",
		Image: "alpine",
		Args: []string{"sh", "-c", fmt.Sprintf(`printf "waiting for steps to finish "
containers=$(echo "%s"|sed "s/[,]*%s//g")
while ! wget -qO- ${MY_IP}:8080/pod/${MY_NAME}.%s?containers=${containers}|grep Completed &>/dev/null
do
  printf "."
  sleep 3
done
echo
rm %s
`, waitFor, debugContainerName, stageName, path.Join(semaphoresPrefixPath, "steps-running")),
		},
		Env: []core.EnvVar{
			{Name: "JOB_IP", Value: "${MY_IP}"},
		},
		VolumeMounts: []core.VolumeMount{
			semaphoreMount,
		},
	}
}

func debugContainer(toolsMount, semaphoreMount core.VolumeMount, resourceNames []string) core.Container {
	c := core.Container{
		Name:  debugContainerName,
		Image: "alpine",
		Args: []string{"sh", "-c", `touch /DELETE_ME_TO_STOP_DEBUG_CONTAINER
echo "waiting for /DELETE_ME_TO_STOP_DEBUG_CONTAINER to be deleted "
while test -f /DELETE_ME_TO_STOP_DEBUG_CONTAINER
do
  sleep 1
  printf "."
done`},
		Env: []core.EnvVar{
			{Name: "JOB_IP", Value: "${MY_IP}"},
		},
		VolumeMounts: []core.VolumeMount{
			toolsMount,
			semaphoreMount,
		},
	}

	for _, resource := range resourceNames {
		c.VolumeMounts = append(c.VolumeMounts,
			core.VolumeMount{Name: resourceVolumePrefix + resource, MountPath: path.Join(resourcesPrefixPath, resource)},
		)
	}

	return c
}
