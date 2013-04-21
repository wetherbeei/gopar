/*
 * Copyright © 2012 Paul Sbarra
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
	"fmt"
	"github.com/tones111/go-opencl/cl"
)

func main() {
	fmt.Println("Number of Platforms:", len(cl.Platforms))
	for _, platform := range cl.Platforms {
		fmt.Println("  Platform Profile:", platform.Property(cl.PLATFORM_PROFILE))
		fmt.Println("  Platform Version:", platform.Property(cl.PLATFORM_VERSION))
		fmt.Println("  Platform Name:", platform.Property(cl.PLATFORM_NAME))
		fmt.Println("  Platform Vendor:", platform.Property(cl.PLATFORM_VENDOR))
		fmt.Printf("  Platform Extensions: %v\n\n", platform.Property(cl.PLATFORM_EXTENSIONS))
		fmt.Println("  Platform Name:", platform.Property(cl.PLATFORM_NAME))

		fmt.Println("Number of devices:", len(platform.Devices))
		for _, device := range platform.Devices {
			fmt.Println("  Device Type:", device.Property(cl.DEVICE_TYPE))
			//fmt.Println("  Device ID:", "TODO")
			//fmt.Println("  Board name:", "TODO")
			fmt.Println("  Max compute units:", device.Property(cl.DEVICE_MAX_COMPUTE_UNITS))
			fmt.Println("  Max work items dimensions:", device.Property(cl.DEVICE_MAX_WORK_ITEM_DIMENSIONS))
			//fmt.Println("    Max work items[]", "TODO")
			fmt.Println("  Max work group size:", device.Property(cl.DEVICE_MAX_WORK_GROUP_SIZE))
			fmt.Println("  Preferred vector width char:", device.Property(cl.DEVICE_PREFERRED_VECTOR_WIDTH_CHAR))
			fmt.Println("  Preferred vector width short:", device.Property(cl.DEVICE_PREFERRED_VECTOR_WIDTH_SHORT))
			fmt.Println("  Preferred vector width int:", device.Property(cl.DEVICE_PREFERRED_VECTOR_WIDTH_INT))
			fmt.Println("  Preferred vector width long:", device.Property(cl.DEVICE_PREFERRED_VECTOR_WIDTH_LONG))
			fmt.Println("  Preferred vector width float:", device.Property(cl.DEVICE_PREFERRED_VECTOR_WIDTH_FLOAT))
			fmt.Println("  Preferred vector width double:", device.Property(cl.DEVICE_PREFERRED_VECTOR_WIDTH_DOUBLE))
			fmt.Println("  Native vector width char:", device.Property(cl.DEVICE_NATIVE_VECTOR_WIDTH_CHAR))
			fmt.Println("  Native vector width short:", device.Property(cl.DEVICE_NATIVE_VECTOR_WIDTH_SHORT))
			fmt.Println("  Native vector width int:", device.Property(cl.DEVICE_NATIVE_VECTOR_WIDTH_INT))
			fmt.Println("  Native vector width long:", device.Property(cl.DEVICE_NATIVE_VECTOR_WIDTH_LONG))
			fmt.Println("  Native vector width float:", device.Property(cl.DEVICE_NATIVE_VECTOR_WIDTH_FLOAT))
			fmt.Println("  Native vector width double:", device.Property(cl.DEVICE_NATIVE_VECTOR_WIDTH_DOUBLE))
			fmt.Printf("  Max clock frequency: %dMhz\n", device.Property(cl.DEVICE_MAX_CLOCK_FREQUENCY))
			fmt.Println("  Address bits:", device.Property(cl.DEVICE_ADDRESS_BITS))
			fmt.Println("  Max memory allocation:", device.Property(cl.DEVICE_MAX_MEM_ALLOC_SIZE))
			fmt.Println("  Image support:", device.Property(cl.DEVICE_IMAGE_SUPPORT))
			fmt.Println("  Max number of images read arguments:", device.Property(cl.DEVICE_MAX_READ_IMAGE_ARGS))
			fmt.Println("  Max number of images write arguments:", device.Property(cl.DEVICE_MAX_WRITE_IMAGE_ARGS))
			fmt.Println("  Max image 2D width:", device.Property(cl.DEVICE_IMAGE2D_MAX_WIDTH))
			fmt.Println("  Max image 2D height:", device.Property(cl.DEVICE_IMAGE2D_MAX_HEIGHT))
			fmt.Println("  Max image 3D width:", device.Property(cl.DEVICE_IMAGE3D_MAX_WIDTH))
			fmt.Println("  Max image 3D height:", device.Property(cl.DEVICE_IMAGE3D_MAX_HEIGHT))
			fmt.Println("  Max image 3D depth:", device.Property(cl.DEVICE_IMAGE3D_MAX_DEPTH))
			fmt.Println("  Max samplers within kernel:", device.Property(cl.DEVICE_MAX_SAMPLERS))
			fmt.Println("  Max size of kernel argument:", device.Property(cl.DEVICE_MAX_PARAMETER_SIZE))
			fmt.Println("  Alignment (bits) of base address:", device.Property(cl.DEVICE_MEM_BASE_ADDR_ALIGN))
			fmt.Println("  Minimum alignment (bytes) for any datatype:", device.Property(cl.DEVICE_MIN_DATA_TYPE_ALIGN_SIZE))

			/*fmt.Println("  Single precision floating point capability")
			fmt.Println("    Denorms:", "TODO")
			fmt.Println("    Quiet NaNs:", "TODO")
			fmt.Println("    Round to nearest even:", "TODO")
			fmt.Println("    Round to zero:", "TODO")
			fmt.Println("    Round to +ve and infinity:", "TODO")
			fmt.Println("    IEEE754-2008 fused multiply-add:", "TODO")
			*/

			//fmt.Println("  Cache type:", "TODO" /*device.Property(cl.DEVICE_GLOBAL_MEM_CACHE_TYPE)*/ )
			//fmt.Println("  Cache line size:", "TODO" /*device.Property(cl.DEVICE_GLOBAL_MEM_CACHELINE_SIZE)*/ )
			fmt.Println("  Cache size:", device.Property(cl.DEVICE_GLOBAL_MEM_CACHE_SIZE))
			fmt.Println("  Global memory size:", device.Property(cl.DEVICE_GLOBAL_MEM_SIZE))
			fmt.Println("  Constant buffer size:", device.Property(cl.DEVICE_MAX_CONSTANT_BUFFER_SIZE))
			fmt.Println("  Max number of constant args:", device.Property(cl.DEVICE_MAX_CONSTANT_ARGS))

			//fmt.Println("  Local memory type:", "TODO" /*device.Property(cl.DEVICE_LOCAL_MEM_TYPE)*/ )
			fmt.Println("  Local memory size:", device.Property(cl.DEVICE_LOCAL_MEM_SIZE))
			//fmt.Println("  Kernel Preferred work group size multiple:", "TODO")
			fmt.Println("  Error correction support:", device.Property(cl.DEVICE_ERROR_CORRECTION_SUPPORT))
			fmt.Println("  Unified memory for Host and Device:", device.Property(cl.DEVICE_HOST_UNIFIED_MEMORY))
			fmt.Println("  Profiling timer resolution:", device.Property(cl.DEVICE_PROFILING_TIMER_RESOLUTION))
			fmt.Println("  Little endian:", device.Property(cl.DEVICE_ENDIAN_LITTLE))
			fmt.Println("  Available:", device.Property(cl.DEVICE_AVAILABLE))
			fmt.Println("  Compiler available:", device.Property(cl.DEVICE_COMPILER_AVAILABLE))

			/*fmt.Println("  Execution capabilities:")
			fmt.Println("    Execute OpenCL kernels:", "TODO")
			fmt.Println("    Execute native function:", "TODO")
			*/

			/*fmt.Println("  Queue properties:")
			fmt.Println("    Out-of-Order:", "TODO")
			fmt.Println("    Profiling:", "TODO")
			*/

			//fmt.Println("  Platform ID:", "TODO" /* device.Property(cl.DEVICE_PLATFORM)*/ )
			fmt.Println("  Name:", device.Property(cl.DEVICE_NAME))
			fmt.Println("  Vendor:", device.Property(cl.DEVICE_VENDOR))
			fmt.Println("  Device OpenCL C version:", device.Property(cl.DEVICE_OPENCL_C_VERSION))
			fmt.Println("  Driver version:", device.Property(cl.DRIVER_VERSION))
			fmt.Println("  Profile:", device.Property(cl.DEVICE_PROFILE))
			fmt.Println("  Version:", device.Property(cl.DEVICE_VERSION))
			fmt.Println("  Extensions:", device.Property(cl.DEVICE_EXTENSIONS))
		}
	}
}
