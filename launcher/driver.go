package main

import (
	"fmt"
	"log"
	"strings"
	"syscall"
	"errors"
	"encoding/json"
)

func (l driverLauncher) exec() error {
	args, err := l.getCommand()
	if err != nil {
		return err
	}
	// clean up any existing executors (can happen when driver pod restarts)
	client := newInClusterClient()
	req := client.mustNewRequest("GET", fmt.Sprintf(
		"/api/v1/namespaces/%s/pods?labelSelector=spark-role=executor", l.podInfo.Metadata.Namespace), nil)
	podList := &PodInfoList{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(podList)
	if err != nil {
		panic(err)
	}
	for _, pod := range podList.Items {
		// make sure we're only looking at executors we own
		if strings.HasPrefix(pod.Metadata.Name, l.executorPodNamePrefix()) {
			deleteReq := client.mustNewRequest("DELETE", fmt.Sprintf(
				"/api/v1/namespaces/%s/pods/%s?gracePeriodSeconds=0", pod.Metadata.Namespace, pod.Metadata.Name), nil)
			resp, err := client.Do(deleteReq)
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				log.Fatalf("Received unsuccessful response status code '%d' when trying to delete pod '%s'", resp.StatusCode, pod.Metadata.Name)
			}
			log.Printf("Cleaned up old executor '%s'", pod.Metadata.Name)
		}
	}

	//log.Printf("INFO: Le command to execute! \n\"%s\",", strings.Join(args, "\",\n\""))
	return syscall.Exec(args[0], args, l.env)
}

func (l driverLauncher) executorPodNamePrefix() string {
	return l.podInfo.Metadata.Name
}

func (l driverLauncher) getCommand() ([]string, error) {
	var driverContainer *Container
	for _, c := range l.podInfo.Spec.Containers {
		if c.Name == DefaultSparkDriverContainerName {
			driverContainer = &c
			break
		}
	}
	if driverContainer == nil {
		return nil, errors.New(fmt.Sprintf("Could not find a container named '%s'", DefaultSparkDriverContainerName))
	}

	args := []string{
		l.javaBinary,
		fmt.Sprintf("-Xms%s", l.env.getOrElse("SPARK_DRIVER_MEMORY", "1g")),
		fmt.Sprintf("-Xmx%s", l.env.getOrElse("SPARK_DRIVER_MEMORY", "1g")),
		"-Dspark.submit.deployMode=cluster",
		"-Dspark.driver.blockManager.port=7079",
		"-Dspark.driver.port=7078",
		"-Dspark.master=k8s://kubernetes.default.svc",
		fmt.Sprintf("-Dspark.kubernetes.namespace=%s", l.podInfo.Metadata.Namespace),
		fmt.Sprintf("-Dspark.kubernetes.driver.pod.name=%s", l.podInfo.Metadata.Name),
		fmt.Sprintf("-Dspark.kubernetes.executor.podNamePrefix=%s", l.executorPodNamePrefix()),
		fmt.Sprintf("-Dspark.driver.host=%s", l.podInfo.Status.PodIP),
		fmt.Sprintf("-Dspark.driver.bindAddress=%s", l.podInfo.Status.PodIP),
		fmt.Sprintf("-Dspark.app.name=%s", l.podInfo.Metadata.Name),
		fmt.Sprintf("-Dspark.kubernetes.container.image=%s", driverContainer.Image),
		fmt.Sprintf("-Dspark.executor.instances=%s", l.env.getOrElse("SPARK_EXECUTOR_INSTANCES", "1")),
		fmt.Sprintf("-Dspark.kubernetes.executor.limit.cores=%s", l.env.getOrElse("SPARK_EXECUTOR_CORES", "1")),
	}

	// set driver jvm flags
	for _, value := range l.jvmFlags {
		args = append(args, value)
	}

	// pass jvm flags to executors
	executorExtraJavaOpts := []string{}
	for _, value := range l.jvmFlags {
		excludeFlag := false
		if strings.Contains(value, "=") {
			// if it has an equals we compare this way..
			flag := strings.Split(value, "=")[0]
			if list(l.nopassthroughJvmFlags).contains(flag) {
				excludeFlag = true
			}
		} else {
			// if there's no equals we compare the flag's prefix to the exclusion list (eg. "-Xms500m/-Xmx1g" don't have equals in them..)
			for _, excluded := range l.nopassthroughJvmFlags {
				if strings.HasPrefix(value, excluded) {
					excludeFlag = true
				}
			}
		}
		if ! excludeFlag {
			executorExtraJavaOpts = append(executorExtraJavaOpts, value)
		}
	}
	if len(executorExtraJavaOpts) > 0 {
		args = append(args, fmt.Sprintf("-Dspark.executor.extraJavaOptions=%s", strings.Join(executorExtraJavaOpts, " ")))
	}

	// pass driver labels to executors
	for key, value := range l.podInfo.Metadata.Labels {
		if ! list(l.nopassthroughLabels).contains(key) {
			args = append(args, fmt.Sprintf("-Dspark.kubernetes.executor.label.%s=%s", key, value))
		}
	}
	// pass driver annotations to executors
	for key, value := range l.podInfo.Metadata.Annotations {
		if ! list(l.nopassthroughAnnotations).contains(key) {
			args = append(args, fmt.Sprintf("-Dspark.kubernetes.executor.annotation.%s=%s", key, value))
		}
	}
	// pass driver environment to executors
	for _, value := range l.env {
		if ! list(l.nopassthroughEnv).contains(value) {
			args = append(args, fmt.Sprintf("-Dspark.executorEnv.%s", value))
		}
	}
	// pass each mount through to executors
	for _, mount := range driverContainer.VolumeMounts {
		// skip the serviceaccount mount
		if mount.MountPath == "/var/run/secrets/kubernetes.io/serviceaccount" {
			continue
		}
		// find the volume associated with the mount
		var volume Volume
		for _, vol := range l.podInfo.Spec.Volumes {
			if vol.Name == mount.Name {
				volume = vol
			}
		}
		if volume.Name == "" {
			log.Printf("WARN: Could not find volume '%s'; it won't be mounted on executors")
			continue
		}
		if volume.Secret.SecretName == "" {
			log.Printf("WARN: Volume '%s' is not a Secret and can't be passed through to the Executor", volume.Name)
			continue
		}
		if ! list(l.nopassthroughVolumes).contains(volume.Name) {
			args = append(args, fmt.Sprintf("-Dspark.kubernetes.executor.secrets.%s=%s", volume.Secret.SecretName, mount.MountPath))
		}
	}
	args = append(args, "-cp", l.classpath, l.mainClass)
	args = append(args, l.appArgs...)

	return args, nil
}
