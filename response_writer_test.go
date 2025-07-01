package ngebut

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewResponseWriter(t *testing.T) {
	// Persiapan
	httpWriter := httptest.NewRecorder()

	// Eksekusi
	rw := NewResponseWriter(httpWriter)

	// Pemeriksaan
	assert.NotNil(t, rw, "ResponseWriter harus tidak nil")
	assert.Equal(t, http.StatusOK, rw.(*httpResponseWriterAdapter).statusCode, "Status code default harus 200 OK")
	assert.False(t, rw.(*httpResponseWriterAdapter).written, "Flag written awal harus false")
}

func TestResponseWriterHeader(t *testing.T) {
	// Persiapan
	httpWriter := httptest.NewRecorder()
	rw := NewResponseWriter(httpWriter)

	// Eksekusi - set header directly on the response writer
	rw.(*httpResponseWriterAdapter).writer.Header().Set("Content-Type", "application/json")

	// Pemeriksaan
	assert.Equal(t, "application/json", httpWriter.Header().Get("Content-Type"), "Header harus diteruskan ke writer asli")
}

func TestResponseWriterWrite(t *testing.T) {
	// Persiapan
	httpWriter := httptest.NewRecorder()
	rw := NewResponseWriter(httpWriter)

	// Eksekusi
	n, err := rw.Write([]byte("hello world"))

	// Pemeriksaan
	assert.NoError(t, err, "Penulisan tidak boleh menimbulkan error")
	assert.Equal(t, 11, n, "Jumlah byte yang ditulis harus sesuai")
	assert.Equal(t, "hello world", httpWriter.Body.String(), "Data harus ditulis ke writer asli")
}

func TestResponseWriterWriteHeader(t *testing.T) {
	// Persiapan
	httpWriter := httptest.NewRecorder()
	rw := NewResponseWriter(httpWriter)

	// Eksekusi
	rw.WriteHeader(http.StatusCreated)

	// Pemeriksaan
	assert.Equal(t, http.StatusCreated, rw.(*httpResponseWriterAdapter).statusCode, "Status code harus diperbarui")
	assert.Equal(t, http.StatusOK, httpWriter.Code, "Status code tidak boleh dikirim sebelum Flush")
}

func TestResponseWriterFlush(t *testing.T) {
	// Persiapan
	httpWriter := httptest.NewRecorder()
	rw := NewResponseWriter(httpWriter)

	// Set status code first
	rw.WriteHeader(http.StatusCreated)

	// Menambahkan data
	rw.Write([]byte("hello world"))

	// Eksekusi
	rw.Flush()

	// Pemeriksaan
	assert.Equal(t, http.StatusCreated, httpWriter.Code, "Status code harus diteruskan ke writer asli")
	assert.Equal(t, "hello world", httpWriter.Body.String(), "Body harus ditulis ke writer asli")
	assert.True(t, rw.(*httpResponseWriterAdapter).written, "Flag written harus diatur ke true")

	// Eksekusi flush kedua kalinya (sekarang menulis langsung)
	rw.Write([]byte(" tambahan"))
	rw.Flush()

	// Pemeriksaan perubahan setelah flush kedua
	assert.Equal(t, "hello world tambahan", httpWriter.Body.String(), "Body harus diperbarui dengan data baru")
}

func TestReleaseResponseWriter(t *testing.T) {
	// Persiapan
	httpWriter := httptest.NewRecorder()
	rw := NewResponseWriter(httpWriter)

	// Eksekusi
	ReleaseResponseWriter(rw)

	// Pemeriksaan
	adapter, ok := rw.(*httpResponseWriterAdapter)
	require.True(t, ok, "ResponseWriter harus bertipe httpResponseWriterAdapter")
	assert.Nil(t, adapter.writer, "Writer harus diatur ke nil setelah release")
	assert.Nil(t, adapter.header, "Header harus diatur ke nil setelah release")
}
