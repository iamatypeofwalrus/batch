package batch

// Logger is a logging interface that Batch will use to communicate internal errors
type Logger interface {
	Print(v ...interface{})
}
