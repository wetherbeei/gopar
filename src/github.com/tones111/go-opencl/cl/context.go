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

cl_context_properties PlatformToContextParameter(cl_platform_id platform) { return (cl_context_properties)platform; }

*/
import "C"

import (
	"io/ioutil"
	"runtime"
	"unsafe"
)

type ContextParameter C.cl_context_properties

const (
	CONTEXT_PLATFORM ContextParameter = C.CL_CONTEXT_PLATFORM
)

type ContextProperty C.cl_context_info

const (
	CONTEXT_REFERENCE_COUNT ContextProperty = C.CL_CONTEXT_REFERENCE_COUNT
	CONTEXT_NUM_DEVICES     ContextProperty = C.CL_CONTEXT_NUM_DEVICES
	CONTEXT_DEVICES         ContextProperty = C.CL_CONTEXT_DEVICES
	CONTEXT_PROPERTIES      ContextProperty = C.CL_CONTEXT_PROPERTIES
)

func ContextProperties() []ContextProperty {
	return []ContextProperty{
		CONTEXT_REFERENCE_COUNT,
		CONTEXT_NUM_DEVICES,
		CONTEXT_DEVICES,
		CONTEXT_PROPERTIES}
}

type Context struct {
	id         C.cl_context
	properties map[ContextProperty]interface{}
}

func createParameters(params map[ContextParameter]interface{}) ([]C.cl_context_properties, error) {
	c_params := make([]C.cl_context_properties, (len(params)<<1)+1)
	i := 0
	for param, value := range params {
		c_params[i] = C.cl_context_properties(param)

		switch param {
		case CONTEXT_PLATFORM:
			if v, ok := value.(Platform); ok {
				c_params[i+1] = C.PlatformToContextParameter(v.id)
			} else {
				return nil, Cl_error(C.CL_INVALID_VALUE)
			}

		default:
			return nil, Cl_error(C.CL_INVALID_VALUE)
		}
		i += 2
	}
	c_params[i] = 0

	return c_params, nil
}

func NewContextOfType(params map[ContextParameter]interface{}, t DeviceType) (*Context, error) {
	var c_params []C.cl_context_properties
	var propErr error
	if c_params, propErr = createParameters(params); propErr != nil {
		return nil, propErr
	}

	var c_context C.cl_context
	var err C.cl_int
	if c_context = C.clCreateContextFromType(&c_params[0], C.cl_device_type(t), nil, nil, &err); err != C.CL_SUCCESS {
		return nil, Cl_error(err)
	}
	c := &Context{id: c_context, properties: make(map[ContextProperty]interface{})}
	runtime.SetFinalizer(c, (*Context).release)

	return c, nil
}

func NewContextOfDevices(params map[ContextParameter]interface{}, devices []Device) (*Context, error) {
	var c_params []C.cl_context_properties
	var propErr error
	if c_params, propErr = createParameters(params); propErr != nil {
		return nil, propErr
	}

	c_devices := make([]C.cl_device_id, len(devices))
	for i, device := range devices {
		c_devices[i] = device.id
	}

	var c_context C.cl_context
	var err C.cl_int
	if c_context = C.clCreateContext(&c_params[0], C.cl_uint(len(c_devices)), &c_devices[0], nil, nil, &err); err != C.CL_SUCCESS {
		return nil, Cl_error(err)
	}
	c := &Context{id: c_context, properties: make(map[ContextProperty]interface{})}
	runtime.SetFinalizer(c, (*Context).release)

	return c, nil
}

func (c *Context) Property(prop ContextProperty) interface{} {
	if value, ok := c.properties[prop]; ok {
		return value
	}

	var data interface{}
	var length C.size_t
	var ret C.cl_int

	switch prop {
	case CONTEXT_REFERENCE_COUNT,
		CONTEXT_NUM_DEVICES:
		var val C.cl_uint
		ret = C.clGetContextInfo(c.id, C.cl_context_info(prop), C.size_t(unsafe.Sizeof(val)), unsafe.Pointer(&val), &length)
		data = val

	case CONTEXT_DEVICES:
		if data := c.Property(CONTEXT_NUM_DEVICES); data == nil {
			return nil
		} else {
			num_devs := data.(C.cl_uint)
			c_devs := make([]C.cl_device_id, num_devs)
			if ret = C.clGetContextInfo(c.id, C.cl_context_info(prop), C.size_t(num_devs*C.cl_uint(unsafe.Sizeof(c_devs[0]))), unsafe.Pointer(&c_devs[0]), &length); ret != C.CL_SUCCESS {
				return nil
			}
			devs := make([]Device, length/C.size_t(unsafe.Sizeof(c_devs[0])))
			for i, val := range c_devs {
				devs[i].id = val
			}
			data = devs
		}

	default:
		return nil
	}

	if ret != C.CL_SUCCESS {
		return nil
	}
	c.properties[prop] = data
	return c.properties[prop]
}

