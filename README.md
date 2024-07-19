# Ngebut

Ngebut adalah sebuah web framework untuk Go yang dirancang untuk kecepatan dan efisiensi.
Ngebut dibangun diatas [gnet](https://github.com/panjf2000/gnet), sebuah library non-blocking networking tercepat untuk Go.

## Peringatan

Ngebut masih dalam tahap pengembangan dan belum siap untuk digunakan di production, disarankan untuk menggunakan Ngebut
saat sudah rilis versi stabil.

## Instalasi

```bash
go get -u github.com/ryanbekhen/ngebut
```

## Contoh Penggunaan

```go
package main

import (
	"github.com/ryanbekhen/ngebut"
	"strconv"
)

func main() {
	server := &ngebut.Server{
		Addr: "tcp://:3000",
		Handler: ngebut.HandlerFunc(func(w ngebut.ResponseWriter, r *ngebut.Request) {
			message := ""
			for k, v := range r.Header {
				message += k + ": " + v[0] + "\n"
			}

			message += "IP: " + r.RemoteAddr + "\n"
			message += "Content-Length: " + strconv.Itoa(int(r.ContentLength)) + "\n"
			message += "Method: " + r.Method + "\n"
			message += "URL: " + r.RequestURI + "\n"
			message += "Proto: " + r.Proto + "\n"

			w.Write([]byte(message))
		}),
	}

	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}

```

## Lisensi

Ngebut dilisensikan di bawah lisensi MIT. Lihat [LISENSI](LISENSI) untuk informasi lebih lanjut.