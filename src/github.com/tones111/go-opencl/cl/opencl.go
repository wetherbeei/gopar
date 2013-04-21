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
	"fmt"
)

var Platforms []Platform

func init() {
	Platforms = GetPlatforms()
}

type Size C.size_t

type Cl_error C.cl_int

func (e Cl_error) Error() string {
	mesg := errMesg[e]
	if mesg == "" {
		return fmt.Sprintf("error %d", int(e))
	}
	return mesg
}

var errMesg = map[Cl_error]string{
	C.CL_SUCCESS: "Success",

	C.CL_BUILD_PROGRAM_FAILURE:                     "Build Program Failure",
	C.CL_COMPILER_NOT_AVAILABLE:                    "Compiler not Available",
	C.CL_DEVICE_NOT_AVAILABLE:                      "Device not Available",
	C.CL_DEVICE_NOT_FOUND:                          "Device not Found",
	C.CL_EXEC_STATUS_ERROR_FOR_EVENTS_IN_WAIT_LIST: "Wait List Event Error",
	C.CL_IMAGE_FORMAT_NOT_SUPPORTED:                "Unsupported Image Format",
	C.CL_INVALID_BINARY:                            "Invalid Binary",
	C.CL_INVALID_BUFFER_SIZE:                       "Invalid Buffer Size",
	C.CL_INVALID_BUILD_OPTIONS:                     "Invalid Build Options",
	C.CL_INVALID_COMMAND_QUEUE:                     "Invalid Command Queue",
	C.CL_INVALID_CONTEXT:                           "Invalid Context",
	C.CL_INVALID_DEVICE:                            "Invalid Device",
	C.CL_INVALID_DEVICE_TYPE:                       "Invalid Device Type",
	C.CL_INVALID_EVENT_WAIT_LIST:                   "Invalid Event Wait List",
	C.CL_INVALID_GLOBAL_OFFSET:                     "Invalid Global Offset",
	C.CL_INVALID_GLOBAL_WORK_SIZE:                  "Invalid Global Work Size",
	C.CL_INVALID_HOST_PTR:                          "Invalid Host Pointer",
	C.CL_INVALID_IMAGE_FORMAT_DESCRIPTOR:           "Invalid Image Format",
	C.CL_INVALID_IMAGE_SIZE:                        "Invalid Image Size",
	C.CL_INVALID_KERNEL:                            "Invalid Kernel",
	C.CL_INVALID_KERNEL_ARGS:                       "Invalid Kernel Arguments",
	C.CL_INVALID_KERNEL_DEFINITION:                 "Invalid Kernel Definition",
	C.CL_INVALID_KERNEL_NAME:                       "Invalid Kernel Name",
	C.CL_INVALID_MEM_OBJECT:                        "Invalid Memory Object",
	C.CL_INVALID_OPERATION:                         "Invalid Operation",
	C.CL_INVALID_PLATFORM:                          "Invalid Platform",
	C.CL_INVALID_PROPERTY:                          "Invalid Property",
	C.CL_INVALID_PROGRAM:                           "Invalid Program",
	C.CL_INVALID_PROGRAM_EXECUTABLE:                "Invalid Program Executable",
	C.CL_INVALID_SAMPLER:                           "Invalid Sampler",
	C.CL_INVALID_VALUE:                             "Invalid Value",
	C.CL_INVALID_WORK_DIMENSION:                    "Invalid Work Dimension",
	C.CL_INVALID_WORK_GROUP_SIZE:                   "Invalid Work Group Size",
	C.CL_INVALID_WORK_ITEM_SIZE:                    "Invalid Work Item Size",
	C.CL_MEM_OBJECT_ALLOCATION_FAILURE:             "Memory Object Allocation Failure",
	C.CL_MISALIGNED_SUB_BUFFER_OFFSET:              "Misaligned Sub-Buffer Offset",
	C.CL_OUT_OF_HOST_MEMORY:                        "Out of Host Memory",
	C.CL_OUT_OF_RESOURCES:                          "Out of Resources",
}
