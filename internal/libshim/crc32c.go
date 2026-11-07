package libshim

import (
	"hash/crc32"
	"unsafe"

	"modernc.org/libc"
)

// castagnoli is the CRC32C table; hash/crc32 uses the SSE4.2 or ARMv8
// CRC instructions with it where they exist.
var castagnoli = crc32.MakeTable(crc32.Castagnoli)

// Xghostty_ccgo_crc32c continues a CRC32C over n bytes at p. The patch
// internal/patches/03-ccgo-crc32c.patch makes ghostty's crc32c module call
// it when Zig's C backend compiles the library, in place of the inline
// CRC instructions (not expressible in C) and the table-driven software
// fallback, which took half of the time of encoding a snapshot.
//
// crc is ghostty's running state, kept inverted (it starts at 0xFFFFFFFF
// and is inverted once more at the end), while hash/crc32 works on the
// non-inverted value, so the state is inverted around the call.
func Xghostty_ccgo_crc32c(tls *libc.TLS, crc uint32, p uintptr, n uint64) uint32 {
	if n == 0 {
		return crc
	}
	return ^crc32.Update(^crc, castagnoli, unsafe.Slice((*byte)(unsafe.Pointer(p)), n))
}
