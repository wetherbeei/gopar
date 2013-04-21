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

type channelOrder C.cl_channel_order

const (
	R         channelOrder = C.CL_R
	A         channelOrder = C.CL_A
	RG        channelOrder = C.CL_RG
	RA        channelOrder = C.CL_RA
	RGB       channelOrder = C.CL_RGB
	RGBA      channelOrder = C.CL_RGBA
	BGRA      channelOrder = C.CL_BGRA
	ARGB      channelOrder = C.CL_ARGB
	INTENSITY channelOrder = C.CL_INTENSITY
	LUMINANCE channelOrder = C.CL_LUMINANCE
	Rx        channelOrder = C.CL_Rx
	RGx       channelOrder = C.CL_RGx
	RGBx      channelOrder = C.CL_RGBx
)

type channelType C.cl_channel_type

const (
	FLOAT            channelType = C.CL_FLOAT
	HALF_FLOAT       channelType = C.CL_HALF_FLOAT
	SIGNED_INT8      channelType = C.CL_SIGNED_INT8
	SIGNED_INT16     channelType = C.CL_SIGNED_INT16
	SIGNED_INT32     channelType = C.CL_SIGNED_INT32
	SNORM_INT8       channelType = C.CL_SNORM_INT8
	SNORM_INT16      channelType = C.CL_SNORM_INT16
	UNORM_INT8       channelType = C.CL_UNORM_INT8
	UNORM_INT16      channelType = C.CL_UNORM_INT16
	UNORM_SHORT_565  channelType = C.CL_UNORM_SHORT_565
	UNORM_SHORT_555  channelType = C.CL_UNORM_SHORT_555
	UNORM_INT_101010 channelType = C.CL_UNORM_INT_101010
	UNSIGNED_INT8    channelType = C.CL_UNSIGNED_INT8
	UNSIGNED_INT16   channelType = C.CL_UNSIGNED_INT16
	UNSIGNED_INT32   channelType = C.CL_UNSIGNED_INT32
)

type ImageFormat struct {
	ChannelOrder    channelOrder
	ChannelDataType channelType
}

type ImageProperty C.cl_image_info

const (
	//IMAGE_FORMAT       ImageProperty = C.CL_IMAGE_FORMAT
	IMAGE_ELEMENT_SIZE ImageProperty = C.CL_IMAGE_ELEMENT_SIZE
	IMAGE_ROW_PITCH    ImageProperty = C.CL_IMAGE_ROW_PITCH
	IMAGE_SLICE_PITCH  ImageProperty = C.CL_IMAGE_SLICE_PITCH
	IMAGE_WIDTH        ImageProperty = C.CL_IMAGE_WIDTH
	IMAGE_HEIGHT       ImageProperty = C.CL_IMAGE_HEIGHT
	IMAGE_DEPTH        ImageProperty = C.CL_IMAGE_DEPTH
)

type Image struct {
	id         C.cl_mem
	format     ImageFormat
	properties map[ImageProperty]Size
}

func (i *Image) Format() ImageFormat {
	return i.format
}

func (i *Image) Property(prop ImageProperty) Size {
	if value, ok := i.properties[prop]; ok {
		return value
	}

	var data C.size_t
	var length C.size_t

	if ret := C.clGetImageInfo(i.id, C.cl_image_info(prop), C.size_t(unsafe.Sizeof(data)), unsafe.Pointer(&data), &length); ret != C.CL_SUCCESS {
		return 0
	}

	i.properties[prop] = Size(data)
	return i.properties[prop]
}

func (im *Image) release() error {
	err := releaseMemObject(im.id)
	im.id = nil
	return err
}
