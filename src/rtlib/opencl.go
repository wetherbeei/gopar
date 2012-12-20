package rtlib

import (
	"fmt"
	"github.com/tones111/go-opencl/cl"
)

func PrintDebug() {
	for _, platform := range cl.Platforms {
		fmt.Println("  Platform Profile:", platform.Property(cl.PLATFORM_PROFILE))
		fmt.Println("  Platform Version:", platform.Property(cl.PLATFORM_VERSION))
		fmt.Println("  Platform Name:", platform.Property(cl.PLATFORM_NAME))
		fmt.Println("  Platform Vendor:", platform.Property(cl.PLATFORM_VENDOR))
	}
}
