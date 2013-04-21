/*
 * Copyright © 2012 go-opencl authors
 *
 * This file is part of go-opencl.
 *
 * go-opencl is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * go-opencl is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with go-opencl.  If not, see <http://www.gnu.org/licenses/>.
 */

package cl

/*
#cgo CFLAGS: -I CL
#cgo LDFLAGS: -lOpenCL

#include "CL/opencl.h"

*/
import "C"

import (
	"unsafe"
)

type DeviceType C.cl_device_type

const (
	DEVICE_TYPE_DEFAULT     DeviceType = C.CL_DEVICE_TYPE_DEFAULT
	DEVICE_TYPE_CPU         DeviceType = C.CL_DEVICE_TYPE_CPU
	DEVICE_TYPE_GPU         DeviceType = C.CL_DEVICE_TYPE_GPU
	DEVICE_TYPE_ACCELERATOR DeviceType = C.CL_DEVICE_TYPE_ACCELERATOR
	//DEVICE_TYPE_CUSTOM      DeviceType = C.CL_DEVICE_TYPE_CUSTOM
	DEVICE_TYPE_ALL DeviceType = C.CL_DEVICE_TYPE_ALL
)

func (t DeviceType) String() string {
	mesg := deviceTypeMesg[t]
	if mesg == "" {
		return t.String()
	}
	return mesg
}

var deviceTypeMesg = map[DeviceType]string{
	DEVICE_TYPE_CPU:         "CPU",
	DEVICE_TYPE_GPU:         "GPU",
	DEVICE_TYPE_ACCELERATOR: "Accelerator",
	//DEVICE_TYPE_CUSTOM: "Custom",
}

type DeviceProperty C.cl_device_info

