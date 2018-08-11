load("@bazel_tools//tools/build_defs/pkg:pkg.bzl", "pkg_deb", "pkg_tar")

exports_files(["uberjar.BUILD"])

java_binary(
    name = "RepackageJar",
    srcs = ["RepackageJar.java"],
    main_class = "spark.internal.RepackageJar",
    visibility = ["//visibility:public"],
)
