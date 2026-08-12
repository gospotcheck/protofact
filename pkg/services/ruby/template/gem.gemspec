Gem::Specification.new do |spec|
  spec.name          = "{{ .GemName }}"
  spec.version       = "{{ .Version }}"
  spec.authors       = ["{{ .Authors }}"]
  spec.email         = ["{{ .Email }}"]

  spec.summary       = "Gem of proto files for {{ .GemName }}"
  spec.description   = "Gem of proto files for {{ .GemName }}"
  spec.homepage      = "{{ .Homepage }}"
  spec.files         = Dir["lib/**/*.rb"]

  # allowed_push_host is intentionally left unset: Protofact publishes this gem
  # to multiple hosts (JFrog and Google Artifact Registry) during the migration,
  # and metadata['allowed_push_host'] can only whitelist a single host. Pushes are
  # always made to an explicit --host, so this is not published to RubyGems.org.

  spec.require_paths = ["lib"]

  spec.add_runtime_dependency 'grpc', '~> 1.52'
end