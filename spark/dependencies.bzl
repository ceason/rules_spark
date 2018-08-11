load("@bazel_tools//tools/build_defs/repo:git.bzl", "git_repository")
load("@io_bazel_rules_docker//container:pull.bzl", "container_pull")
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive", "http_file")
load("//spark:uberjar.bzl", "maven_uberjar")

SPARK_VERSION = "2.4.0-palantir.20"

HADOOP_VERSION = "2.8.3"
HADOOP_TGZ_SHA256 = "e8bf9a53337b1dca3b152b0a5b5e277dc734e76520543e525c301a050bb27eae"
HADOOP_MIRRORS = [
    "http://archive.apache.org/dist",
    "http://apache.mirrors.ionfish.org",
    "http://apache.claz.org",
]

def spark_repositories():
    existing = native.existing_rules()

    if "hadoop_archive" not in existing:
        http_archive(
            name = "hadoop_archive",
            workspace_file_content = """
workspace(name = "hadoop_archive")
""",
            build_file_content = """
filegroup (
   name = "lib_native",
   srcs = glob(["lib/native/*"]),
   visibility = ["//visibility:public"],
)
""",
            strip_prefix = "hadoop-%s" % HADOOP_VERSION,
            urls = ["{mirror}/hadoop/common/hadoop-{version}/hadoop-{version}.tar.gz".format(
                mirror = mirror,
                version = HADOOP_VERSION,
            ) for mirror in HADOOP_MIRRORS],
            sha256 = HADOOP_TGZ_SHA256,
        )

    if "io_bazel_rules_go" not in existing:
        git_repository(
            name = "io_bazel_rules_go",
            remote = "git@github.com:bazelbuild/rules_go.git",
            tag = "0.12.0",
        )

    if "bazel_gazelle" not in existing:
        git_repository(
            name = "bazel_gazelle",
            remote = "git@github.com:bazelbuild/bazel-gazelle.git",
            tag = "0.12.0",
        )

    if "io_bazel_rules_scala" not in existing:
        git_repository(
            name = "io_bazel_rules_scala",
            commit = "e199467847dec882cb3d8af412bd939544cc3177",
            remote = "git@github.com:bazelbuild/rules_scala.git",
        )

    if "io_bazel_rules_docker" not in existing:
        git_repository(
            name = "io_bazel_rules_docker",
            commit = "7401cb256222615c497c0dee5a4de5724a4f4cc7",
            remote = "git@github.com:bazelbuild/rules_docker.git",
        )

    if "io_bazel_rules_k8s" not in existing:
        git_repository(
            name = "io_bazel_rules_k8s",
            commit = "d6e1b65317246fe044482f9e042556c77e6893b8",
            remote = "git@github.com:bazelbuild/rules_k8s.git",
        )

    if "openjdk_8" not in existing:
        container_pull(
            name = "openjdk_8",
            registry = "gcr.io",
            repository = "distroless/java",
            tag = "debug", # 'debug' tag is necessary because it contains cli commands required for hadoop
            visibility = ["//visibility:public"],
        )

    if "rules_spark_jar_bundle" not in existing:
        maven_uberjar(
            name = "rules_spark_jar_bundle",
            dependencies = [
                "org.apache.spark:spark-dist_2.11-hadoop-palantir:pom:" + SPARK_VERSION,
                "org.apache.spark:spark-streaming-kafka-0-10_2.11:" + SPARK_VERSION,
                "org.apache.spark:spark-sql-kafka-0-10_2.11:" + SPARK_VERSION,
                "org.apache.spark:spark-streaming-kinesis-asl_2.11:" + SPARK_VERSION,
                "com.amazonaws:aws-java-sdk:1.11.45",
                "com.squareup.okio:okio:1.13.0",
            ],
            repositories = [
                "https://dl.bintray.com/palantir/releases/",
            ],
            shaded_packages = [
                "com.google.protobuf",
                "com.google.guava",
            ],
        )