func (c *Context) NewCommandQueue(device Device, param CommandQueueParameter) (*CommandQueue, error) {
	var c_queue C.cl_command_queue
	var err C.cl_int
	if c_queue = C.clCreateCommandQueue(c.id, device.id, C.cl_command_queue_properties(param), &err); err != C.CL_SUCCESS {
		return nil, Cl_error(err)
	}
	queue := &CommandQueue{id: c_queue}
	runtime.SetFinalizer(queue, (*CommandQueue).release)

	return queue, nil
}

func (c *Context) NewProgramFromSource(prog []byte) (*Program, error) {
	var c_program C.cl_program
	var err C.cl_int

	srcPtr := (*C.char)(unsafe.Pointer(&prog[0]))
	length := (C.size_t)(len(prog))
	if c_program = C.clCreateProgramWithSource(c.id, 1, &srcPtr, &length, &err); err != C.CL_SUCCESS {
		return nil, Cl_error(err)
	}

	program := &Program{id: c_program}
	runtime.SetFinalizer(program, (*Program).release)

	return program, nil
}

func (c *Context) NewProgramFromFile(filename string) (*Program, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return c.NewProgramFromSource(content)
}

func (c *Context) NewBuffer(flags MemoryFlags, size uint32) (*Buffer, error) {
	var c_buffer C.cl_mem
	var err C.cl_int

	if c_buffer = C.clCreateBuffer(c.id, C.cl_mem_flags(flags), C.size_t(size), nil, &err); err != C.CL_SUCCESS {
		return nil, Cl_error(err)
	}

	buffer := &Buffer{id: c_buffer}
	runtime.SetFinalizer(buffer, (*Buffer).release)

	return buffer, nil
}

func (c *Context) NewImage2D(flags MemoryFlags, format ImageFormat, width, height, rowPitch uint32, data *byte) (*Image, error) {
	var c_buffer C.cl_mem
	var err C.cl_int

	c_format := &C.cl_image_format{
		image_channel_order:     C.cl_channel_order(format.ChannelOrder),
		image_channel_data_type: C.cl_channel_type(format.ChannelDataType)}

	if c_buffer = C.clCreateImage2D(c.id, C.cl_mem_flags(flags), c_format, C.size_t(width), C.size_t(height), C.size_t(rowPitch), unsafe.Pointer(data), &err); err != C.CL_SUCCESS {
		return nil, Cl_error(err)
	}

	image := &Image{id: c_buffer, format: format, properties: make(map[ImageProperty]Size)}
	runtime.SetFinalizer(image, (*Image).release)

	return image, nil
}

func (c *Context) NewImage3D(flags MemoryFlags, format ImageFormat, width, height, depth, rowPitch, slicePitch uint32, data *byte) (*Image, error) {
	var c_buffer C.cl_mem
	var err C.cl_int

	c_format := &C.cl_image_format{
		image_channel_order:     C.cl_channel_order(format.ChannelOrder),
		image_channel_data_type: C.cl_channel_type(format.ChannelDataType)}

	if c_buffer = C.clCreateImage3D(c.id, C.cl_mem_flags(flags), c_format, C.size_t(width), C.size_t(height), C.size_t(depth), C.size_t(rowPitch), C.size_t(slicePitch), unsafe.Pointer(data), &err); err != C.CL_SUCCESS {
		return nil, Cl_error(err)
	}

	image := &Image{id: c_buffer, format: format, properties: make(map[ImageProperty]Size)}
	runtime.SetFinalizer(image, (*Image).release)

	return image, nil
}

func (c *Context) NewSampler(normalizedCoords bool, addressingMode AddressingMode, filterMode FilterMode) (*Sampler, error) {
	var c_sampler C.cl_sampler
	var err C.cl_int

	cNormalizedCoords := C.cl_bool(C.CL_FALSE)
	if normalizedCoords {
		cNormalizedCoords = C.CL_TRUE
	}

	if c_sampler = C.clCreateSampler(c.id, cNormalizedCoords, C.cl_addressing_mode(addressingMode), C.cl_filter_mode(filterMode), &err); err != C.CL_SUCCESS {
		return nil, Cl_error(err)
	}

	sampler := &Sampler{id: c_sampler, properties: make(map[SamplerProperty]interface{})}
	runtime.SetFinalizer(sampler, (*Sampler).release)

	return sampler, nil
}

func (c *Context) release() error {
	if c.id != nil {
		if err := C.clReleaseContext(c.id); err != C.CL_SUCCESS {
			return Cl_error(err)
		}
		c.id = nil
	}
	return nil
}
