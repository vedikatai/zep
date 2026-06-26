// Package embeddings provides float16 storage codecs for embedding vectors.
//
// Wire format (v1):
//
//	magic[4] = "ZEP1"
//	version  uint8  = 1
//	dtype    uint8  = 1  (1 = IEEE-754 binary16 / float16)
//	ndims    uint32 little-endian
//	payload  ndims * 2 bytes (little-endian uint16 bit patterns)
//
// In-memory compute paths still use float32; only durable / on-the-wire
// storage uses float16 (half the RAM and disk of float32 arrays).
package embeddings

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
)

const (
	// Magic identifies Zep embedding blobs.
	Magic = "ZEP1"
	// FormatVersion is the current on-disk format version.
	FormatVersion uint8 = 1
	// DTypeFloat16 is IEEE-754 binary16.
	DTypeFloat16 uint8 = 1
	// DTypeFloat32 is legacy IEEE-754 binary32 (migration source only).
	DTypeFloat32 uint8 = 2

	headerSize = 4 + 1 + 1 + 4 // magic + version + dtype + ndims
)

var (
	ErrEmptyVector   = errors.New("embeddings: empty vector")
	ErrBadMagic      = errors.New("embeddings: bad magic")
	ErrBadVersion    = errors.New("embeddings: unsupported version")
	ErrTruncated     = errors.New("embeddings: truncated payload")
	ErrDimMismatch   = errors.New("embeddings: dimension mismatch")
	ErrUnknownDType  = errors.New("embeddings: unknown dtype")
)

// Vector is an embedding held in float16 storage form (2 bytes per dim).
// Call Float32() before similarity / MMR compute paths that need float32.
type Vector []uint16

// FromFloat32 converts a float32 embedding into float16 storage form.
func FromFloat32(src []float32) Vector {
	if len(src) == 0 {
		return nil
	}
	out := make(Vector, len(src))
	for i, v := range src {
		out[i] = Float32ToFloat16bits(v)
	}
	return out
}

// Float32 expands the vector to float32 for compute (cosine, MMR, etc.).
func (v Vector) Float32() []float32 {
	if len(v) == 0 {
		return nil
	}
	out := make([]float32, len(v))
	for i, bits := range v {
		out[i] = Float16bitsToFloat32(bits)
	}
	return out
}

// Dims returns the number of dimensions.
func (v Vector) Dims() int { return len(v) }

// ByteSize returns storage size in bytes (2 * dims).
func (v Vector) ByteSize() int { return len(v) * 2 }

// Encode serializes the vector to the ZEP1 float16 blob format.
func (v Vector) Encode() ([]byte, error) {
	if len(v) == 0 {
		return nil, ErrEmptyVector
	}
	buf := make([]byte, headerSize+len(v)*2)
	copy(buf[0:4], Magic)
	buf[4] = FormatVersion
	buf[5] = DTypeFloat16
	binary.LittleEndian.PutUint32(buf[6:10], uint32(len(v)))
	off := headerSize
	for _, bits := range v {
		binary.LittleEndian.PutUint16(buf[off:off+2], bits)
		off += 2
	}
	return buf, nil
}

// EncodeFloat32Legacy serializes a float32 slice in the legacy ZEP1 float32 blob
// (used only as migration input / tests). Prefer Vector.Encode for new data.
func EncodeFloat32Legacy(src []float32) ([]byte, error) {
	if len(src) == 0 {
		return nil, ErrEmptyVector
	}
	buf := make([]byte, headerSize+len(src)*4)
	copy(buf[0:4], Magic)
	buf[4] = FormatVersion
	buf[5] = DTypeFloat32
	binary.LittleEndian.PutUint32(buf[6:10], uint32(len(src)))
	off := headerSize
	for _, f := range src {
		binary.LittleEndian.PutUint32(buf[off:off+4], math.Float32bits(f))
		off += 4
	}
	return buf, nil
}

