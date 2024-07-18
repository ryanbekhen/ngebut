# Ngebut

Ngebut adalah sebuah web framework yang terinspirasi dari [Fiber](https://github.com/gofiber/fiber) untuk Go. 
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
	"log"
)

func main() {
	app := ngebut.New(ngebut.Config{
		Addr: "tcp://:3000",
	})

	app.Get("/", func(c *ngebut.Context) error {
		return c.SendString("Hello, World!")
	})

	log.Fatal(app.Run())
}

```

## Lisensi

Ngebut dilisensikan di bawah lisensi MIT. Lihat [LISENSI](LISENSI) untuk informasi lebih lanjut.