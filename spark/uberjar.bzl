_MAVEN_VERSION = "3.5.4"
_MAVEN_TGZ_SHA256 = "ce50b1c91364cb77efe3776f756a6d92b76d9038b0a0782f7d53acf1e997a14d"
_MAVEN_MIRRORS = [
    "http://apache.claz.org",
    "http://apache.cs.utah.edu",
    "http://apache.mirrors.hoobly.com",
    "http://apache.mirrors.ionfish.org",
    "http://apache.mirrors.lucidnetworks.net",
    "http://apache.mirrors.pair.com",
    "http://apache.mirrors.tds.net",
    "http://apache.osuosl.org",
    "http://apache.spinellicreations.com",
    "http://download.nextag.com",
    "http://ftp.wayne.edu",
    "http://mirror.cc.columbia.edu",
    "http://mirror.cogentco.com",
    "http://mirror.metrocast.net",
    "http://mirror.olnevhost.net",
    "http://mirror.reverse.net",
    "http://mirrors.advancedhosters.com",
    "http://mirrors.gigenet.com",
    "http://mirrors.ibiblio.org",
    "http://mirrors.koehn.com",
    "http://mirrors.ocf.berkeley.edu",
    "http://mirrors.sonic.net",
    "http://mirrors.sorengard.com",
    "http://mirror.stjschools.org",
    "http://us.mirrors.quenda.co",
    "http://www.gtlib.gatech.edu",
    "http://www.trieuvan.com",
]

_POM_TEMPLATE = """
<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
			xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
			xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
	<modelVersion>4.0.0</modelVersion>
	<groupId>com.yxrkt</groupId>
	<artifactId>{artifact}</artifactId>
	<version>1.0-SNAPSHOT</version>
	{dependencies_xml}
	<build>
		<plugins>
           {shade_plugin_xml}
		</plugins>
	</build>
</project>"""

_SQUASH_JAR_CONTENT = """
#/usr/bin/env bash
set -euo pipefail
err_report() { echo "errexit on line $(caller)" >&2; }
trap err_report ERR

# extract jar content to temp_dir

# set timestamp of each file to 19700101000000

# package jar back up


"""

def _pom_shade_plugin(artifact_includes = [], artifact_excludes = [], shaded_packages = [], shaded_package_prefix = ""):
    """
    Return plugin> XML for maven-shade-plugin
    """
    template = """
             <plugin>
                 <groupId>org.apache.maven.plugins</groupId>
                 <artifactId>maven-shade-plugin</artifactId>
                 <version>3.1.1</version>
                 <configuration>
                     <shadedArtifactAttached>false</shadedArtifactAttached>
                     <createSourcesJar>true</createSourcesJar>
                     <artifactSet>
                         <includes>
                             {artifact_includes}
                         </includes>
                         <excludes>
                             {artifact_excludes}
                         </excludes>
                     </artifactSet>
                     <relocations>
                         {relocations}
                     </relocations>
                     <filters>
                         <filter>
                             <artifact>*:*</artifact>
                             <excludes>
                                 <exclude>META-INF/*.SF</exclude>
                                 <exclude>META-INF/*.DSA</exclude>
                                 <exclude>META-INF/*.RSA</exclude>
                                 <exclude>META-INF/maven/**</exclude>
                                 <exclude>META-INF/DEPENDENCIES</exclude>
                                 <exclude>META-INF/NOTICE</exclude>
                                 <exclude>META-INF/MANIFEST.MF</exclude>
                             </excludes>
                         </filter>
                     </filters>
                 </configuration>
                 <executions>
                     <execution>
                         <phase>package</phase>
                         <goals>
                             <goal>shade</goal>
                         </goals>
                         <configuration>
                             <transformers>
                                 <transformer implementation="org.apache.maven.plugins.shade.resource.ServicesResourceTransformer"/>
                                 <transformer implementation="org.apache.maven.plugins.shade.resource.DontIncludeResourceTransformer">
                                     <resource>log4j.properties</resource>
                                     <resource>reference.conf</resource>
                                 </transformer>
                             </transformers>
                         </configuration>
                     </execution>
                 </executions>
             </plugin>
    """

    relocation_template = """<relocation>
                        <pattern>{pattern}</pattern>
                        <shadedPattern>{shaded_pattern}</shadedPattern>
                     </relocation>"""

    return template.format(
        artifact_includes = "" if not artifact_includes else "\n                             ".join([
            "<include>%s</include>" % inc
            for inc in artifact_includes
        ]),
        artifact_excludes = "" if not artifact_excludes else "\n                             ".join([
            "<exclude>%s</exclude>" % exc
            for exc in artifact_excludes
        ]),
        relocations = "" if not shaded_packages else "\n                         ".join([
            relocation_template.format(
                pattern = pkg,
                shaded_pattern = shaded_package_prefix + pkg,
            )
            for pkg in shaded_packages
        ]),
    )

