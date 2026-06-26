package embeddings

import (
	"math"
	"testing"
)

func TestRoundTripFloat16(t *testing.T) {
	// Realistic embedding magnitudes (L2-normalised components ~[-1, 1] plus a few larger).
	cases := []float32{0, 1, -1, 0.5, -0.5, 3.14159, 1e-3, 1e2, 0.123456, -0.987654}
	for _, c := range cases {
		bits := Float32ToFloat16bits(c)
		back := Float16bitsToFloat32(bits)
		// float16 has ~3 decimal digits; allow relative error
		if math.Abs(float64(back-c)) > 0.01*math.Abs(float64(c))+1e-3 {
			t.Fatalf("roundtrip %v -> %v (bits=%04x)", c, back, bits)
		}
	}
}

func TestEncodeDecodeFloat16(t *testing.T) {
	src := []float32{0.1, 0.2, 0.3, -0.4, 0.5}
	v := FromFloat32(src)
	blob, err := v.Encode()
	if err != nil {
		t.Fatal(err)
	}
	if !IsFloat16Blob(blob) {
		t.Fatal("expected float16 blob")
	}
	got, err := Decode(blob)
	if err != nil {
		t.Fatal(err)
	}
	if got.Dims() != len(src) {
		t.Fatalf("dims %d != %d", got.Dims(), len(src))
	}
	// storage is half of float32
	if got.ByteSize() != len(src)*2 {
		t.Fatalf("byte size %d", got.ByteSize())
	}
}

func TestDecodeLegacyFloat32Blob(t *testing.T) {
	src := []float32{0.25, 0.5, 0.75, 1.0}
	blob, err := EncodeFloat32Legacy(src)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Decode(blob)
	if err != nil {
		t.Fatal(err)
	}
	// Re-encode as f16
	out, err := got.Encode()
	if err != nil {
		t.Fatal(err)
	}
	if !IsFloat16Blob(out) {
		t.Fatal("migration should produce float16")
	}
	if len(out) >= len(blob) {
		t.Fatalf("float16 blob should be smaller: f16=%d f32=%d", len(out), len(blob))
	}
}

func TestDecodeRawFloat32(t *testing.T) {
	// Unheadered little-endian float32 array (legacy CE dumps)
	src := []float32{1, 0, 0, 0}
	raw := make([]byte, len(src)*4)
	for i, f := range src {
		// use EncodeFloat32Legacy payload only — hand-roll
		u := math.Float32bits(f)
		raw[i*4] = byte(u)
		raw[i*4+1] = byte(u >> 8)
		raw[i*4+2] = byte(u >> 16)
		raw[i*4+3] = byte(u >> 24)
	}
	got, err := Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got.Dims() != 4 {
		t.Fatalf("dims=%d", got.Dims())
	}
}

func TestCosineSimilarityIdentical(t *testing.T) {
	v := FromFloat32([]float32{1, 0, 0})
	s, err := CosineSimilarity(v, v)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(float64(s-1)) > 1e-3 {
		t.Fatalf("sim=%v", s)
	}
}

func TestVectorSearchRetrieval(t *testing.T) {
	// Mini vector search: query vs corpus, ensure nearest neighbour preserved under f16.
	corpusF32 := [][]float32{
		{1, 0, 0, 0},
		{0.9, 0.1, 0, 0},
		{0, 1, 0, 0},
		{0, 0, 1, 0},
	}
	queryF32 := []float32{1, 0, 0, 0}
	corpus := make([]Vector, len(corpusF32))
	for i, row := range corpusF32 {
		blob, _ := FromFloat32(row).Encode()
		v, err := Decode(blob)
		if err != nil {
			t.Fatal(err)
		}
		corpus[i] = v
	}
	q := FromFloat32(queryF32)
	bestIdx, bestSim := -1, float32(-2)
	for i, row := range corpus {
		s, err := CosineSimilarity(q, row)
		if err != nil {
			t.Fatal(err)
		}
		if s > bestSim {
			bestSim = s
			bestIdx = i
		}
	}
	if bestIdx != 0 {
		t.Fatalf("expected nn=0 got %d (sim=%v)", bestIdx, bestSim)
	}
}

func TestMaximalMarginalRelevanceF16(t *testing.T) {
	q := FromFloat32([]float32{1, 0, 0, 0})
	corpus := []Vector{
		FromFloat32([]float32{1, 0, 0, 0}),
		FromFloat32([]float32{0.9, 0.1, 0, 0}),
		FromFloat32([]float32{0, 1, 0, 0}),
		FromFloat32([]float32{0, 0, 1, 0}),
	}
	for i := range corpus {
		blob, _ := corpus[i].Encode()
		v, err := Decode(blob)
		if err != nil {
			t.Fatal(err)
		}
		corpus[i] = v
	}
	idxs, err := MaximalMarginalRelevance(q, corpus, 0.5, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(idxs) != 2 || idxs[0] != 0 {
		t.Fatalf("got %v", idxs)
	}
}
