package configs

import _ "embed"

// SampleYAML is the bundled example configuration for dnsbro.
//
//go:embed config.yaml
var SampleYAML string
