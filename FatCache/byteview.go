package fatcache

type ByteView struct {
	b []byte
}

func (bv ByteView) Len() int {
	return len(bv.b)
}

func (bv ByteView) ByteSlice() []byte {
	copied := make([]byte, len(bv.b))
	copy(copied, bv.b)
	return copied
}

func (bv ByteView) String() string {
	return string(bv.b)
}
