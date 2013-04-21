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

type CommandQueueParameter C.cl_command_queue_properties

const (
	QUEUE_OUT_OF_ORDER_EXEC_MODE_ENABLE CommandQueueParameter = C.CL_QUEUE_OUT_OF_ORDER_EXEC_MODE_ENABLE
	QUEUE_PROFILING_ENABLE              CommandQueueParameter = C.CL_QUEUE_PROFILING_ENABLE
	QUEUE_NIL                           CommandQueueParameter = 0
)

type CommandQueue struct {
	id C.cl_command_queue
}

func (cq *CommandQueue) EnqueueKernel(k *Kernel, offset, gsize, lsize []Size) error {

	cptr := func(w []Size) *C.size_t {
		if len(w) == 0 {
			return nil
		}
		return (*C.size_t)(unsafe.Pointer(&w[0]))
	}

	c_offset := cptr(offset)
	c_gsize := cptr(gsize)
	c_lsize := cptr(lsize)

	if ret := C.clEnqueueNDRangeKernel(cq.id, k.id, C.cl_uint(len(gsize)), c_offset, c_gsize, c_lsize, 0, nil, nil); ret != C.CL_SUCCESS {
		return Cl_error(ret)
	}
	return nil
}

func (cq *CommandQueue) EnqueueReadBuffer(buf *Buffer, offset uint32, size uint32) ([]byte, error) {
	bytes := make([]byte, size)
	if ret := C.clEnqueueReadBuffer(cq.id, buf.id, C.CL_TRUE, C.size_t(offset), C.size_t(size), unsafe.Pointer(&bytes[0]), 0, nil, nil); ret != C.CL_SUCCESS {
		return nil, Cl_error(ret)
	}
	return bytes, nil
}

func (cq *CommandQueue) EnqueueWriteBuffer(buf *Buffer, data []byte, offset uint32) error {

	if ret := C.clEnqueueWriteBuffer(cq.id, buf.id, C.CL_TRUE, C.size_t(offset), C.size_t(len(data)), unsafe.Pointer(&data[0]), 0, nil, nil); ret != C.CL_SUCCESS {
		return Cl_error(ret)
	}
	return nil
}

func (cq *CommandQueue) EnqueueReadImage(i *Image, blocking bool, origin, region [3]Size, rowPitch, slicePitch Size) ([]byte, error) {
	c_blocking := C.cl_bool(C.CL_FALSE)
	if blocking {
		c_blocking = C.CL_TRUE
	}

	size := Size(0)
	if i.Property(IMAGE_DEPTH) == 0 || i.Property(IMAGE_DEPTH) == 1 { // 2D image
		if rowPitch == 0 {
			rowPitch = i.Property(IMAGE_WIDTH) * i.Property(IMAGE_ELEMENT_SIZE)
		}
		size = rowPitch * i.Property(IMAGE_HEIGHT)
	} else {
		if slicePitch == 0 {
			if rowPitch == 0 {
				rowPitch = i.Property(IMAGE_WIDTH) * i.Property(IMAGE_ELEMENT_SIZE)
			}
			slicePitch = rowPitch * i.Property(IMAGE_DEPTH)
		}
		size = slicePitch * i.Property(IMAGE_DEPTH)
	}
	bytes := make([]byte, size)

	if ret := C.clEnqueueReadImage(
		cq.id, i.id, c_blocking,
		(*C.size_t)(unsafe.Pointer(&origin[0])), (*C.size_t)(unsafe.Pointer(&region[0])),
		C.size_t(rowPitch), C.size_t(slicePitch), unsafe.Pointer(&bytes[0]),
		0, nil, nil); ret != C.CL_SUCCESS {
		return nil, Cl_error(ret)
	}
	return bytes, nil
}

func (cq *CommandQueue) EnqueueWriteImage(img *Image, blocking bool, origin, region [3]Size, rowPitch, slicePitch Size, data []byte) error {
	c_blocking := C.cl_bool(C.CL_FALSE)
	if blocking {
		c_blocking = C.CL_TRUE
	}

	if ret := C.clEnqueueWriteImage(
		cq.id, img.id, c_blocking,
		(*C.size_t)(unsafe.Pointer(&origin[0])), (*C.size_t)(unsafe.Pointer(&region[0])),
		C.size_t(rowPitch), C.size_t(slicePitch),
		unsafe.Pointer(&data[0]),
		C.cl_uint(0), nil, nil); ret != C.CL_SUCCESS {
		return Cl_error(ret)
	}
	return nil
}

func (q *CommandQueue) release() error {
	if q.id != nil {
		if err := C.clReleaseCommandQueue(q.id); err != C.CL_SUCCESS {
			return Cl_error(err)
		}
		q.id = nil
	}
	return nil
}
