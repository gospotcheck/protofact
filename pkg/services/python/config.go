package python

// Config holds the values needed to package and publish a Python
// distribution. Unlike ruby/npm, the source repo already ships a
// complete pyproject.toml per SourceDir (name, deps, namespace package
// config), so this service only needs to know which subdirectory to
// package and where/how to publish it - it does not template a manifest.
type Config struct {
	Publish      bool
	SourceDir    string
	RegistryURL  string
	RegistryUser string
	RegistryPass string
}
