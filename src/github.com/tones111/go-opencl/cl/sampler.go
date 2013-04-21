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
#cgo CFLAGS: -I .
#cgo LDFLAGS: -lOpenCL

#include "CL/opencl.h"

*/
import "C"

import (
	"unsafe"
)

type AddressingMode C.cl_addressing_mode

const (
	ADDRESS_NONE            AddressingMode = C.CL_ADDRESS_NONE
	ADDRESS_CLAMP_TO_EDGE   AddressingMode = C.CL_ADDRESS_CLAMP_TO_EDGE
	ADDRESS_CLAMP           AddressingMode = C.CL_ADDRESS_CLAMP
	ADDRESS_REPEAT          AddressingMode = C.CL_ADDRESS_REPEAT
	ADDRESS_MIRRORED_REPEAT AddressingMode = C.CL_ADDRESS_MIRRORED_REPEAT
)

type FilterMode C.cl_filter_mode

const (
	FILTER_NEAREST FilterMode = C.CL_FILTER_NEAREST
	FILTER_LINEAR  FilterMode = C.CL_FILTER_LINEAR
)

type SamplerProperty C.cl_sampler_info

const (
	SAMPLER_REFERENCE_COUNT SamplerProperty = C.CL_SAMPLER_REFERENCE_COUNT
	//SAMPLER_CONTEXT           SamplerProperty = C.CL_SAMPLER_CONTEXT
	SAMPLER_NORMALIZED_COORDS SamplerProperty = C.CL_SAMPLER_NORMALIZED_COORDS
	SAMPLER_ADDRESSING_MODE   SamplerProperty = C.CL_SAMPLER_ADDRESSING_MODE
	SAMPLER_FILTER_MODE       SamplerProperty = C.CL_SAMPLER_FILTER_MODE
)

type Sampler struct {
	id         C.cl_sampler
	properties map[SamplerProperty]interface{}
}

func (s *Sampler) Property(prop SamplerProperty) interface{} {
	if value, ok := s.properties[prop]; ok {
		return value
	}

	var data interface{}
	var length C.size_t
	var ret C.cl_int

	switch prop {
	case SAMPLER_REFERENCE_COUNT:
		var val C.cl_uint
		ret = C.clGetSamplerInfo(s.id, C.cl_sampler_info(prop), C.size_t(unsafe.Sizeof(val)), unsafe.Pointer(&val), &length)
		data = val
	case SAMPLER_NORMALIZED_COORDS:
		var val C.cl_bool
		ret = C.clGetSamplerInfo(s.id, C.cl_sampler_info(prop), C.size_t(unsafe.Sizeof(val)), unsafe.Pointer(&val), &length)
		data = val == C.CL_TRUE
	case SAMPLER_ADDRESSING_MODE:
		var val C.cl_addressing_mode
		ret = C.clGetSamplerInfo(s.id, C.cl_sampler_info(prop), C.size_t(unsafe.Sizeof(val)), unsafe.Pointer(&val), &length)
		data = AddressingMode(ret)
	case SAMPLER_FILTER_MODE:
		var val C.cl_filter_mode
		ret = C.clGetSamplerInfo(s.id, C.cl_sampler_info(prop), C.size_t(unsafe.Sizeof(val)), unsafe.Pointer(&val), &length)
		data = FilterMode(ret)
	default:
		return nil
	}

	if ret != C.CL_SUCCESS {
		return nil
	}
	s.properties[prop] = data
	return s.properties[prop]
}

func (s *Sampler) release() error {
	if s.id != nil {
		if err := C.clReleaseSampler(C.cl_sampler(s.id)); err != C.CL_SUCCESS {
			return Cl_error(err)
		}
	}
	return nil
}
