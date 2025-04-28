package compute

// PackageMapping represents the mapping between standard package sizes and VirtFusion resources
type PackageMapping struct {
	Memory int `json:"memory"` // Memory in MB
	Disk   int `json:"disk"`   // Disk in GB
	CPU    int `json:"cpu"`    // CPU cores
}

// StandardPackages maps standard package names to VirtFusion resource configurations
var StandardPackages = map[string]PackageMapping{
	"s-1vcpu-1gb": {
		Memory: 1024,
		Disk:   25,
		CPU:    1,
	},
	"s-2vcpu-2gb": {
		Memory: 2048,
		Disk:   50,
		CPU:    2,
	},
	"s-4vcpu-4gb": {
		Memory: 4096,
		Disk:   80,
		CPU:    4,
	},
	"s-8vcpu-8gb": {
		Memory: 8192,
		Disk:   160,
		CPU:    8,
	},
}

// GetPackageMapping returns the VirtFusion resource configuration for a given standard package size
func GetPackageMapping(packageSize string) (PackageMapping, bool) {
	mapping, exists := StandardPackages[packageSize]
	return mapping, exists
}
