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
	"github.com/tones111/raw"
	"testing"
)

func Test_VectAdd(t *testing.T) {
	const elements = 100000

	A := make([]int, elements)
	B := make([]int, elements)
	for i := 0; i < elements; i++ {
		A[i], B[i] = i, i
	}
	bytes := uint32(len(raw.ByteSlice(A)))

	for _, platform := range Platforms {
		for _, dev := range platform.Devices {

			var err error
			var context *Context
			var queue *CommandQueue
			var bufA, bufB, bufC *Buffer
			var program *Program
			var kernel *Kernel

			if context, err = NewContextOfDevices(nil, []Device{dev}); err != nil {
				t.Fatal(err)
			}

			if queue, err = context.NewCommandQueue(dev, QUEUE_NIL); err != nil {
				t.Fatal(err)
			}

			if bufA, err = context.NewBuffer(MEM_READ_ONLY, bytes); err != nil {
				t.Fatal(err)
			}
			if bufB, err = context.NewBuffer(MEM_READ_ONLY, bytes); err != nil {
				t.Fatal(err)
			}
			if bufC, err = context.NewBuffer(MEM_WRITE_ONLY, bytes); err != nil {
				t.Fatal(err)
			}

			if err = queue.EnqueueWriteBuffer(bufA, raw.ByteSlice(A), 0); err != nil {
				t.Fatal(err)
			}
			if err = queue.EnqueueWriteBuffer(bufB, raw.ByteSlice(B), 0); err != nil {
				t.Fatal(err)
			}

			if program, err = context.NewProgramFromFile("vector.cl"); err != nil {
				t.Fatal(err)
			}

			if err = program.Build([]Device{dev}, ""); err != nil {
				t.Fatal(err)
			}

			if kernel, err = program.NewKernelNamed("vectAddInt"); err != nil {
				t.Fatal(err)
			}

			if err = kernel.SetArgs(0, []interface{}{bufA, bufB, bufC}); err != nil {
				t.Fatal(err)
			}

			if err = queue.EnqueueKernel(kernel, []Size{0}, []Size{elements}, []Size{1}); err != nil {
				t.Fatal(err)
			}

			if outBuf, err := queue.EnqueueReadBuffer(bufC, 0, bytes); err != nil {
				t.Fatal(err)
			} else {
				C := raw.IntSlice(outBuf)

				for i := 0; i < elements; i++ {
					if C[i] != i<<1 {
						t.Fatal("Output is incorrect")
					}
				}
			}
		}
	}
}
