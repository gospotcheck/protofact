package scala

type Config struct {
	Description            string
	GRPCPackagesVersion    string
	MavenRepoPublishTarget string
	MavenRepoHost          string
	MavenRepoUser          string
	MavenRepoPassword      string
	JarName                string
	Organization           string
	Publish                bool
	Realm                  string
	SBTVersion             string
	SBTAssemblyVersion     string
	ScalaVersion           string
}
