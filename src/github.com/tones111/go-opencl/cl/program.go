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
	"runtime"
	"unsafe"
)

type BuildProperty C.cl_program_build_info

const (
	//BUILD_STATUS  BuildProperty = C.CL_PROGRAM_BUILD_STATUS
	BUILD_OPTIONS BuildProperty = C.CL_PROGRAM_BUILD_OPTIONS
	BUILD_LOG     BuildProperty = C.CL_PROGRAM_BUILD_LOG
)

type BuildStatus C.cl_build_status

const (
	BUILD_NONE        BuildStatus = C.CL_BUILD_NONE
	BUILD_ERROR       BuildStatus = C.CL_BUILD_ERROR
	BUILD_SUCCESS     BuildStatus = C.CL_BUILD_SUCCESS
	BUILD_IN_PROGRESS BuildStatus = C.CL_BUILD_IN_PROGRESS
)

func (status BuildStatus) String() string {
	switch status {
	case BUILD_NONE:
		return "None"
	case BUILD_ERROR:
		return "Error"
	case BUILD_SUCCESS:
		return "Success"
	case BUILD_IN_PROGRESS:
		return "In Progress"
	}
	return "Unknown"
}

type Program struct {
	id C.cl_program
}

func (p *Program) Build(devices []Device, options string) error {
	cs := C.CString(options)
	defer C.free(unsafe.Pointer(cs))

	if len(devices) < 1 {
		if ret := C.clBuildProgram(p.id, 0, nil, cs, nil, nil); ret != C.CL_SUCCESS {
			return Cl_error(ret)
		}
		return nil
	}

	c_devices := make([]C.cl_device_id, len(devices))
	for i, device := range devices {
		c_devices[i] = device.id
	}
	if ret := C.clBuildProgram(p.id, C.cl_uint(len(c_devices)), &c_devices[0], cs, nil, nil); ret != C.CL_SUCCESS {
		return Cl_error(ret)
	}
	return nil
}

func (p *Program) NewKernelNamed(name string) (*Kernel, error) {
	var c_kernel C.cl_kernel
	var err C.cl_int

	cs := C.CString(name)
	defer C.free(unsafe.Pointer(cs))

	if c_kernel = C.clCreateKernel(p.id, cs, &err); err != C.CL_SUCCESS {
		return nil, Cl_error(err)
	}

	kernel := &Kernel{id: c_kernel}
	runtime.SetFinalizer(kernel, (*Kernel).release)

	return kernel, nil
}

func (p *Program) BuildStatus(dev Device) BuildStatus {
	var c_status C.cl_build_status
	var count C.size_t
	if ret := C.clGetProgramBuildInfo(p.id, dev.id, C.CL_PROGRAM_BUILD_STATUS, C.size_t(unsafe.Sizeof(c_status)), unsafe.Pointer(&c_status), &count); ret != C.CL_SUCCESS {
		return BUILD_ERROR
	}
	return BuildStatus(c_status)

}

func (p *Program) Property(dev Device, prop BuildProperty) string {
	var count C.size_t
	if ret := C.clGetProgramBuildInfo(p.id, dev.id, C.cl_program_build_info(prop), 0, nil, &count); ret != C.CL_SUCCESS || count < 1 {
		return ""
	}

	buf := make([]C.char, count)
	if ret := C.clGetProgramBuildInfo(p.id, dev.id, C.cl_program_build_info(prop), count, unsafe.Pointer(&buf[0]), &count); ret != C.CL_SUCCESS || count < 1 {
		return ""
	}
	return C.GoStringN(&buf[0], C.int(count-1))
}

func (p *Program) release() error {
	if p.id != nil {
		if err := C.clReleaseProgram(p.id); err != C.CL_SUCCESS {
			return Cl_error(err)
		}
		p.id = nil
	}
	return nil
}
