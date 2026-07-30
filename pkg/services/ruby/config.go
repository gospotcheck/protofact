package ruby

type Config struct {
	Authors     string
	Email       string
	GemRepoUser string
	GemRepoPass string
	GemRepoHost string
	// GARHost is the Google Artifact Registry gem repository URL. When set (and
	// Publish is true) the built gem is also pushed to GAR, authenticating with a
	// short-lived access token from the GCE/GKE metadata server (Workload Identity).
	// This enables publishing to both JFrog and GAR during the migration.
	GARHost     string
	GemName     string
	GRPCVersion string
	Homepage    string
	Publish     bool
}
