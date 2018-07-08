## Usage

#### WORKSPACE
```python
load("@bazel_tools//tools/build_defs/repo:git.bzl", "git_repository")

git_repository(
    name = "rules_spark",
    remote = "git@github.com:ceason/rules_spark.git",
    commit = "{MASTER}", # to get latest: git ls-remote git@github.com:ceason/rules_spark.git refs/heads/master|cut -f1
)

load("@rules_spark//spark:dependencies.bzl", "spark_repositories")

spark_repositories()

load("@io_bazel_rules_go//go:def.bzl", "go_rules_dependencies", "go_register_toolchains")

go_rules_dependencies()

go_register_toolchains()

load("@io_bazel_rules_scala//scala:scala.bzl", "scala_repositories")

scala_repositories()

load("@io_bazel_rules_scala//scala:toolchains.bzl", "scala_register_toolchains")

scala_register_toolchains()

load(
    "@io_bazel_rules_docker//scala:image.bzl",
    scala_image_repositories = "repositories",
)

scala_image_repositories()

load("@io_bazel_rules_k8s//k8s:k8s.bzl", "k8s_repositories")

k8s_repositories()
```

#### BUILD
```python
load("@rules_spark//spark:scala.bzl", "spark_scala_image")
load("@io_bazel_rules_k8s//k8s:object.bzl", "k8s_object")

spark_scala_image(
    name = "streamingwordcount",
    srcs = ["Main.scala"],
    layers = [
        "@rules_spark//jar_bundle",
    ],
    main_class = "com.example.streamingwordcount.Main",
)

k8s_object(
    name = "deployment",
    cluster = "{STABLE_K8S_CLUSTER}",
    image_chroot = "{STABLE_IMAGE_CHROOT}",
    images = {
        "streamingwordcount:dev": ":streamingwordcount",
    },
    kind = "deployment",
    template = ":deployment.yaml",
)
```