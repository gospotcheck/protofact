package scala

type Config struct {
	Description                   string
	MavenRepoPublishTarget        string
	MavenRepoHost                 string
	MavenRepoUser                 string
	MavenRepoPassword             string
	JarName                       string
	Organization                  string
	Publish                       bool
	Realm                         string
	SBTVersion                    string
	SBTProtocPluginPackageVersion string
	ScalaVersion                  string
	LegacyScalaVersion            string
	ScalaPBRuntimePackageVersion  string
}