const (
	DEVICE_ADDRESS_BITS DeviceProperty = C.CL_DEVICE_ADDRESS_BITS
	DEVICE_AVAILABLE    DeviceProperty = C.CL_DEVICE_AVAILABLE
	//DEVICE_BUILT_IN_KERNELS DeviceProperty = C.CL_DEVICE_BUILT_IN_KERNELS
	DEVICE_COMPILER_AVAILABLE DeviceProperty = C.CL_DEVICE_COMPILER_AVAILABLE
	//DEVICE_DOUBLE_FP_CONFIG          DeviceProperty = C.CL_DEVICE_DOUBLE_FP_CONFIG
	DEVICE_ENDIAN_LITTLE             DeviceProperty = C.CL_DEVICE_ENDIAN_LITTLE
	DEVICE_ERROR_CORRECTION_SUPPORT  DeviceProperty = C.CL_DEVICE_ERROR_CORRECTION_SUPPORT
	DEVICE_EXECUTION_CAPABILITIES    DeviceProperty = C.CL_DEVICE_EXECUTION_CAPABILITIES
	DEVICE_EXTENSIONS                DeviceProperty = C.CL_DEVICE_EXTENSIONS
	DEVICE_GLOBAL_MEM_CACHE_SIZE     DeviceProperty = C.CL_DEVICE_GLOBAL_MEM_CACHE_SIZE
	DEVICE_GLOBAL_MEM_CACHE_TYPE     DeviceProperty = C.CL_DEVICE_GLOBAL_MEM_CACHE_TYPE
	DEVICE_GLOBAL_MEM_CACHELINE_SIZE DeviceProperty = C.CL_DEVICE_GLOBAL_MEM_CACHELINE_SIZE
	DEVICE_GLOBAL_MEM_SIZE           DeviceProperty = C.CL_DEVICE_GLOBAL_MEM_SIZE
	//DEVICE_HALF_FP_CONFIG                DeviceProperty = C.CL_DEVICE_HALF_FP_CONFIG
	DEVICE_HOST_UNIFIED_MEMORY DeviceProperty = C.CL_DEVICE_HOST_UNIFIED_MEMORY
	DEVICE_IMAGE_SUPPORT       DeviceProperty = C.CL_DEVICE_IMAGE_SUPPORT
	DEVICE_IMAGE2D_MAX_HEIGHT  DeviceProperty = C.CL_DEVICE_IMAGE2D_MAX_HEIGHT
	DEVICE_IMAGE2D_MAX_WIDTH   DeviceProperty = C.CL_DEVICE_IMAGE2D_MAX_WIDTH
	DEVICE_IMAGE3D_MAX_DEPTH   DeviceProperty = C.CL_DEVICE_IMAGE3D_MAX_DEPTH
	DEVICE_IMAGE3D_MAX_HEIGHT  DeviceProperty = C.CL_DEVICE_IMAGE3D_MAX_HEIGHT
	DEVICE_IMAGE3D_MAX_WIDTH   DeviceProperty = C.CL_DEVICE_IMAGE3D_MAX_WIDTH
	//DEVICE_IMAGE_MAX_BUFFER_SIZE      DeviceProperty = C.CL_DEVICE_IMAGE_MAX_BUFFER_SIZE
	//DEVICE_IMAGE_MAX_ARRAY_SIZE       DeviceProperty = C.CL_DEVICE_IMAGE_MAX_ARRAY_SIZE
	//DEVICE_LINKER_AVAILABLE           DeviceProperty = C.CL_DEVICE_LINKER_AVAILABLE
	DEVICE_LOCAL_MEM_SIZE             DeviceProperty = C.CL_DEVICE_LOCAL_MEM_SIZE
	DEVICE_LOCAL_MEM_TYPE             DeviceProperty = C.CL_DEVICE_LOCAL_MEM_TYPE
	DEVICE_MAX_CLOCK_FREQUENCY        DeviceProperty = C.CL_DEVICE_MAX_CLOCK_FREQUENCY
	DEVICE_MAX_COMPUTE_UNITS          DeviceProperty = C.CL_DEVICE_MAX_COMPUTE_UNITS
	DEVICE_MAX_CONSTANT_ARGS          DeviceProperty = C.CL_DEVICE_MAX_CONSTANT_ARGS
	DEVICE_MAX_CONSTANT_BUFFER_SIZE   DeviceProperty = C.CL_DEVICE_MAX_CONSTANT_BUFFER_SIZE
	DEVICE_MAX_MEM_ALLOC_SIZE         DeviceProperty = C.CL_DEVICE_MAX_MEM_ALLOC_SIZE
	DEVICE_MAX_PARAMETER_SIZE         DeviceProperty = C.CL_DEVICE_MAX_PARAMETER_SIZE
	DEVICE_MAX_READ_IMAGE_ARGS        DeviceProperty = C.CL_DEVICE_MAX_READ_IMAGE_ARGS
	DEVICE_MAX_SAMPLERS               DeviceProperty = C.CL_DEVICE_MAX_SAMPLERS
	DEVICE_MAX_WORK_GROUP_SIZE        DeviceProperty = C.CL_DEVICE_MAX_WORK_GROUP_SIZE
	DEVICE_MAX_WORK_ITEM_DIMENSIONS   DeviceProperty = C.CL_DEVICE_MAX_WORK_ITEM_DIMENSIONS
	DEVICE_MAX_WORK_ITEM_SIZES        DeviceProperty = C.CL_DEVICE_MAX_WORK_ITEM_SIZES
	DEVICE_MAX_WRITE_IMAGE_ARGS       DeviceProperty = C.CL_DEVICE_MAX_WRITE_IMAGE_ARGS
	DEVICE_MEM_BASE_ADDR_ALIGN        DeviceProperty = C.CL_DEVICE_MEM_BASE_ADDR_ALIGN
	DEVICE_MIN_DATA_TYPE_ALIGN_SIZE   DeviceProperty = C.CL_DEVICE_MIN_DATA_TYPE_ALIGN_SIZE
	DEVICE_NAME                       DeviceProperty = C.CL_DEVICE_NAME
	DEVICE_NATIVE_VECTOR_WIDTH_CHAR   DeviceProperty = C.CL_DEVICE_NATIVE_VECTOR_WIDTH_CHAR
	DEVICE_NATIVE_VECTOR_WIDTH_SHORT  DeviceProperty = C.CL_DEVICE_NATIVE_VECTOR_WIDTH_SHORT
	DEVICE_NATIVE_VECTOR_WIDTH_INT    DeviceProperty = C.CL_DEVICE_NATIVE_VECTOR_WIDTH_INT
	DEVICE_NATIVE_VECTOR_WIDTH_LONG   DeviceProperty = C.CL_DEVICE_NATIVE_VECTOR_WIDTH_LONG
	DEVICE_NATIVE_VECTOR_WIDTH_FLOAT  DeviceProperty = C.CL_DEVICE_NATIVE_VECTOR_WIDTH_FLOAT
	DEVICE_NATIVE_VECTOR_WIDTH_DOUBLE DeviceProperty = C.CL_DEVICE_NATIVE_VECTOR_WIDTH_DOUBLE
	DEVICE_NATIVE_VECTOR_WIDTH_HALF   DeviceProperty = C.CL_DEVICE_NATIVE_VECTOR_WIDTH_HALF
	DEVICE_OPENCL_C_VERSION           DeviceProperty = C.CL_DEVICE_OPENCL_C_VERSION
	//DEVICE_PARENT_DEVICE              DeviceProperty = C.CL_DEVICE_PARENT_DEVICE
	//DEVICE_PARTITION_MAX_SUB_DEVICES     DeviceProperty = C.CL_DEVICE_PARTITION_MAX_SUB_DEVICES
	//DEVICE_PARTITION_PROPERTIES          DeviceProperty = C.CL_DEVICE_PARTITION_PROPERTIES
	//DEVICE_PARTITION_AFFINITY_DOMAIN     DeviceProperty = C.CL_DEVICE_PARTITION_AFFINITY_DOMAIN
	//DEVICE_PARTITION_TYPE                DeviceProperty = C.CL_DEVICE_PARTITION_TYPE
	DEVICE_PLATFORM                      DeviceProperty = C.CL_DEVICE_PLATFORM
	DEVICE_PREFERRED_VECTOR_WIDTH_CHAR   DeviceProperty = C.CL_DEVICE_PREFERRED_VECTOR_WIDTH_CHAR
	DEVICE_PREFERRED_VECTOR_WIDTH_SHORT  DeviceProperty = C.CL_DEVICE_PREFERRED_VECTOR_WIDTH_SHORT
	DEVICE_PREFERRED_VECTOR_WIDTH_INT    DeviceProperty = C.CL_DEVICE_PREFERRED_VECTOR_WIDTH_INT
	DEVICE_PREFERRED_VECTOR_WIDTH_LONG   DeviceProperty = C.CL_DEVICE_PREFERRED_VECTOR_WIDTH_LONG
	DEVICE_PREFERRED_VECTOR_WIDTH_FLOAT  DeviceProperty = C.CL_DEVICE_PREFERRED_VECTOR_WIDTH_FLOAT
	DEVICE_PREFERRED_VECTOR_WIDTH_DOUBLE DeviceProperty = C.CL_DEVICE_PREFERRED_VECTOR_WIDTH_DOUBLE
	DEVICE_PREFERRED_VECTOR_WIDTH_HALF   DeviceProperty = C.CL_DEVICE_PREFERRED_VECTOR_WIDTH_HALF
	//DEVICE_PRINTF_BUFFER_SIZE            DeviceProperty = C.CL_DEVICE_PRINTF_BUFFER_SIZE
	//DEVICE_PREFERRED_INTEROP_USER_SYNC   DeviceProperty = C.CL_DEVICE_PREFERRED_INTEROP_USER_SYNC
	DEVICE_PROFILE                    DeviceProperty = C.CL_DEVICE_PROFILE
	DEVICE_PROFILING_TIMER_RESOLUTION DeviceProperty = C.CL_DEVICE_PROFILING_TIMER_RESOLUTION
	DEVICE_QUEUE_PROPERTIES           DeviceProperty = C.CL_DEVICE_QUEUE_PROPERTIES
	//DEVICE_REFERENCE_COUNT            DeviceProperty = C.CL_DEVICE_REFERENCE_COUNT
	DEVICE_SINGLE_FP_CONFIG DeviceProperty = C.CL_DEVICE_SINGLE_FP_CONFIG
	DEVICE_TYPE             DeviceProperty = C.CL_DEVICE_TYPE
	DEVICE_VENDOR           DeviceProperty = C.CL_DEVICE_VENDOR
	DEVICE_VENDOR_ID        DeviceProperty = C.CL_DEVICE_VENDOR_ID
	DEVICE_VERSION          DeviceProperty = C.CL_DEVICE_VERSION
	DRIVER_VERSION          DeviceProperty = C.CL_DRIVER_VERSION
)