#def _pom_shade_config_relocations()

def _pom_dependencies(dependency_coords, exclude_coords):
    """
    Return the dependencies> section of 'pom.xml'
    """
    dep_template = """<dependency>
           <groupId>{group}</groupId>
           <artifactId>{artifact}</artifactId>
           <version>{version}</version>
           <type>{packaging}</type>
           <exclusions>
               {exclusions}
           </exclusions>
       </dependency>"""
    exclusion_template = """<exclusion>
               <artifactId>{artifact}</artifactId>
               <groupId>{group}</groupId>
           </exclusion>"""

    exclusions = "\n        ".join([
        exclusion_template.format(
            artifact = exc.artifact,
            group = exc.group,
        )
        for exc in exclude_coords
    ])

    dependencies_xml = []

    for dep in dependency_coords:
        if not dep.version:
            fail("Invalid maven coordinate '%s', version not specified" % dep.user_input)
        xml = dep_template.format(
            group = dep.group,
            artifact = dep.artifact,
            version = dep.version,
            packaging = dep.packaging,
            exclusions = exclusions,
        )
        dependencies_xml.append(xml)

    return """<dependencies>
           {dependencies}
       </dependencies>""".format(
        dependencies = "\n      ".join(dependencies_xml),
    )

def _settings_xml_content(ctx):
    """
    """
    template = """
<settings xmlns="http://maven.apache.org/SETTINGS/1.0.0"
  xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
  xsi:schemaLocation="http://maven.apache.org/SETTINGS/1.0.0
                      https://maven.apache.org/xsd/settings-1.0.0.xsd">

  <profiles>
    <profile>
    <id>uberjar-repository-rule</id>
	<repositories>
		{repositories}
	</repositories>
    </profile>
  </profiles>

  <activeProfiles>
    <activeProfile>uberjar-repository-rule</activeProfile>
  </activeProfiles>
</settings>
"""
    repository_template = """<repository>
        <id>{id}</id>
        <name>{name}</name>
        <url>{url}</url>
        <layout>default</layout>
        <releases>
            <enabled>true</enabled>
            <updatePolicy>never</updatePolicy>
            <checksumPolicy>fail</checksumPolicy>
        </releases>
        <snapshots>
            <enabled>true</enabled>
            <updatePolicy>always</updatePolicy>
            <checksumPolicy>fail</checksumPolicy>
        </snapshots>
    </repository>"""
    repositories = []
    for url in ctx.attr.repositories:
        urlparts = [s for s in url.split("/")[1:] if s]
        id = "-".join(urlparts).replace(".", "-")
        repository_xml = repository_template.format(
            id = id,
            name = id,
            url = url,
        )
        repositories.append(repository_xml)

    return template.format(
        repositories = "\n        ".join(repositories),
    )

def _parse_mvn_coord(coord):
    """
    Parse a colon-delimited maven coordinate into its constituent parts
    """
    parts = coord.split(":")
    num_parts = len(parts)
    packaging = "jar"
    classifier = ""
    version = ""
    if num_parts == 2:
        version = ""
    elif num_parts == 3:
        version = parts[2]
    elif num_parts == 4:
        packaging = parts[2]
        version = parts[3]
    elif num_parts == 5:
        packaging = parts[2]
        classifier = parts[3]
        version = parts[4]
    else:
        fail("Invalid maven coordinate '%s', must be formatted as 'groupId:artifactId[:packaging[:classifier]]:version'" % coord)
    return struct(
        group = parts[0],
        artifact = parts[1],
        version = version,
        packaging = packaging,
        classifier = classifier,
        user_input = coord,
    )

