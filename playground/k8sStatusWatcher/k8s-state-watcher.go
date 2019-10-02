package k8sStatusWatcher

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	Terminated = "Terminated"
	Running    = "Running"
	Waiting    = "Waiting"
	Completed  = "Completed"
	Failed     = "Failed"
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

type PodInfoStatus struct {
	State   string
	Success *bool `json:",omitempty"`
}

type PodInfo struct {
	Name           string
	Containers     map[string]PodInfoStatus
	InitContainers map[string]PodInfoStatus
}

func (pi PodInfo) State(containers ...string) string {
	states := map[string]bool{}

	fmt.Printf("containers: %#v\n", containers)
	if len(containers) == 1 && containers[0] == "" {
		return Completed
	}

	for _, c := range containers {
		switch {
		case pi.Containers[c].State == Running:
			states[Running] = true
		case pi.Containers[c].State == Completed:
			states[Completed] = true
		case pi.Containers[c].State == Waiting:
			states[Waiting] = true
		case pi.Containers[c].State == Terminated && *pi.Containers[c].Success:
			states[Completed] = true
		case pi.Containers[c].State == Terminated && !*pi.Containers[c].Success:
			states[Failed] = true
		}
	}

	switch {
	case states[Failed]:
		return Failed
	case states[Completed] && !states[Running]:
		return Completed
	case states[Waiting] && !states[Running]:
		return Waiting
	case states[Running]:
		return Running
	}

	return "Unknown"
}

func state2PodInfoStatus(state state) PodInfoStatus {
	var podInfoStatus PodInfoStatus

	switch {
	case state.Running != nil:
		podInfoStatus.State = Running
	case state.Terminated != nil:
		podInfoStatus.State = Terminated
		podInfoStatus.Success = new(bool)
		*podInfoStatus.Success = state.Terminated.ExitCode == 0
	case state.Waiting != nil:
		podInfoStatus.State = Waiting
	default:
		podInfoStatus.State = "Unknown"
	}

	return podInfoStatus
}

func NewPodInfoFromJson(jsonString string) PodInfo {
	var pod pod
	dec := json.NewDecoder(strings.NewReader(jsonString))
	dec.Decode(&pod)

	return NewPodInfo(pod)
}

func NewPodInfo(pod pod) PodInfo {
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
