//go:build windows

package main

import "unsafe"

func uintptrOf(p *[2]int32) uintptr { return uintptr(unsafe.Pointer(p)) }

func ptrOf(p uintptr) unsafe.Pointer { return unsafe.Pointer(p) }