def _mvn(ctx, *args, **kwargs):
    arguments = ["maven/bin/mvn", "-T", "1C", "--settings", "settings.xml"]
    for arg in args:
        arguments.append(arg)
    res = ctx.execute(arguments, **kwargs)
    if res.return_code != 0:
        fail("Failed to run 'mvn {args}':\n{stdout}\n{stderr}".format(
            args = " ".join([a for a in args]),
            stdout = res.stdout,
            stderr = res.stderr,
        ))
    return res

def _write_pom(ctx, artifact, artifact_includes = [], artifact_excludes = []):
    """
    """
    dependency_coords = [_parse_mvn_coord(dep) for dep in ctx.attr.dependencies]
    exclude_coords = [_parse_mvn_coord(exc) for exc in ctx.attr.exclusions]

    pom_dependencies_xml = _pom_dependencies(dependency_coords, exclude_coords)
    pom_shade_plugin_xml = _pom_shade_plugin(
        artifact_includes = artifact_includes,
        artifact_excludes = artifact_excludes,
        shaded_packages = ctx.attr.shaded_packages,
        shaded_package_prefix = ctx.attr.shaded_packages_prefix,
    )
    pom_xml = _POM_TEMPLATE.format(
        artifact = artifact,
        dependencies_xml = pom_dependencies_xml,
        shade_plugin_xml = pom_shade_plugin_xml,
    )
    ctx.file("%s/pom.xml" % artifact, content = pom_xml, executable = False)

def _uberjar_impl(ctx):
    """
    """

    # write workspace & build files
    ctx.file("WORKSPACE", content = """workspace(name = "%s")""" % ctx.attr.name, executable = False)
    ctx.template("BUILD.bazel", ctx.attr._uberjar_build_template, executable = False, substitutions = {
        "{name}": ctx.attr.name,
    })

    # write settings xml
    settings_xml = _settings_xml_content(ctx)
    ctx.file("settings.xml", content = settings_xml, executable = False)

    # write shaded poms for scala & java + an aggregator pom
    scala_artifact_patterns = ["*:*_2.%d" % v for v in range(10, 15)]
    _write_pom(
        ctx,
        "uberjar-scala",
        artifact_includes = scala_artifact_patterns,
    )
    _write_pom(
        ctx,
        "uberjar-java",
        artifact_includes = ["*:*"],
        artifact_excludes = scala_artifact_patterns + [
            "org.scala-lang:*",
        ],
    )
    ctx.file("pom.xml", content = """
<project xmlns="http://maven.apache.org/POM/4.0.0"
  xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
  xsi:schemaLocation="http://maven.apache.org/POM/4.0.0
                      https://maven.apache.org/xsd/maven-4.0.0.xsd">
  <modelVersion>4.0.0</modelVersion>
  <groupId>com.yxrkt</groupId>
  <artifactId>uberjar</artifactId>
  <version>1.0-SNAPSHOT</version>
  <packaging>pom</packaging>
  <modules>
    <module>uberjar-java</module>
    <module>uberjar-scala</module>
  </modules>
</project>""", executable = False)

    # download maven
    ctx.download_and_extract(
        [
            "{mirror}/maven/maven-3/{version}/binaries/apache-maven-{version}-bin.tar.gz".format(
                mirror = mirror,
                version = _MAVEN_VERSION,
            )
            for mirror in _MAVEN_MIRRORS
        ],
        output = "maven",
        sha256 = _MAVEN_TGZ_SHA256,
        stripPrefix = "apache-maven-%s" % _MAVEN_VERSION,
    )
    ctx.execute(["echo", "Fetching '@%s//:jar' (this might take a minute)..." % ctx.attr.name], quiet = False)
    _mvn(ctx, "package")

maven_uberjar = repository_rule(
    implementation = _uberjar_impl,
    attrs = {
        "dependencies": attr.string_list(default = []),
        "exclusions": attr.string_list(default = []),
        "repositories": attr.string_list(default = []),
        "shaded_packages": attr.string_list(default = []),
        "shaded_packages_prefix": attr.string(default = "shaded."),
        "_uberjar_build_template": attr.label(
            default = "@rules_spark//spark/internal:uberjar.BUILD",
            allow_single_file = True,
        ),
    },
)
