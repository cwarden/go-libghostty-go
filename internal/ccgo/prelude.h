#define __has_attribute(x) __ccgo_has_attribute_##x
#define __ccgo_has_attribute_aligned 1
#define __ccgo_has_attribute_packed 1
#define __ccgo_has_attribute_noinline 1
#define __ccgo_has_attribute_weak 1
#define __ccgo_has_attribute_visibility 1
#define __ccgo_has_attribute_unused 1
#define __ccgo_has_attribute_noreturn 1
#define __ccgo_has_attribute_section 1
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <fcntl.h>
#include <errno.h>
#include <sys/mman.h>
#include <sys/uio.h>
long __syscall6(long, long, long, long, long, long, long);
/* Zig's unreachable becomes a panic with a Go traceback instead of libc's
 * exit(1), so the failing code path can be found. */
__attribute__((noreturn)) void ccgo_unreachable(void);
#define __builtin_unreachable() ccgo_unreachable()
