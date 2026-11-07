// Package libshim implements the C functions that the translated
// libghostty-vt code calls and modernc.org/libc does not provide. ccgo
// links it as -lshim: every exported function whose first parameter is
// *libc.TLS is visible to the generated code as the C function of the
// same name without the X prefix.
package libshim

import (
	"sync/atomic"
	"unsafe"

	"modernc.org/libc"
)

// X__atomic_loadUintptr is GCC's __atomic_load for a pointer-sized
// object: *ret = *ptr. libc provides the fixed-width variants only.
func X__atomic_loadUintptr(t *libc.TLS, ptr, ret uintptr, _ int32) {
	*(*uintptr)(unsafe.Pointer(ret)) = atomic.LoadUintptr((*uintptr)(unsafe.Pointer(ptr)))
}

// X__atomic_exchangeUintptr is GCC's __atomic_exchange for a pointer-sized
// object: *ret = *ptr; *ptr = *val.
func X__atomic_exchangeUintptr(t *libc.TLS, ptr, val, ret uintptr, _ int32) {
	*(*uintptr)(unsafe.Pointer(ret)) = atomic.SwapUintptr((*uintptr)(unsafe.Pointer(ptr)), *(*uintptr)(unsafe.Pointer(val)))
}

// X__atomic_compare_exchangeUintptr is GCC's __atomic_compare_exchange for
// a pointer-sized object. It returns 1 and stores *desired when *ptr equals
// *expected, and otherwise returns 0 after copying *ptr to *expected.
func X__atomic_compare_exchangeUintptr(t *libc.TLS, ptr, expected, desired uintptr, weak, success, failure int32) int32 {
	p := (*uintptr)(unsafe.Pointer(ptr))
	exp := (*uintptr)(unsafe.Pointer(expected))
	des := *(*uintptr)(unsafe.Pointer(desired))
	for {
		old := *exp
		if atomic.CompareAndSwapUintptr(p, old, des) {
			return 1
		}
		if cur := atomic.LoadUintptr(p); cur != old {
			*exp = cur
			return 0
		}
	}
}
