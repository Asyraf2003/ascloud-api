package zipsec

type Options struct {
	MaxTotalBytes int64
	MaxFiles      int
	MaxDepth      int
	MaxFileBytes  int64
}

type Result struct {
	Files      int
	TotalBytes int64
}

func DefaultOptions() Options {
	return Options{
		MaxTotalBytes: 20 * 1024 * 1024,
		MaxFiles:      2000,
		MaxDepth:      20,
		MaxFileBytes:  20 * 1024 * 1024,
	}
}
