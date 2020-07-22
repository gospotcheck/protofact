logLevel := sbt.Level.Warn

addSbtPlugin("com.thesamet" % "sbt-protoc" % "{{ .SBTProtocPluginPackageVersion }}")

libraryDependencies += "com.thesamet.scalapb" %% "compilerplugin" % "{{ .ScalaPBRuntimePackageVersion  }}"
