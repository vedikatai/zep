package embeddings

import (
	"runtime"
	"testing"
)

func synth(n, dims int) [][]float32 {
	out := make([][]float32, n)
	for i := range out {
		row := make([]float32, dims)
		for d := 0; d < dims; d++ {
			row[d] = float32((i*dims+d)%1000) / 1000
		}
		out[i] = row
	}
	return out
}

func memStats() uint64 {
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc
}

// BenchmarkMemoryFloat32VsFloat16 reports alloc bytes for 10k x 1536 vectors.
func BenchmarkMemoryFloat32VsFloat16(b *testing.B) {
	const n, dims = 10_000, 1536
	data := synth(n, dims)

	b.Run("float32_slices", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = data // already allocated; measure retention via KeepAlive
			runtime.KeepAlive(data)
		}
		b.ReportMetric(float64(n*dims*4), "bytes/corpus")
	})

	b.Run("float16_vectors", func(b *testing.B) {
		b.ReportAllocs()
		var corpus []Vector
		b.StopTimer()
		corpus = make([]Vector, n)
		for i := range data {
			corpus[i] = FromFloat32(data[i])
		}
		b.StartTimer()
		for i := 0; i < b.N; i++ {
			runtime.KeepAlive(corpus)
		}
		b.ReportMetric(float64(n*dims*2), "bytes/corpus")
	})
}

func BenchmarkEncodeDecode(b *testing.B) {
	v := FromFloat32(synth(1, 1536)[0])
	blob, _ := v.Encode()
	b.Run("encode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = v.Encode()
		}
	})
	b.Run("decode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = Decode(blob)
		}
	})
}

// MemoryFootprintSynthetic is a helper test that prints before/after sizes (not a fail gate).
func TestMemoryFootprintSynthetic(t *testing.T) {
	const n, dims = 5000, 768
	data := synth(n, dims)
	f32Bytes := n * dims * 4

	before := memStats()
	_ = before
	corpus := make([]Vector, n)
	for i := range data {
		corpus[i] = FromFloat32(data[i])
	}
	f16Bytes := 0
	for _, v := range corpus {
		f16Bytes += v.ByteSize()
	}
	t.Logf("synthetic corpus n=%d dims=%d", n, dims)
	t.Logf("float32 logical bytes: %d (%.2f MiB)", f32Bytes, float64(f32Bytes)/1024/1024)
	t.Logf("float16 logical bytes: %d (%.2f MiB)", f16Bytes, float64(f16Bytes)/1024/1024)
	t.Logf("savings: %.1f%%", 100*(1-float64(f16Bytes)/float64(f32Bytes)))
	if f16Bytes*2 != f32Bytes {
		t.Fatalf("expected exactly 50%% storage: f16=%d f32=%d", f16Bytes, f32Bytes)
	}
	runtime.KeepAlive(corpus)
}
