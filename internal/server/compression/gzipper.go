package compression

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/alphaonly/gomart/internal/schema"
)

func GZipCompressionHandler(next http.Handler) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("GZipCompressionHandler invoked")
		log.Printf("requsest Content-Encoding:%v", r.Header.Get("Content-Encoding"))
		//read body
		var bytesData []byte
		var err error
		var prev schema.PreviousBytes
		if p := r.Context().Value(schema.PKey1); p != nil {
			prev = p.(schema.PreviousBytes)
		}
		if prev != nil {
			//body from previous handler
			bytesData = prev
			log.Printf("got body from previous handler:%v", string(bytesData))
		} else {
			//body from request if there is no previous handler
			bytesData, err = io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotImplemented)
				return
			}
			log.Printf("got body from request:%v", string(bytesData))
		}
		//compression validation
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			//Compression logic
			compressedByteData, err := GzipCompress(bytesData)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotImplemented)
				return
			}
			bytesData = *compressedByteData
		}
		log.Printf("Check response Content-Encoding in final header, value:%v", w.Header().Get("Content-Encoding"))
		log.Printf("Check response Content-Type in final header, value:%v", w.Header().Get("Content-Type"))

		//compressed response to next handler
		if next != nil {
			//write compressed body for further handle
			prev = bytesData
			ctx := context.WithValue(r.Context(), schema.PKey1, prev)
			//call further handler with context parameters
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		log.Fatal("Gzip comp handler requires next handler not nil")

	}
}

func GZipDeCompressionHandler(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("GZipDeCompressionHandler invoked")

		//validation for decompression
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {

			//read compressed body
			byteData, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotImplemented)
				return
			}
			//Decompression logic

			decompressedByteData, err := GzipDecompress(byteData)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotImplemented)
				return
			}
			//decompressed response to next handler
			if next != nil {
				r.Body = io.NopCloser(bytes.NewReader(decompressedByteData))
				next.ServeHTTP(w, r)
				return
			}
			//decompressed response to requester
			_, err = w.Write(decompressedByteData)
			if err != nil {
				log.Println("decompressed writing error")
				http.Error(w, "decompressed writing error", http.StatusInternalServerError)
				return
			}
		}

		//compressed response
		//initial response to next handler
		if next != nil {
			next.ServeHTTP(w, r)
			return
		}
		log.Fatal("Gzip deComp handler requires next handler not nil")

	}
}

func GzipCompress(data []byte) (*[]byte, error) {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)

	_, err := w.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed write data to GzipCompress temporary buffer: %v", err)
	}
	err = w.Flush()
	if err != nil {
		return nil, err
	}
	err = w.Close()
	if err != nil {
		return nil, fmt.Errorf("failed GzipCompress data: %v", err)
	}

	compressedBytes := b.Bytes()

	return &compressedBytes, nil
}

func GzipDecompress(data []byte) ([]byte, error) {
	reader := bytes.NewReader(data)
	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		log.Fatalf("Gzip decompress error:%v", err)
	}
	defer func() {
		err := gzipReader.Close()
		if err != nil {
			log.Fatal("Close error")
		}
	}()
	decompressedBytes, err := io.ReadAll(gzipReader)
	if err != nil {
		return nil, fmt.Errorf("failed GzipDecompress data: %v", err)
	}

	return decompressedBytes, nil
}
