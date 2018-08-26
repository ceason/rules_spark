package main

import (
	"testing"
	"encoding/json"
	"io/ioutil"
	"reflect"
	"strings"
)

func TestDriverLauncher(t *testing.T) {
	f, err := ioutil.ReadFile("testdata/driver-pod.json")
	if err != nil {
		t.Error(err)
	}
	podInfo := &PodInfo{}
	err = json.Unmarshal(f, podInfo)
	if err != nil {
		t.Error(err)
	}
	l := driverLauncher{
		podInfo:   podInfo,
		mainClass: "com.example.Main",
		launcherCommon: launcherCommon{
			javaBinary: "/usr/bin/java",
			classpath:  "asdf.jar:abcdefg.jar",
			env: environmentVars{
				"ASDF=ONETWOTHREE",
			},
		},
		appArgs: []string{
			"--args=for", "-my", "app",
		},
	}

	actual, err := l.getCommand()
	if err != nil {
		t.Error(err)
	}
	expected := []string{
		"/usr/bin/java",
		"-Xms1g",
		"-Xmx1g",
		"-Dspark.submit.deployMode=cluster",
		"-Dspark.driver.blockManager.port=7079",
		"-Dspark.driver.port=7078",
		"-Dspark.master=k8s://kubernetes.default.svc",
		"-Dspark.kubernetes.namespace=default",
		"-Dspark.kubernetes.driver.pod.name=kinesistomssql-74d7dc47b5-8cqvx",
		"-Dspark.kubernetes.executor.podNamePrefix=kinesistomssql-74d7dc47b5-8cqvx",
		"-Dspark.driver.host=172.17.0.9",
		"-Dspark.driver.bindAddress=172.17.0.9",
		"-Dspark.app.id=kinesistomssql-74d7dc47b5-8cqvx-1136214245",
		"-Dspark.app.name=kinesistomssql-74d7dc47b5-8cqvx",
		"-Dspark.kubernetes.container.image=registry.kube-system.svc.cluster.local:80/kinesistomssql@sha256:8f687ce9204a646e7053a5e58fbb90bdac2a5457620002cfd744f03ab9d48a48",
		"-Dspark.executor.instances=1",
		"-Dspark.kubernetes.executor.limit.cores=1",
		"-Dspark.kubernetes.executor.label.app=kinesistomssql",
		"-Dspark.executorEnv.ASDF=ONETWOTHREE",
		"-Dspark.kubernetes.executor.secrets.kinesistomssql-aws=/var/run/secrets/aws",
		"-cp",
		"asdf.jar:abcdefg.jar",
		"com.example.Main",
		"--args=for",
		"-my",
		"app",
	}
	if ! reflect.DeepEqual(actual, expected) {
		t.Errorf("Did not match the expected args; actual: \n\"%s\",", strings.Join(actual, "\",\n\""))
	}
}

func TestExecutorLauncher(t *testing.T) {

	l := executorLauncher{
		launcherCommon: launcherCommon{
			javaBinary: "/usr/bin/java",
			classpath:  "asdf.jar:abcdefg.jar",
			env: environmentVars{
				"SPARK_EXECUTOR_MEMORY=1g",
				"SPARK_DRIVER_URL=spark://CoarseGrainedScheduler@172.17.0.8:7078",
				"SPARK_EXECUTOR_ID=4",
				"SPARK_EXECUTOR_CORES=1",
				"SPARK_APPLICATION_ID=spark-application-1526560579316",
				"SPARK_EXECUTOR_POD_IP=172.17.0.10",
			},
		},
	}

	actual, err := l.getCommand()
	if err != nil {
		t.Error(err)
	}
	expected := []string{
		"/usr/bin/java",
		"-Xms1g",
		"-Xmx1g",
		"-cp", "asdf.jar:abcdefg.jar",
		"org.apache.spark.executor.CoarseGrainedExecutorBackend",
		"--driver-url", "spark://CoarseGrainedScheduler@172.17.0.8:7078",
		"--executor-id", "4",
		"--cores", "1",
		"--app-id", "spark-application-1526560579316",
		"--hostname", "172.17.0.10",
	}
	if ! reflect.DeepEqual(actual, expected) {
		t.Errorf("Did not match the expected args; actual: \n\"%s\",", strings.Join(actual, "\",\n\""))
	}
}
