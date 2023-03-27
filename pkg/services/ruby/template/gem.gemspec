Gem::Specification.new do |spec|
  spec.name          = "{{ .GemName }}"
  spec.version       = "{{ .Version }}"
  spec.authors       = ["{{ .Authors }}"]
  spec.email         = ["{{ .Email }}"]

  spec.summary       = "Gem of proto files for {{ .GemName }}"
  spec.description   = "Gem of proto files for {{ .GemName }}"
  spec.homepage      = "{{ .Homepage }}"
  spec.files         = Dir["lib/**/*.rb"]

  # Prevent pushing this gem to RubyGems.org. To allow pushes either set the 'allowed_push_host'
  # to allow pushing to a single host or delete this section to allow pushing to any host.
  if spec.respond_to?(:metadata)
    spec.metadata["allowed_push_host"] = "{{ .GemRepoHost }}"
  else
    raise "RubyGems 2.0 or newer is required to protect against " \
      "public gem pushes."
  end

  spec.require_paths = ["lib"]

  spec.add_runtime_dependency 'grpc', '~> 1.52'
end