
This is the launcher for spark docker images running on k8s

- Removes the need for spark-submit
- Wraps both executor & driver spark processes

- Figures out launch arguments for executor/driver
    - Driver:
      - Passes through annotations & labels (_except_ `spark-role`) to executors via `spark.kubernetes.executor.(label|annotation).[Name]` java property
      - Passes through env vars to executors via `spark.executorEnv.[EnvironmentVariableName]` java property
        - ?? Should this be filtered by prefix??
      - Mirrors mounted secrets to executors via `spark.kubernetes.executor.secrets.[SecretName]` java property

    - Driver & Executor:
      - Sets JVM memory limits based on pod requests/limits (accounting for overhead)
      - Reads classpath from `CLASSPATH_FILE` environment variable (baked in at image creation)

- Calls out to `syscall.Exec` to replace current process with
