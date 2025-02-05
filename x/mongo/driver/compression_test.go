// Copyright (C) MongoDB, Inc. 2017-present.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

package driver

import (
	"os"
	"testing"

	"go.mongodb.org/mongo-driver/internal/assert"
	"go.mongodb.org/mongo-driver/x/mongo/driver/wiremessage"
)

func TestCompression(t *testing.T) {
	compressors := []wiremessage.CompressorID{
		wiremessage.CompressorNoOp,
		wiremessage.CompressorSnappy,
		wiremessage.CompressorZLib,
		wiremessage.CompressorZstd,
	}

	for _, compressor := range compressors {
		t.Run(compressor.String(), func(t *testing.T) {
			payload := []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt")
			opts := CompressionOpts{
				Compressor:       compressor,
				ZlibLevel:        wiremessage.DefaultZlibLevel,
				ZstdLevel:        wiremessage.DefaultZstdLevel,
				UncompressedSize: int32(len(payload)),
			}
			compressed, err := CompressPayload(payload, opts)
			assert.NoError(t, err)
			assert.NotEqual(t, 0, len(compressed))
			decompressed, err := DecompressPayload(compressed, opts)
			assert.NoError(t, err)
			assert.Equal(t, payload, decompressed)
		})
	}
}

func TestDecompressFailures(t *testing.T) {
	t.Parallel()

	t.Run("snappy decompress huge size", func(t *testing.T) {
		t.Parallel()

		opts := CompressionOpts{
			Compressor:       wiremessage.CompressorSnappy,
			UncompressedSize: 100, // reasonable size
		}
		// Compressed data is twice as large as declared above.
		// In test we use actual compression so that the decompress action would pass without fix (thus failing test).
		// When decompression starts it allocates a buffer of the defined size, regardless of a valid compressed body following.
		compressedData, err := CompressPayload(make([]byte, opts.UncompressedSize*2), opts)
		assert.NoError(t, err, "premature error making compressed example")

		_, err = DecompressPayload(compressedData, opts)
		assert.Error(t, err)
	})
}

func BenchmarkCompressPayload(b *testing.B) {
	payload := func() []byte {
		buf, err := os.ReadFile("compression.go")
		if err != nil {
			b.Log(err)
			b.FailNow()
		}
		for i := 1; i < 10; i++ {
			buf = append(buf, buf...)
		}
		return buf
	}()

	compressors := []wiremessage.CompressorID{
		wiremessage.CompressorSnappy,
		wiremessage.CompressorZLib,
		wiremessage.CompressorZstd,
	}

	for _, compressor := range compressors {
		b.Run(compressor.String(), func(b *testing.B) {
			opts := CompressionOpts{
				Compressor: compressor,
				ZlibLevel:  wiremessage.DefaultZlibLevel,
				ZstdLevel:  wiremessage.DefaultZstdLevel,
			}
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					_, err := CompressPayload(payload, opts)
					if err != nil {
						b.Error(err)
					}
				}
			})
		})
	}
}
