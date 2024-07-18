package ngebut

type Config struct {
	Addr         string
	MultiCore    bool
	ErrorHandler ErrorHandlerFunc
}
