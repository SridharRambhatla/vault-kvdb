package utils

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
)

// CompressData compresses the given data using gzip
func CompressData(data interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)

	if _, err := zw.Write(jsonData); err != nil {
		return nil, err
	}

	if err := zw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// DecompressData decompresses the given data using gzip
func DecompressData(compressedData []byte) ([]byte, error) {
	zr, err := gzip.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	return io.ReadAll(zr)
}

// CompressAndMarshal compresses and marshals the given data
func CompressAndMarshal(data interface{}) ([]byte, error) {
	compressed, err := CompressData(data)
	if err != nil {
		return nil, err
	}

	return json.Marshal(compressed)
}

// UnmarshalAndDecompress unmarshals and decompresses the given data
func UnmarshalAndDecompress(data []byte, v interface{}) error {
	var compressed []byte
	if err := json.Unmarshal(data, &compressed); err != nil {
		return err
	}

	decompressed, err := DecompressData(compressed)
	if err != nil {
		return err
	}

	return json.Unmarshal(decompressed, v)
}
