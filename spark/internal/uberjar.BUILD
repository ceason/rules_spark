load("@io_bazel_rules_scala//scala:scala_import.bzl", "scala_import")

_REPO_NAME = repository_name()[1:]

[genrule(
    name = "squashed/uberjar-%s%s" % (lang, classifier),
    srcs = ["uberjar-{lang}/target/uberjar-{lang}-1.0-SNAPSHOT{classifier}.jar".format(
        lang = lang,
        classifier = classifier,
    )],
    outs = ["squashed/%s-%s%s.jar" % (_REPO_NAME, lang, classifier)],
    cmd = """$(location @rules_spark//spark/internal:RepackageJar) "$<" "$@" """,
    tools = ["@rules_spark//spark/internal:RepackageJar"],
) for lang in ["java", "scala"] for classifier in ["", "-sources"]]

filegroup(
    name = "file",
    srcs = [
        ":squashed/uberjar-java",
        ":squashed/uberjar-scala",
        ":squashed/uberjar-java-sources",
        ":squashed/uberjar-scala-sources",
        #        "uberjar-java/target/uberjar-java-1.0-SNAPSHOT-sources.jar",
        #        "uberjar-scala/target/uberjar-scala-1.0-SNAPSHOT-sources.jar",
    ],
)

scala_import(
    name = "jar",
    jars = [
        ":file",
    ],
    #    deps = [":java_jar"],
    #    exports = [":java_jar"],
    visibility = ["//visibility:public"],
)
