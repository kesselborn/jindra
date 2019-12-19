package k8spodstatus

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	terminated = "Terminated"
	running    = "Running"
	waiting    = "Waiting"
	completed  = "Completed"
	failed     = "Failed"
)

type state struct {
	Terminated *struct {
		ExitCode int
	}
	Running interface{}
	Waiting interface{}
}

type status struct {
	Name  string
	State state
}

type pod struct {
	MetaData struct {
		Name string
	}
	Status struct {
		// pod.status.containerStatuses.state
		ContainerStatuses     []status
		InitContainerStatuses []status
	}
}

// PodInfoStatus represents the status of the current container
type PodInfoStatus struct {
	State   string
	Success *bool `json:",omitempty"`
}

// PodInfo represents the status of a pods containers
type PodInfo struct {
	Name           string
	Containers     map[string]PodInfoStatus
	InitContainers map[string]PodInfoStatus
}

func (pi PodInfo) ContainersState() string {
	containerNames := []string{}
	for name, _ := range pi.Containers {
		containerNames = append(containerNames, name)
	}

	return State(pi.Containers, containerNames...)
}

func (pi PodInfo) InitContainersState() string {
	containerNames := []string{}
	for name, _ := range pi.InitContainers {
		containerNames = append(containerNames, name)
	}

	return State(pi.InitContainers, containerNames...)
}

// State returns an aggregated State of the whole Pod
func State(containers map[string]PodInfoStatus, containerNames ...string) string {
	states := map[string]bool{}

	fmt.Printf("containers: %#v\n", containerNames)
	if len(containerNames) == 1 && containerNames[0] == "" {
		return completed
	}

	for _, c := range containerNames {
		switch {
		case containers[c].State == running:
			states[running] = true
		case containers[c].State == completed:
			states[completed] = true
		case containers[c].State == waiting:
			states[waiting] = true
		case containers[c].State == terminated && *containers[c].Success:
			states[completed] = true
		case containers[c].State == terminated && !*containers[c].Success:
			states[failed] = true
		}
	}

	switch {
	case states[failed]:
		return failed
	case states[waiting] && !states[running]:
		return waiting
	case states[completed] && !states[running]:
		return completed
	case states[running]:
		return running
	}

	return "Unknown"
}

func state2PodInfoStatus(state state) PodInfoStatus {
	var podInfoStatus PodInfoStatus

	switch {
	case state.Running != nil:
		podInfoStatus.State = running
	case state.Terminated != nil:
		podInfoStatus.State = terminated
		podInfoStatus.Success = new(bool)
		*podInfoStatus.Success = state.Terminated.ExitCode == 0
	case state.Waiting != nil:
		podInfoStatus.State = waiting
	default:
		podInfoStatus.State = "Unknown"
	}

	return podInfoStatus
}

// NewPodInfo returns a PodInfo for ns/pod
func NewPodInfo(ns string, pod string) (PodInfo, error) {
	jsonString, err := PodJson(ns, pod)
	if err != nil {
		return PodInfo{}, fmt.Errorf("error retrieving pod json: %s", err)
	}

	return NewPodInfoFromJSON(jsonString), nil
}

func NewPodInfoFromJSON(jsonString string) PodInfo {
	var pod pod
	dec := json.NewDecoder(strings.NewReader(jsonString))
	dec.Decode(&pod)

	return podInfoFromPod(pod)
}

func podInfoFromPod(pod pod) PodInfo {
	podInfo := PodInfo{
		Name:           pod.MetaData.Name,
		Containers:     map[string]PodInfoStatus{},
		InitContainers: map[string]PodInfoStatus{},
	}

	for _, c := range pod.Status.ContainerStatuses {
		podInfo.Containers[c.Name] = state2PodInfoStatus(c.State)
	}

	for _, c := range pod.Status.InitContainerStatuses {
		podInfo.InitContainers[c.Name] = state2PodInfoStatus(c.State)
	}

	return podInfo
}
