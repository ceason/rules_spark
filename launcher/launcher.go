package main

import (
	"fmt"
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



func (l executorLauncher) exec() error {
	args, err := l.getCommand()
	if err != nil {
		return err
	}
	//log.Printf("INFO: Le command to execute! \n\"%s\",", strings.Join(args, "\",\n\""))
	return syscall.Exec(args[0], args, l.env)
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
