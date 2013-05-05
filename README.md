# GoPar: Automatic Loop Parallelization of Go Programs

Ian Wetherbee <ian.wetherbee@gmail.com>

- University of Illinois at Urbana-Champaign
- Senior Thesis, Computer Engineering
- Adviser: Wen-Mei Hwu, IMPACT Research Group

## Abstract

> Parallel computation hardware has achieved widespread consumer adoption, but software developers still need to manually exploit parallelism. Popular parallelization techniques require developers to make one or more program modifications, including source annotations, separate parallel kernel languages, manual data marshalling code, and framework-specific data containers. This additional burden forms a barrier to widespread adoption of parallel programming, and makes programs more verbose and difficult to analyze and modify. To solve these problems, this thesis introduces GoPar, an automatic loop-parallelizing compiler for the Go language that targets multicore CPUs. It aims to require no extra work from the developer to exploit parallelism and supports transforming many of Go's language features that enable compact and expressive code. GoPar is based on a new multi-pass compiler architecture containing analysis and transformation passes for detecting parallelizable loops and outputting transformed code. GoPar removes the developer barrier to exploiting parallel hardware without sacrificing maintainable code.

[Read full thesis](http://goo.gl/BKSPc)

# Goal

Recognize parallel `for` loops and automatically paralleize them. Supports `range` loops with slices or arrays as arguments. Slice values can be any type that does not contain pointers (so no interfaces, but structures are supported). Supports interprocedural analysis and imported packages (including the standard Go packages).

    a := make([]..., N)
    for i, v := range a {
        a[i] = ...
    }

Compiler output will explain why a loop could not be parallelized:

    Can't parallelize loop at /home/ian/gopar/src/nbody_shootout/nbody.go:71:2
    -> px.ReadWrite is accessed by all iterations

## Download

    git clone git://github.com/wetherbeei/gopar.git
    cd gopar
    GOPATH=`pwd`
    go install gopar
    PATH=$PATH:`pwd`/bin

## Usage

Replace `go install pkg` with `gopar install pkg` to build your project.

    GOPATH=<project root>
    gopar install project

## Tests

    cd gopar
    ./build.sh [nbody_shootout|stencil_bench]
    # run benchmarks on original and GOMAXPROCS={1,2,4,8}
    ./bench.sh [nbody_shootout|stencil_bench]

# Contributing/Bugs

See [GitHub issues](https://github.com/wetherbeei/gopar/issues), all contributions are welcome!
