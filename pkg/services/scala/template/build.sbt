import sbt.Keys.credentials

organization := "{{ .Organization }}"
name := "{{ .Name }}"
description := "{{ .Description }}"
currentScalaVersion := "{{ .CurrentScalaVersion }}"
legacyScalaVersion := "{{ .LegacyScalaVersion }}"

scalaVersion := currentScalaVersion
crossScalaVersions := Seq(currentScalaVersion, legacyScalaVersion)

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

lazy val app = (project in file(".")).settings(fork in run := true)
