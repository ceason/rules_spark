package main

import "time"

type list []string

type environmentVars []string

type sparkLauncher interface {
	exec() error
}

type launcherCommon struct {
	javaBinary string
	classpath  string
	env        environmentVars
	jvmFlags   []string
}

type driverLauncher struct {
	launcherCommon
	driverContainer          string
	podInfo                  *PodInfo
	mainClass                string
	appArgs                  []string
	now                      time.Time
	nopassthroughLabels      []string
	nopassthroughAnnotations []string
	nopassthroughEnv         []string
	nopassthroughJvmFlags    []string
	nopassthroughVolumes     []string
}

type executorLauncher struct {
	launcherCommon
}

type initLauncher struct {
}

type Volume struct {
	Name string `json:"name"`
	Secret struct {
		DefaultMode int    `json:"defaultMode"`
		Optional    bool   `json:"optional"`
		SecretName  string `json:"secretName"`
	} `json:"secret"`
}

type Container struct {
	Env []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"env"`
	EnvFrom []struct {
		ConfigMapRef struct {
			Name     string `json:"name"`
			Optional bool   `json:"optional"`
		} `json:"configMapRef,omitempty"`
	} `json:"envFrom"`
	Image           string `json:"image"`
	ImagePullPolicy string `json:"imagePullPolicy"`
	Name            string `json:"name"`
	Ports []struct {
		ContainerPort int    `json:"containerPort"`
		Name          string `json:"name"`
		Protocol      string `json:"protocol"`
	} `json:"ports"`
	Resources struct {
		Limits struct {
			Memory string `json:"memory"`
			CPU    string `json:"cpu"`
		} `json:"limits"`
		Requests struct {
			CPU    string `json:"cpu"`
			Memory string `json:"memory"`
		} `json:"requests"`
	} `json:"resources"`
	TerminationMessagePath   string `json:"terminationMessagePath"`
	TerminationMessagePolicy string `json:"terminationMessagePolicy"`
	VolumeMounts []struct {
		MountPath string `json:"mountPath"`
		Name      string `json:"name"`
		ReadOnly  bool   `json:"readOnly,omitempty"`
	} `json:"volumeMounts"`
}

type PodInfo struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Metadata struct {
		CreationTimestamp time.Time         `json:"creationTimestamp"`
		GenerateName      string            `json:"generateName"`
		Labels            map[string]string `json:"labels"`
		Annotations       map[string]string
		Name              string            `json:"name"`
		Namespace         string            `json:"namespace"`
		OwnerReferences []struct {
			APIVersion         string `json:"apiVersion"`
			BlockOwnerDeletion bool   `json:"blockOwnerDeletion"`
			Controller         bool   `json:"controller"`
			Kind               string `json:"kind"`
			Name               string `json:"name"`
			UID                string `json:"uid"`
		} `json:"ownerReferences"`
		ResourceVersion string `json:"resourceVersion"`
		SelfLink        string `json:"selfLink"`
		UID             string `json:"uid"`
	} `json:"metadata"`
	Spec struct {
		Containers    []Container `json:"containers"`
		DNSPolicy     string      `json:"dnsPolicy"`
		NodeName      string      `json:"nodeName"`
		RestartPolicy string      `json:"restartPolicy"`
		SchedulerName string      `json:"schedulerName"`
		SecurityContext struct {
		} `json:"securityContext"`
		ServiceAccount                string `json:"serviceAccount"`
		ServiceAccountName            string `json:"serviceAccountName"`
		TerminationGracePeriodSeconds int    `json:"terminationGracePeriodSeconds"`
		Tolerations []struct {
			Effect            string `json:"effect"`
			Key               string `json:"key"`
			Operator          string `json:"operator"`
			TolerationSeconds int    `json:"tolerationSeconds"`
		} `json:"tolerations"`
		Volumes []Volume `json:"volumes"`
	} `json:"spec"`
	Status struct {
		Conditions []struct {
			LastProbeTime      interface{} `json:"lastProbeTime"`
			LastTransitionTime time.Time   `json:"lastTransitionTime"`
			Status             string      `json:"status"`
			Type               string      `json:"type"`
			Message            string      `json:"message,omitempty"`
			Reason             string      `json:"reason,omitempty"`
		} `json:"conditions"`
		ContainerStatuses []struct {
			ContainerID string `json:"containerID"`
			Image       string `json:"image"`
			ImageID     string `json:"imageID"`
			LastState struct {
				Terminated struct {
					ContainerID string    `json:"containerID"`
					ExitCode    int       `json:"exitCode"`
					FinishedAt  time.Time `json:"finishedAt"`
					Message     string    `json:"message"`
					Reason      string    `json:"reason"`
					StartedAt   time.Time `json:"startedAt"`
				} `json:"terminated"`
			} `json:"lastState"`
			Name         string `json:"name"`
			Ready        bool   `json:"ready"`
			RestartCount int    `json:"restartCount"`
			State struct {
				Waiting struct {
					Message string `json:"message"`
					Reason  string `json:"reason"`
				} `json:"waiting"`
			} `json:"state"`
		} `json:"containerStatuses"`
		HostIP    string    `json:"hostIP"`
		Phase     string    `json:"phase"`
		PodIP     string    `json:"podIP"`
		QosClass  string    `json:"qosClass"`
		StartTime time.Time `json:"startTime"`
	} `json:"status"`
}
