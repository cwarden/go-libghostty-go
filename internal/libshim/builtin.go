package libshim

import (
	"math"
	"unsafe"

	"modernc.org/libc"
)

// GCC's __builtin_{add,sub,mul}_overflow(a, b, &res) store a op b in *res
// and return whether the result wrapped. ccgo names them by the result
// type; modernc.org/libc provides only some of the widths, and the rest
// are here.

type signed interface {
	~int8 | ~int16 | ~int32 | ~int64
}

type unsigned interface {
	~uint8 | ~uint16 | ~uint32 | ~uint64
}

func store[T any](res uintptr, v T) { *(*T)(unsafe.Pointer(res)) = v }

func b2i(b bool) int32 {
	if b {
		return 1
	}
	return 0
}

func addS[T signed](a, b T, res uintptr) int32 {
	r := a + b
	store(res, r)
	return b2i((a > 0 && b > 0 && r < 0) || (a < 0 && b < 0 && r >= 0))
}

func subS[T signed](a, b T, res uintptr) int32 {
	r := a - b
	store(res, r)
	return b2i((a >= 0 && b < 0 && r < 0) || (a < 0 && b > 0 && r >= 0))
}

// isMin reports whether x is the most negative value of its type, the one
// value whose negation overflows.
func isMin[T signed](x T) bool { return x < 0 && x == -x }

func mulS[T signed](a, b T, res uintptr) int32 {
	r := a * b
	store(res, r)
	if a == 0 || b == 0 {
		return 0
	}
	return b2i(r/b != a || (a == -1 && isMin(b)) || (b == -1 && isMin(a)))
}

func addU[T unsigned](a, b T, res uintptr) int32 {
	r := a + b
	store(res, r)
	return b2i(r < a)
}

func subU[T unsigned](a, b T, res uintptr) int32 {
	store(res, a-b)
	return b2i(a < b)
}

func mulU[T unsigned](a, b T, res uintptr) int32 {
	r := a * b
	store(res, r)
	return b2i(a != 0 && r/a != b)
}

// The widths modernc.org/libc provides (Int64, Uint32 and Uint64 for add,
// Int64 for sub, Int64 and Uint64 for mul) are not repeated here.

func X__builtin_add_overflowInt8(t *libc.TLS, a, b int8, res uintptr) int32   { return addS(a, b, res) }
func X__builtin_add_overflowInt16(t *libc.TLS, a, b int16, res uintptr) int32 { return addS(a, b, res) }
func X__builtin_add_overflowInt32(t *libc.TLS, a, b int32, res uintptr) int32 { return addS(a, b, res) }
func X__builtin_sub_overflowInt8(t *libc.TLS, a, b int8, res uintptr) int32   { return subS(a, b, res) }
func X__builtin_sub_overflowInt16(t *libc.TLS, a, b int16, res uintptr) int32 { return subS(a, b, res) }
func X__builtin_sub_overflowInt32(t *libc.TLS, a, b int32, res uintptr) int32 { return subS(a, b, res) }
func X__builtin_mul_overflowInt8(t *libc.TLS, a, b int8, res uintptr) int32   { return mulS(a, b, res) }
func X__builtin_mul_overflowInt16(t *libc.TLS, a, b int16, res uintptr) int32 { return mulS(a, b, res) }
func X__builtin_mul_overflowInt32(t *libc.TLS, a, b int32, res uintptr) int32 { return mulS(a, b, res) }
func X__builtin_add_overflowUint8(t *libc.TLS, a, b uint8, res uintptr) int32 { return addU(a, b, res) }
func X__builtin_add_overflowUint16(t *libc.TLS, a, b uint16, res uintptr) int32 {
	return addU(a, b, res)
}
func X__builtin_sub_overflowUint8(t *libc.TLS, a, b uint8, res uintptr) int32 { return subU(a, b, res) }
func X__builtin_sub_overflowUint16(t *libc.TLS, a, b uint16, res uintptr) int32 {
	return subU(a, b, res)
}
func X__builtin_sub_overflowUint32(t *libc.TLS, a, b uint32, res uintptr) int32 {
	return subU(a, b, res)
}
func X__builtin_sub_overflowUint64(t *libc.TLS, a, b uint64, res uintptr) int32 {
	return subU(a, b, res)
}
func X__builtin_mul_overflowUint8(t *libc.TLS, a, b uint8, res uintptr) int32 { return mulU(a, b, res) }
func X__builtin_mul_overflowUint16(t *libc.TLS, a, b uint16, res uintptr) int32 {
	return mulU(a, b, res)
}
func X__builtin_mul_overflowUint32(t *libc.TLS, a, b uint32, res uintptr) int32 {
	return mulU(a, b, res)
}

// X__builtin_fmaxf is fmaxf: the larger of a and b, ignoring a NaN operand.
func X__builtin_fmaxf(t *libc.TLS, a, b float32) float32 {
	return float32(math.Max(float64(a), float64(b)))
}

// X__builtin_fminf is fminf: the smaller of a and b, ignoring a NaN
// operand.
func X__builtin_fminf(t *libc.TLS, a, b float32) float32 {
	return float32(math.Min(float64(a), float64(b)))
}
