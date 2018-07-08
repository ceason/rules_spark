package main

import (
	"os"
	"time"
	"flag"
	"strings"
	"regexp"
)

type arrayFlags []string

// Split listable args on commas, spaces, newlines
var listArgSeparator = regexp.MustCompile("[\n, ]")

var (
	driverMainClass          = flag.String("driver-main-class", "", "Main class to run when executing as the driver, eg: -main-class=com.example.WordCount")
	driverContainer          = flag.String("driver-container", DefaultSparkDriverContainerName, "Name of driver container, used for discovery to launch executors properly")
	nopassthroughLabels      = flag.String("nopassthrough-labels", "spark-role,pod-template-hash", "Pod labels on the driver which should *not* be passed through to executors")
	nopassthroughAnnotations = flag.String("nopassthrough-annotations", "kubectl.kubernetes.io/last-applied-configuration", "Pod annotations on the driver which should *not* be passed through to executors")
	nopassthroughEnv         = flag.String("nopassthrough-env", "", "Env vars on the driver which should *not* be passed through to executors")
	nopassthroughJvmFlags    = flag.String("nopassthrough-jvm-flags", "-Xms,-Xms", "Env vars on the driver which should *not* be passed through to executors")
	nopassthroughVolumes     = flag.String("nopassthrough-volumes", "", "Volumes on the driver which should *not* be passed through to executors")
	jvmFlags                 arrayFlags
	classpath                string
)

func init() {
	flag.StringVar(&classpath, "classpath", "", "JVM classpath. If not set, classpath will be inferred on a best-effort basis")
	flag.Var(&jvmFlags, "jvm-flag", "Flag to pass to jvm, may be specified multiple times")
}

func main() {
	flag.Parse()
	args := flag.Args()
	if classpath == "" {
		classpath = withinContainerGetClasspath()
	}

	commonConfig := launcherCommon{
		"/usr/bin/java",
		classpath,
		os.Environ(),
		jvmFlags,
	}
	var launcher sparkLauncher
	firstArg := ""
	if len(args) > 0 {
		firstArg = args[0]
	}
	switch firstArg {
	case "init":
		launcher = initLauncher{}
	case "executor":
		launcher = executorLauncher{commonConfig}
	default:
		launcher = driverLauncher{
			commonConfig,
			*driverContainer,
			withinClusterGetPodInfo(),
			*driverMainClass,
			args,
			time.Now(),
			listArgSeparator.Split(*nopassthroughLabels, -1),
			listArgSeparator.Split(*nopassthroughAnnotations, -1),
			listArgSeparator.Split(*nopassthroughEnv, -1),
			listArgSeparator.Split(*nopassthroughJvmFlags, -1),
			listArgSeparator.Split(*nopassthroughVolumes, -1),
		}
	}
	err := launcher.exec()
	if err != nil {
		panic(err)
	}
}

func (i *arrayFlags) String() string {
	return strings.Join(*i, ",")
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}
