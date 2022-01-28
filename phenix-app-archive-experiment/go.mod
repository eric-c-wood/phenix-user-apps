module phenix-app-archive-experiment

go 1.14

replace phenix-apps => github.com/sandia-minimega/phenix-apps/src/go v0.0.0-20211026161558-a10db861435d

replace phenix => github.com/sandia-minimega/phenix/src/go v0.0.0-20220110161443-a353cbf362b3

require (
	github.com/mitchellh/mapstructure v1.4.1
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	phenix v0.0.0-00010101000000-000000000000
	phenix-apps v0.0.0-00010101000000-000000000000
)
