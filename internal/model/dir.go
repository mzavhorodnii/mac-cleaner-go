package model

type Dir struct {
	Path string
	Size int64
}

func (d Dir) SizeGB() float64 {
	return float64(d.Size) / 1e9
}
