package compute

// TemplateMapping maps standard image names to VirtFusion template names
var TemplateMapping = map[string]string{
	"ubuntu-22.04":     "ubuntu-22.04-x64",
	"ubuntu-22.04-x64": "ubuntu-22.04-x64",
	"ubuntu-20.04":     "ubuntu-20.04-x64",
	"ubuntu-20.04-x64": "ubuntu-20.04-x64",
	"debian-11":        "debian-11-x64",
	"debian-11-x64":    "debian-11-x64",
	"debian-10":        "debian-10-x64",
	"debian-10-x64":    "debian-10-x64",
	"centos-7":         "centos-7-x64",
	"centos-7-x64":     "centos-7-x64",
	"rocky-8":          "rocky-8-x64",
	"rocky-8-x64":      "rocky-8-x64",
	"rocky-9":          "rocky-9-x64",
	"rocky-9-x64":      "rocky-9-x64",
}

// GetTemplateMapping returns the VirtFusion template name for a given standard image name
func GetTemplateMapping(imageName string) (string, bool) {
	template, exists := TemplateMapping[imageName]
	return template, exists
}
