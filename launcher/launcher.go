package main

import (
	"fmt"
	"errors"
	"log"
	"os"
	"strings"
	"syscall"
)


const (
	DefaultSparkDriverContainerName = "spark-kubernetes-driver"
)


func (l list) contains(str string) bool {
	for _, value := range l {
		if str == value {
			return true
		}
	}
	return false
}

func (l initLauncher) exec() error {
	log.Printf("INFO: Ignoring 'init' command and moving on..")
	return nil
}

func (l driverLauncher) exec() error {
	args, err := l.getCommand()
	if err != nil {
		return err
	}
	log.Printf("INFO: Le command to execute! \n\"%s\",", strings.Join(args, "\",\n\""))
	return syscall.Exec(args[0], args, l.env)
}

func (l executorLauncher) exec() error {
	args, err := l.getCommand()
	if err != nil {
		return err
	}
	log.Printf("INFO: Le command to execute! \n\"%s\",", strings.Join(args, "\",\n\""))
	return syscall.Exec(args[0], args, l.env)
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
		fmt.Sprintf("-Dspark.kubernetes.executor.podNamePrefix=%s", l.podInfo.Metadata.Name),
		fmt.Sprintf("-Dspark.driver.host=%s", l.podInfo.Status.PodIP),
		fmt.Sprintf("-Dspark.driver.bindAddress=%s", l.podInfo.Status.PodIP),
		fmt.Sprintf("-Dspark.app.id=%s", l.podInfo.Metadata.Name),
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

func (l executorLauncher) getCommand() ([]string, error) {
	args := []string{
		l.javaBinary,
		fmt.Sprintf("-Xms%s", l.env.getOrElse("SPARK_EXECUTOR_MEMORY", "1g")),
		fmt.Sprintf("-Xmx%s", l.env.getOrElse("SPARK_EXECUTOR_MEMORY", "1g")),
		"-cp", l.classpath,
		"org.apache.spark.executor.CoarseGrainedExecutorBackend",
		"--driver-url", l.env.mustGet("SPARK_DRIVER_URL"),
		"--executor-id", l.env.mustGet("SPARK_EXECUTOR_ID"),
		"--cores", l.env.mustGet("SPARK_EXECUTOR_CORES"),
		"--app-id", l.env.mustGet("SPARK_APPLICATION_ID"),
		"--hostname", l.env.mustGet("SPARK_EXECUTOR_POD_IP"),
	}

	return args, nil
}

func (e environmentVars) getOrElse(name, defaultValue string) string {
	prefix := fmt.Sprintf("%s=", name)
	for _, v := range os.Environ() {
		if strings.HasPrefix(v, prefix) {
			return strings.TrimPrefix(v, prefix)
		}
	}
	return defaultValue
}

func (e environmentVars) mustGet(name string) string {
	prefix := fmt.Sprintf("%s=", name)
	for _, v := range e {
		if strings.HasPrefix(v, prefix) {
			return strings.TrimPrefix(v, prefix)
		}
	}
	panic(fmt.Sprintf("Could not find env var '%s'", name))
}
