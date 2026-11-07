static zig_u128 rt_mul(zig_u128 a, zig_u128 b) {
	uint64_t a0 = a.lo & 0xffffffffu, a1 = a.lo >> 32;
	uint64_t b0 = b.lo & 0xffffffffu, b1 = b.lo >> 32;
	uint64_t p00 = a0 * b0, p01 = a0 * b1, p10 = a1 * b0, p11 = a1 * b1;
	uint64_t mid = (p00 >> 32) + (p01 & 0xffffffffu) + (p10 & 0xffffffffu);
	zig_u128 r;
	r.lo = (mid << 32) | (p00 & 0xffffffffu);
	r.hi = p11 + (p01 >> 32) + (p10 >> 32) + (mid >> 32) + a.lo * b.hi + a.hi * b.lo;
	return r;
}

static zig_u128 rt_udivmod(zig_u128 n, zig_u128 d, zig_u128 *rem) {
	zig_u128 q = {0}, r = {0};
	int i;
	if (n.hi == 0 && d.hi == 0) {
		q.lo = n.lo / d.lo;
		r.lo = n.lo % d.lo;
		*rem = r;
		return q;
	}
	for (i = 127; i >= 0; i--) {
		r.hi = (r.hi << 1) | (r.lo >> 63);
		r.lo = (r.lo << 1) | ((i >= 64 ? n.hi >> (i - 64) : n.lo >> i) & 1);
		if (r.hi > d.hi || (r.hi == d.hi && r.lo >= d.lo)) {
			uint64_t borrow = r.lo < d.lo;
			r.lo -= d.lo;
			r.hi -= d.hi + borrow;
			if (i >= 64) q.hi |= (uint64_t)1 << (i - 64);
			else q.lo |= (uint64_t)1 << i;
		}
	}
	*rem = r;
	return q;
}

static zig_u128 rt_u(zig_i128 a) { zig_u128 r; r.lo = a.lo; r.hi = (uint64_t)a.hi; return r; }
static zig_i128 rt_i(zig_u128 a) { zig_i128 r; r.lo = a.lo; r.hi = (int64_t)a.hi; return r; }
static zig_u128 rt_neg(zig_u128 a) { zig_u128 r; r.lo = ~a.lo + 1; r.hi = ~a.hi + (r.lo == 0); return r; }
static zig_u128 rt_abs(zig_i128 a) { return a.hi < 0 ? rt_neg(rt_u(a)) : rt_u(a); }

zig_i128 __multi3(zig_i128 a, zig_i128 b) { return rt_i(rt_mul(rt_u(a), rt_u(b))); }
zig_u128 __udivti3(zig_u128 a, zig_u128 b) { zig_u128 r; return rt_udivmod(a, b, &r); }
zig_u128 __umodti3(zig_u128 a, zig_u128 b) { zig_u128 r; rt_udivmod(a, b, &r); return r; }
zig_i128 __divti3(zig_i128 a, zig_i128 b) {
	zig_u128 r, q = rt_udivmod(rt_abs(a), rt_abs(b), &r);
	return rt_i((a.hi < 0) != (b.hi < 0) ? rt_neg(q) : q);
}
zig_i128 __modti3(zig_i128 a, zig_i128 b) {
	zig_u128 r;
	rt_udivmod(rt_abs(a), rt_abs(b), &r);
	return rt_i(a.hi < 0 ? rt_neg(r) : r);
}
