package libghostty

// System-level configuration wrapping ghostty_sys_set().
// These are process-global settings that must be configured at startup.

import (
	"unsafe"

	"go.mitchellh.com/libghostty/internal/vt"
	"modernc.org/libc"
)

// SysImage is a decoded image, returned by a [SysDecodePngFn].
//
// C: GhosttySysImage
type SysImage struct {
	// Width is the image width in pixels.
	Width uint32

	// Height is the image height in pixels.
	Height uint32

	// Data holds the pixels as 8-bit RGBA, four bytes per pixel. Rows
	// are stored in order starting from the top-left corner, with no
	// padding between them, so a complete image is Width*Height*4 bytes.
	Data []byte
}

// SysDecodePngFn decodes the PNG image in data and returns its pixels.
// Return a non-nil error if the image cannot be decoded. The image is then
// rejected.
//
// data is only valid until the function returns, so copy it if you need
// it later. The pixels in the returned [SysImage] are copied into memory
// owned by libghostty, so the function does not need to keep them alive.
//
// C: GhosttySysDecodePngFn
type SysDecodePngFn func(data []byte) (*SysImage, error)

// sysDecodePngFn is the currently installed Go decode-PNG callback.
var sysDecodePngFn SysDecodePngFn

// SysRandomSecureFn fills data with cryptographically secure random bytes.
// Implementations must use a real CSPRNG because libghostty uses this output
// for secrets such as Kitty clipboard protocol one-time passwords. Return a
// non-nil error when secure entropy is unavailable.
//
// C: GhosttySysRandomSecureFn
type SysRandomSecureFn func(data []byte) error

// sysRandomSecureFn is the currently installed Go secure-random callback.
var sysRandomSecureFn SysRandomSecureFn

// sysSet installs a callback function pointer, or clears it when fp is 0.
func sysSet(option vt.GhosttySysOption, fp uintptr) error {
	tls := getTLS()
	defer putTLS(tls)
	return resultError(vt.Xghostty_sys_set(tls, option, fp))
}

// SysSetDecodePng installs a Go callback that decodes PNG image data
// into RGBA pixels. This enables PNG support in the Kitty Graphics
// Protocol. Pass nil to clear the callback and disable PNG decoding.
//
// This function is not safe for concurrent use. Callers must ensure
// that decode configuration is not modified while terminals may
// process image data (e.g. configure at startup before creating
// terminals).
func SysSetDecodePng(fn SysDecodePngFn) error {
	sysDecodePngFn = fn
	if fn == nil {
		return sysSet(vt.GHOSTTY_SYS_OPT_DECODE_PNG, 0)
	}
	return sysSet(vt.GHOSTTY_SYS_OPT_DECODE_PNG, vt.Fp(goSysDecodePngTrampoline))
}

// SysSetRandomSecure overrides libghostty's platform secure-random source.
// Pass nil to restore the platform default. Targets without a platform source,
// such as wasm32-freestanding, require an override for operations that mint
// secrets.
//
// This function is not safe for concurrent use. Configure it at startup before
// terminals can perform operations that require secure entropy.
func SysSetRandomSecure(fn SysRandomSecureFn) error {
	sysRandomSecureFn = fn
	if fn == nil {
		return sysSet(vt.GHOSTTY_SYS_OPT_RANDOM_SECURE, 0)
	}
	return sysSet(vt.GHOSTTY_SYS_OPT_RANDOM_SECURE, vt.Fp(goSysRandomSecureTrampoline))
}

// SysLogLevel represents the severity level of a log message from the
// library. Maps directly to the C enum values.
// C: GhosttySysLogLevel
type SysLogLevel int

const (
	// SysLogLevelError is the error log level.
	SysLogLevelError SysLogLevel = vt.GHOSTTY_SYS_LOG_LEVEL_ERROR

	// SysLogLevelWarning is the warning log level.
	SysLogLevelWarning SysLogLevel = vt.GHOSTTY_SYS_LOG_LEVEL_WARNING

	// SysLogLevelInfo is the info log level.
	SysLogLevelInfo SysLogLevel = vt.GHOSTTY_SYS_LOG_LEVEL_INFO

	// SysLogLevelDebug is the debug log level.
	SysLogLevelDebug SysLogLevel = vt.GHOSTTY_SYS_LOG_LEVEL_DEBUG
)

// String returns a human-readable name for the log level.
func (l SysLogLevel) String() string {
	switch l {
	case SysLogLevelError:
		return "error"
	case SysLogLevelWarning:
		return "warning"
	case SysLogLevelInfo:
		return "info"
	case SysLogLevelDebug:
		return "debug"
	default:
		return "unknown"
	}
}

