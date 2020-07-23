import sbt.Keys.credentials

organization := "{{ .Organization }}"
name := "{{ .Name }}"
description := "{{ .Description }}"
scalaVersion := "{{ .ScalaVersion }}"

resolvers ++= Seq(
  "Maven Central" at "https://repo1.maven.org/maven2/",
  "{{ .Realm }}" at "{{ .MavenRepoPublishTarget }}/"
)

credentials += {
  Credentials("{{ .Realm }} Realm", "{{ .MavenRepoHost }}", "{{ .MavenRepoUser }}", "{{ .MavenRepoPassword }}")
}

publishTo := Some("{{ .Realm }} Realm" at "{{ .MavenRepoPublishTarget }}")

val orgDeps = Seq()
val vendorDeps = Seq()
val testDeps = Seq()

libraryDependencies ++= (orgDeps ++ vendorDeps ++ testDeps)

scalaSource in Compile := baseDirectory.value / "{{ .JarDir }}"

lazy val commonSettings = Seq(
  organization := "{{ .Organization }}",
  scalaVersion := "{{ .ScalaVersion }}",
  fork in run := true
)

lazy val app = (project in file(".")).settings(commonSettings: _*)
