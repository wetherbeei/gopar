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

import (
	"testing"
)

func getPlatform(t *testing.T) Platform {
	if len(Platforms) == 0 {
		t.Fatal("No platforms found")
	}
	return Platforms[0]
}

func getCPUDevice(p Platform, t *testing.T) []Device {
	if len(p.Devices) == 0 {
		t.Fatal("No devices found")
	}
	return p.Devices[:1]
}

func getContext(p Platform, d []Device, t *testing.T) *Context {
	if context, err := NewContextOfDevices(map[ContextParameter]interface{}{CONTEXT_PLATFORM: p}, d); err != nil {
		t.Fatal("Error creating context:", err)
	} else {
		return context
	}
	return nil
}

func getProgram(c *Context, filename string, t *testing.T) *Program {
	if program, err := c.NewProgramFromFile(filename); err != nil {
		t.Fatal("Error creating program:", err)
	} else {
		return program
	}
	return nil
}

func getKernel(p *Program, name string, t *testing.T) *Kernel {
	if kernel, err := p.NewKernelNamed(name); err != nil {
		t.Fatal("Error creating kernel:", err)
	} else {
		return kernel
	}
	return nil
}

func getQueue(c *Context, d Device, t *testing.T) *CommandQueue {
	if queue, err := c.NewCommandQueue(d, QUEUE_NIL); err != nil {
		t.Fatal("Error creating command queue:", err)
	} else {
		return queue
	}
	return nil
}

func Test_OpenCl(t *testing.T) {
	inData := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	var err error

	platform := getPlatform(t)
	devices := getCPUDevice(platform, t)
	context := getContext(platform, devices, t)
	queue := getQueue(context, devices[0], t)

	program := getProgram(context, "vector.cl", t)
	if err = program.Build(nil, ""); err != nil {
		t.Fatal("Error building program:", err)
	}
	kernel := getKernel(program, "vectSquareUChar", t)

	var inBuf, outBuf *Buffer
	if inBuf, err = context.NewBuffer(MEM_READ_ONLY, 100); err != nil {
		t.Fatal("Error creating in buffer:", err)
	} else if outBuf, err = context.NewBuffer(MEM_WRITE_ONLY, 100); err != nil {
		t.Fatal("Error creating out buffer:", err)
	}

	if err = queue.EnqueueWriteBuffer(inBuf, inData, 0); err != nil {
		t.Fatal("Error enquing data:", err)
	}

	if err = kernel.SetArgs(0, []interface{}{inBuf, outBuf}); err != nil {
		t.Fatal("Error setting kernel arguments :", err)
	} else if err = queue.EnqueueKernel(kernel, []Size{0}, []Size{Size(len(inData))}, []Size{Size(len(inData))}); err != nil {
		t.Fatal("Error enquing kernel:", err)
	}

	var data []byte
	if data, err = queue.EnqueueReadBuffer(outBuf, 0, uint32(len(inData))); err != nil {
		t.Fatal("Error reading data:", err)
	}

	for i, v := range data {
		if v != inData[i]*inData[i] {
			t.Fatal("Incorrect results")
		}
	}
}
