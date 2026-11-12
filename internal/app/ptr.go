package app

import "unsafe"

func uintptrOf(p *int32) uintptr  { return uintptr(unsafe.Pointer(p)) }
func uintptrOf8(p *uint8) uintptr { return uintptr(unsafe.Pointer(p)) }