type Device struct {
	id         C.cl_device_id
	properties map[DeviceProperty]interface{}
}

func (d *Device) Property(prop DeviceProperty) interface{} {
	if value, ok := d.properties[prop]; ok {
		return value
	}

	var data interface{}
	var length C.size_t
	var ret C.cl_int

	switch prop {
	case DEVICE_AVAILABLE,
		DEVICE_COMPILER_AVAILABLE,
		DEVICE_ENDIAN_LITTLE,
		DEVICE_ERROR_CORRECTION_SUPPORT,
		DEVICE_HOST_UNIFIED_MEMORY,
		DEVICE_IMAGE_SUPPORT:
		//DEVICE_LINKER_AVAILABLE,
		//DEVICE_PREFERRED_INTEROP_USER_SYNC:
		var val C.cl_bool
		ret = C.clGetDeviceInfo(d.id, C.cl_device_info(prop), C.size_t(unsafe.Sizeof(val)), unsafe.Pointer(&val), &length)
		data = val == C.CL_TRUE

	case DEVICE_ADDRESS_BITS,
		DEVICE_MAX_CLOCK_FREQUENCY,
		DEVICE_MAX_COMPUTE_UNITS,
		DEVICE_MAX_CONSTANT_ARGS,
		DEVICE_MAX_READ_IMAGE_ARGS,
		DEVICE_MAX_SAMPLERS,
		DEVICE_MAX_WORK_ITEM_DIMENSIONS,
		DEVICE_MAX_WRITE_IMAGE_ARGS,
		DEVICE_MEM_BASE_ADDR_ALIGN,
		DEVICE_MIN_DATA_TYPE_ALIGN_SIZE,
		DEVICE_NATIVE_VECTOR_WIDTH_CHAR,
		DEVICE_NATIVE_VECTOR_WIDTH_SHORT,
		DEVICE_NATIVE_VECTOR_WIDTH_INT,
		DEVICE_NATIVE_VECTOR_WIDTH_LONG,
		DEVICE_NATIVE_VECTOR_WIDTH_FLOAT,
		DEVICE_NATIVE_VECTOR_WIDTH_DOUBLE,
		DEVICE_NATIVE_VECTOR_WIDTH_HALF,
		//DEVICE_PARTITION_MAX_SUB_DEVICES,
		DEVICE_PREFERRED_VECTOR_WIDTH_CHAR,
		DEVICE_PREFERRED_VECTOR_WIDTH_SHORT,
		DEVICE_PREFERRED_VECTOR_WIDTH_INT,
		DEVICE_PREFERRED_VECTOR_WIDTH_LONG,
		DEVICE_PREFERRED_VECTOR_WIDTH_FLOAT,
		DEVICE_PREFERRED_VECTOR_WIDTH_DOUBLE,
		DEVICE_PREFERRED_VECTOR_WIDTH_HALF,
		//DEVICE_REFERENCE_COUNT,
		DEVICE_VENDOR_ID:
		var val C.cl_uint
		ret = C.clGetDeviceInfo(d.id, C.cl_device_info(prop), C.size_t(unsafe.Sizeof(val)), unsafe.Pointer(&val), &length)
		data = val

	case DEVICE_IMAGE2D_MAX_HEIGHT,
		DEVICE_IMAGE2D_MAX_WIDTH,
		DEVICE_IMAGE3D_MAX_DEPTH,
		DEVICE_IMAGE3D_MAX_HEIGHT,
		DEVICE_IMAGE3D_MAX_WIDTH,
		//DEVICE_IMAGE_MAX_BUFFER_SIZE,
		//DEVICE_IMAGE_MAX_ARRAY_SIZE,
		DEVICE_MAX_PARAMETER_SIZE,
		DEVICE_MAX_WORK_GROUP_SIZE,
		//DEVICE_PRINTF_BUFFER_SIZE,
		DEVICE_PROFILING_TIMER_RESOLUTION:
		var val C.size_t
		ret = C.clGetDeviceInfo(d.id, C.cl_device_info(prop), C.size_t(unsafe.Sizeof(val)), unsafe.Pointer(&val), &length)
		data = val

	case DEVICE_GLOBAL_MEM_CACHE_SIZE,
		DEVICE_GLOBAL_MEM_SIZE,
		DEVICE_LOCAL_MEM_SIZE,
		DEVICE_MAX_CONSTANT_BUFFER_SIZE,
		DEVICE_MAX_MEM_ALLOC_SIZE:
		var val C.cl_ulong
		ret = C.clGetDeviceInfo(d.id, C.cl_device_info(prop), C.size_t(unsafe.Sizeof(val)), unsafe.Pointer(&val), &length)
		data = val

	/*case DEVICE_PLATFORM:
	var val C.cl_platform_id
	ret = C.clGetDeviceInfo(d.id, C.cl_device_info(prop), C.size_t(unsafe.Sizeof(val)), unsafe.Pointer(&val), &length)
	data = Platform{id: val}*/

	/*case DEVICE_PARENT_DEVICE:
	var val C.cl_device_id
	ret = C.clGetDeviceInfo(d.id, C.cl_device_info(prop), C.size_t(unsafe.Sizeof(val)), unsafe.Pointer(&val), &length)
	data = Device{id: val}*/

	case DEVICE_TYPE:
		var val C.cl_device_type
		ret = C.clGetDeviceInfo(d.id, C.cl_device_info(prop), C.size_t(unsafe.Sizeof(val)), unsafe.Pointer(&val), &length)
		data = DeviceType(val)

	case //DEVICE_BUILT_IN_KERNELS,
		DEVICE_EXTENSIONS,
		DEVICE_NAME,
		DEVICE_OPENCL_C_VERSION,
		DEVICE_PROFILE,
		DEVICE_VENDOR,
		DEVICE_VERSION,
		DRIVER_VERSION:
		if ret = C.clGetDeviceInfo(d.id, C.cl_device_info(prop), 0, nil, &length); ret != C.CL_SUCCESS || length < 1 {
			data = ""
			break
		}

		buf := make([]C.char, length)
		if ret = C.clGetDeviceInfo(d.id, C.cl_device_info(prop), length, unsafe.Pointer(&buf[0]), &length); ret != C.CL_SUCCESS || length < 1 {
			data = ""
			break
		}
		data = C.GoStringN(&buf[0], C.int(length-1))

	default:
		return nil
	}

	if ret != C.CL_SUCCESS {
		return nil
	}
	d.properties[prop] = data
	return d.properties[prop]
}
