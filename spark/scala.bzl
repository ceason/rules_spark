load("@io_bazel_rules_docker//container:image.bzl", "container_image")
load("@io_bazel_rules_scala//scala:scala.bzl", "scala_binary")
load(
    "@io_bazel_rules_docker//java:image.bzl",
    "DEFAULT_JAVA_BASE",
    "jar_dep_layer",
    "jar_app_layer",
)


def spark_scala_image(name, main_class, base=None,
                deps=[], runtime_deps=[], layers=[], jvm_flags=[], files=[],
                **kwargs):
  """Builds a container image overlaying the scala_binary.

  Args:
    layers: Augments "deps" with dependencies that should be put into
           their own layers.
    **kwargs: See scala_binary.
  """

  binary_name = name + ".binary"
  app_layer_name = "%s.%d" % (name, len(layers))

  scala_binary(name=binary_name, main_class=main_class,
                      # If the rule is turning a JAR built with java_library into
                      # a binary, then it will appear in runtime_deps.  We are
                      # not allowed to pass deps (even []) if there is no srcs
                      # kwarg.
                      deps=(deps + layers) or None, runtime_deps=runtime_deps,
                      jvm_flags=jvm_flags, **kwargs)


  base = base or DEFAULT_JAVA_BASE
  for index, dep in enumerate(layers):
    this_name = "%s.%d" % (name, index)
    jar_dep_layer(name=this_name, base=base, dep=dep)
    base = this_name

  directory = "/app"

  visibility = kwargs.get('visibility', None)
  tags = kwargs.get('tags', None)
  jar_app_layer(name=app_layer_name, base=base, binary=binary_name,
                 main_class=main_class, jvm_flags=jvm_flags,
                 deps=deps, runtime_deps=runtime_deps, jar_layers=layers,
                 visibility=visibility, tags=tags, directory=directory)

  container_image(
      name = name,
      base = app_layer_name,
      cmd = None,
      entrypoint = [
          "/launcher",
          "-driver-main-class=%s" % main_class,
      ] + ["-jvm-flag=" + flag for flag in jvm_flags],
      directory = directory,
      visibility = visibility,
      layers = [
          "@rules_spark//spark:launcher.layer",
          "@rules_spark//spark:hadoop_native.layer",
      ],
      files = files,
  )