// SysLogFn is the Go callback type for log messages from the library.
// The scope identifies the subsystem (e.g. "osc", "kitty"); it is
// empty for unscoped (default) log messages. The message and scope
// are only valid for the duration of the call.
// C: GhosttySysLogFn
type SysLogFn func(level SysLogLevel, scope string, message string)

// sysLogFn is the currently installed Go log callback.
var sysLogFn SysLogFn

// SysSetLog installs a Go callback that receives internal library log
// messages. Pass nil to clear the callback and discard log messages.
//
// Which log levels are emitted depends on the build mode of the
// library. Debug builds emit all levels; release builds emit info
// and above.
//
// This function is not safe for concurrent use. Callers must ensure
// that log configuration is not modified while log messages may be
// delivered (e.g. configure at startup before creating terminals).
func SysSetLog(fn SysLogFn) error {
	sysLogFn = fn
	if fn == nil {
		return sysSet(vt.GHOSTTY_SYS_OPT_LOG, 0)
	}
	return sysSet(vt.GHOSTTY_SYS_OPT_LOG, vt.Fp(goSysLogTrampoline))
}

// SysSetLogStderr installs the built-in stderr log callback provided
// by libghostty. Each message is formatted as "[level](scope): message\n"
// and written to stderr in a thread-safe manner.
//
// This function is not safe for concurrent use. See [SysSetLog].
func SysSetLogStderr() error {
	sysLogFn = nil
	return sysSet(vt.GHOSTTY_SYS_OPT_LOG, vt.Fp(vt.Xghostty_sys_log_stderr))
}

// goSysLogTrampoline is the GhosttySysLogFn installed by SysSetLog.
func goSysLogTrampoline(
	_ *libc.TLS,
	_ uintptr,
	level vt.GhosttySysLogLevel,
	scopePtr uintptr,
	scopeLen uint64,
	messagePtr uintptr,
	messageLen uint64,
) {
	fn := sysLogFn
	if fn == nil {
		return
	}

	var scope string
	if scopeLen > 0 {
		scope = string(goBytes(scopePtr, int(scopeLen)))
	}

	var message string
	if messageLen > 0 {
		message = string(goBytes(messagePtr, int(messageLen)))
	}

	fn(SysLogLevel(level), scope, message)
}

// goSysDecodePngTrampoline is the GhosttySysDecodePngFn installed by
// SysSetDecodePng.
func goSysDecodePngTrampoline(
	tls *libc.TLS,
	_ uintptr,
	allocator uintptr,
	dataPtr uintptr,
	dataLen uint64,
	out uintptr,
) uint8 {
	fn := sysDecodePngFn
	if fn == nil {
		return cFalse
	}

	// Build a Go slice over the input PNG data without copying.
	data := unsafe.Slice((*byte)(unsafe.Pointer(dataPtr)), int(dataLen))

	img, err := fn(data)
	if err != nil || img == nil {
		return cFalse
	}

	// Allocate output pixel buffer through the library's allocator so
	// the library can free it later.
	pixelLen := uint64(len(img.Data))
	buf := vt.Xghostty_alloc(tls, allocator, pixelLen)
	if buf == 0 {
		return cFalse
	}

	// Copy decoded pixels into the library-owned buffer.
	copy(unsafe.Slice((*byte)(unsafe.Pointer(buf)), int(pixelLen)), img.Data)

	image := (*vt.GhosttySysImage)(unsafe.Pointer(out))
	image.Fwidth = img.Width
	image.Fheight = img.Height
	image.Fdata = buf
	image.Fdata_len = pixelLen

	return cTrue
}

// goSysRandomSecureTrampoline is the GhosttySysRandomSecureFn installed by
// SysSetRandomSecure.
func goSysRandomSecureTrampoline(
	_ *libc.TLS,
	_ uintptr,
	buf uintptr,
	length uint64,
) (ok uint8) {
	fn := sysRandomSecureFn
	if fn == nil {
		return cFalse
	}

	// A panic must never unwind into libghostty. A failed random source
	// is represented by false, which makes the originating operation fail
	// with GHOSTTY_IO_ERROR.
	defer func() {
		if recover() != nil {
			ok = cFalse
		}
	}()

	count, valid := ghosttySizeToInt(length)
	if !valid || (count > 0 && buf == 0) {
		return cFalse
	}
	data := unsafe.Slice((*byte)(unsafe.Pointer(buf)), count)
	if err := fn(data); err != nil {
		return cFalse
	}
	return cTrue
}