// Decode parses a ZEP1 blob into a float16 Vector. Float32 legacy blobs are
// converted on the fly so readers are format-agnostic during migration.
func Decode(blob []byte) (Vector, error) {
	if len(blob) < headerSize {
		return nil, ErrTruncated
	}
	if string(blob[0:4]) != Magic {
		// Raw float32 little-endian array (no header) — treat as legacy CE dumps.
		if len(blob)%4 == 0 && len(blob) > 0 {
			return decodeRawFloat32(blob), nil
		}
		return nil, ErrBadMagic
	}
	if blob[4] != FormatVersion {
		return nil, fmt.Errorf("%w: %d", ErrBadVersion, blob[4])
	}
	dtype := blob[5]
	ndims := int(binary.LittleEndian.Uint32(blob[6:10]))
	if ndims <= 0 {
		return nil, ErrEmptyVector
	}
	switch dtype {
	case DTypeFloat16:
		need := headerSize + ndims*2
		if len(blob) < need {
			return nil, ErrTruncated
		}
		out := make(Vector, ndims)
		off := headerSize
		for i := 0; i < ndims; i++ {
			out[i] = binary.LittleEndian.Uint16(blob[off : off+2])
			off += 2
		}
		return out, nil
	case DTypeFloat32:
		need := headerSize + ndims*4
		if len(blob) < need {
			return nil, ErrTruncated
		}
		tmp := make([]float32, ndims)
		off := headerSize
		for i := 0; i < ndims; i++ {
			tmp[i] = math.Float32frombits(binary.LittleEndian.Uint32(blob[off : off+4]))
			off += 4
		}
		return FromFloat32(tmp), nil
	default:
		return nil, fmt.Errorf("%w: %d", ErrUnknownDType, dtype)
	}
}

func decodeRawFloat32(blob []byte) Vector {
	n := len(blob) / 4
	tmp := make([]float32, n)
	for i := 0; i < n; i++ {
		tmp[i] = math.Float32frombits(binary.LittleEndian.Uint32(blob[i*4 : i*4+4]))
	}
	return FromFloat32(tmp)
}

// IsFloat16Blob reports whether blob is already ZEP1 float16.
func IsFloat16Blob(blob []byte) bool {
	return len(blob) >= headerSize &&
		string(blob[0:4]) == Magic &&
		blob[4] == FormatVersion &&
		blob[5] == DTypeFloat16
}

// Float32ToFloat16bits converts a float32 to IEEE-754 binary16 bits (round-to-nearest-even).
func Float32ToFloat16bits(f float32) uint16 {
	bits := math.Float32bits(f)
	sign := uint16((bits >> 16) & 0x8000)
	exp := int((bits>>23)&0xff) - 127 + 15
	mant := bits & 0x7fffff

	switch {
	case (bits>>23)&0xff == 0xff: // Inf / NaN
		if mant != 0 {
			return sign | 0x7e00 // qNaN
		}
		return sign | 0x7c00 // Inf
	case exp >= 0x1f: // overflow -> Inf
		return sign | 0x7c00
	case exp <= 0:
		// subnormal or zero in f16
		if exp < -10 {
			return sign
		}
		mant |= 0x800000
		shift := uint32(14 - exp)
		mant = (mant + (1 << (shift - 1))) >> shift
		return sign | uint16(mant)
	default:
		// normal
		rounded := mant + 0x1000 // round to nearest even (sticky on lower bits simplified)
		if rounded&0x800000 != 0 {
			// mantissa overflow
			exp++
			rounded = 0
			if exp >= 0x1f {
				return sign | 0x7c00
			}
		}
		return sign | uint16(exp<<10) | uint16(rounded>>13)
	}
}

// Float16bitsToFloat32 expands IEEE-754 binary16 bits to float32.
func Float16bitsToFloat32(h uint16) float32 {
	sign := uint32(h&0x8000) << 16
	exp := (h >> 10) & 0x1f
	mant := uint32(h & 0x3ff)

	switch exp {
	case 0:
		if mant == 0 {
			return math.Float32frombits(sign)
		}
		// subnormal -> normalize
		exp32 := uint32(127 - 15 + 1)
		for mant&0x400 == 0 {
			mant <<= 1
			exp32--
		}
		mant &= 0x3ff
		return math.Float32frombits(sign | (exp32 << 23) | (mant << 13))
	case 0x1f:
		if mant == 0 {
			return math.Float32frombits(sign | 0x7f800000) // Inf
		}
		return math.Float32frombits(sign | 0x7fc00000) // NaN
	default:
		exp32 := uint32(exp) - 15 + 127
		return math.Float32frombits(sign | (exp32 << 23) | (mant << 13))
	}
}

// CosineSimilarity computes cosine similarity on float16 vectors (promoted to f32).
func CosineSimilarity(a, b Vector) (float32, error) {
	if len(a) == 0 || len(b) == 0 {
		return 0, ErrEmptyVector
	}
	if len(a) != len(b) {
		return 0, ErrDimMismatch
	}
	fa, fb := a.Float32(), b.Float32()
	var dot, na, nb float64
	for i := range fa {
		dot += float64(fa[i]) * float64(fb[i])
		na += float64(fa[i]) * float64(fa[i])
		nb += float64(fb[i]) * float64(fb[i])
	}
	if na == 0 || nb == 0 {
		return 0, nil
	}
	return float32(dot / (math.Sqrt(na) * math.Sqrt(nb))), nil
}
