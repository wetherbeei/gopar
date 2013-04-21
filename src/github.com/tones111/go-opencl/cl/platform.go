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

type PlatformProperty C.cl_platform_info

const (
	PLATFORM_PROFILE    PlatformProperty = C.CL_PLATFORM_PROFILE
	PLATFORM_VERSION    PlatformProperty = C.CL_PLATFORM_VERSION
	PLATFORM_NAME       PlatformProperty = C.CL_PLATFORM_NAME
	PLATFORM_VENDOR     PlatformProperty = C.CL_PLATFORM_VENDOR
	PLATFORM_EXTENSIONS PlatformProperty = C.CL_PLATFORM_EXTENSIONS
)

type Platform struct {
	id         C.cl_platform_id
	Devices    []Device
	properties map[PlatformProperty]string
}

func GetPlatforms() []Platform {
	var count C.cl_uint
	if ret := C.clGetPlatformIDs(0, (*C.cl_platform_id)(nil), &count); ret != C.CL_SUCCESS || count == 0 {
		return nil
	}

	c_platforms := make([]C.cl_platform_id, count)
	if ret := C.clGetPlatformIDs(count, &c_platforms[0], &count); ret != C.CL_SUCCESS || count == 0 {
		return nil
	}
	platforms := make([]Platform, 0, count)

	for _, pid := range c_platforms {
		if ret := C.clGetDeviceIDs(pid, C.cl_device_type(DEVICE_TYPE_ALL), 0, (*C.cl_device_id)(nil), &count); ret != C.CL_SUCCESS || count == 0 {
			continue
		}

		c_devices := make([]C.cl_device_id, count)
		if ret := C.clGetDeviceIDs(pid, C.cl_device_type(DEVICE_TYPE_ALL), count, &c_devices[0], &count); ret != C.CL_SUCCESS || count == 0 {
			continue
		}

		platform := Platform{
			id:         pid,
			Devices:    make([]Device, count),
			properties: make(map[PlatformProperty]string),
		}
		for i, did := range c_devices {
			platform.Devices[i].id = did
			platform.Devices[i].properties = make(map[DeviceProperty]interface{})
		}
		platforms = append(platforms, platform)
	}
	return platforms
}

func (p *Platform) Property(prop PlatformProperty) string {
	if value, ok := p.properties[prop]; ok {
		return value
	}

	var count C.size_t
	if ret := C.clGetPlatformInfo(p.id, C.cl_platform_info(prop), 0, nil, &count); ret != C.CL_SUCCESS || count < 1 {
		return ""
	}

	buf := make([]C.char, count)
	if ret := C.clGetPlatformInfo(p.id, C.cl_platform_info(prop), count, unsafe.Pointer(&buf[0]), &count); ret != C.CL_SUCCESS || count < 1 {
		return ""
	}
	p.properties[prop] = C.GoStringN(&buf[0], C.int(count-1))
	return p.properties[prop]
}
