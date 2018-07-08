load("@bazel_tools//tools/build_defs/repo:git.bzl", "git_repository")

HADOOP_VERSION="2.8.3"
MAVEN_VERSION="3.5.3"
MAVEN_MIRRORS=[
    "http://ftp.wayne.edu",
]
HADOOP_MIRRORS=[
    "http://archive.apache.org/dist",
    "http://apache.mirrors.ionfish.org",
    "http://apache.claz.org",
]

def spark_repositories():

    existing = native.existing_rules().keys()

    if "maven" not in existing:
        native.new_http_archive(
            name = "maven",
            workspace_file_content = """
workspace(name = "maven")
""",
            build_file_content = """
sh_binary(
    name = "mvn",
    srcs = ["bin/mvn"],
    data = glob(["**"]),
    visibility = ["//visibility:public"],
)
""",
            strip_prefix = "apache-maven-%s" % MAVEN_VERSION,
            urls = ["{mirror}/apache/maven/maven-3/{version}/binaries/apache-maven-{version}-bin.tar.gz".format(
                mirror = mirror,
                version = MAVEN_VERSION,
            ) for mirror in MAVEN_MIRRORS],
        )


    if "hadoop_archive" not in existing:
        native.new_http_archive(
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
            ) for mirror in HADOOP_MIRRORS]
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
