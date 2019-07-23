import sbt.Keys.credentials

organization := "{{ .Organization }}"
name := "{{ .Name }}"
description := "{{ .Description }}"
scalaVersion := "{{ .ScalaVersion }}"

resolvers ++= Seq(
  "Maven Central" at "https://repo1.maven.org/maven2/",
  "{{ .Realm }}" at "{{ .MavenRepoPublishTarget }}/",
  "JitPack" at "https://jitpack.io"
)

credentials += {
  Credentials("{{ .Realm }} Realm", "{{ .MavenRepoHost }}", "{{ .MavenRepoUser }}", "{{ .MavenRepoPassword }}")
}

publishTo := Some("{{ .Realm }} Realm" at "{{ .MavenRepoPublishTarget }}")

assemblyMergeStrategy in assembly := {
  case PathList("META-INF", "native", xs@_*) => MergeStrategy.first
  case PathList("META-INF", xs@_*) => MergeStrategy.discard
  case _ => MergeStrategy.last
}

test in assembly := {}

val orgDeps = Seq()

val vendorDeps = Seq(
    "io.grpc" % "grpc-protobuf" % "{{ .GRPCPackagesVersion}}",
    "io.grpc" % "grpc-services" % "{{ .GRPCPackagesVersion }}",
    "io.grpc" % "grpc-netty-shaded" % "{{ .GRPCPackagesVersion }}",
)

libraryDependencies ++= (orgDeps ++ vendorDeps ++ testDeps)

scalaSource in Compile := baseDirectory.value / "{{ .JarDir }}"

lazy val commonSettings = Seq(
  organization := "{{ .Organization }}",
  scalaVersion := "{{ .ScalaVersion }}",
  fork in run := true
)

lazy val app = (project in file(".")).settings(commonSettings: _*)