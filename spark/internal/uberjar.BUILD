load("@io_bazel_rules_scala//scala:scala_import.bzl", "scala_import")

genrule(
    name = "scala_jar",
    srcs = ["uberjar-scala/target/uberjar-scala-1.0-SNAPSHOT.jar"],
    outs = ["uberjar-scala/target/uberjar-scala-1.0-SNAPSHOT-squashed.jar"],
    cmd = """$(location @rules_spark//spark/internal:RepackageJar) "$<" "$@" """,
    tools = ["@rules_spark//spark/internal:RepackageJar"],
)

genrule(
    name = "java_jar",
    srcs = ["uberjar-java/target/uberjar-java-1.0-SNAPSHOT.jar"],
    outs = ["uberjar-java/target/uberjar-java-1.0-SNAPSHOT-squashed.jar"],
    cmd = """$(location @rules_spark//spark/internal:RepackageJar) "$<" "$@" """,
    tools = ["@rules_spark//spark/internal:RepackageJar"],
)

scala_import(
    name = "jar",
    jars = [
        "uberjar-java/target/uberjar-java-1.0-SNAPSHOT-squashed.jar",
        "uberjar-scala/target/uberjar-scala-1.0-SNAPSHOT-squashed.jar",
    ],
    #    deps = [":java_jar"],
    #    exports = [":java_jar"],
    visibility = ["//visibility:public"],
)
