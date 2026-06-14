package greet

type Formatter struct {
	Prefix string
}

func (f Formatter) Format(s string) string {
	return f.Prefix + Hello(s)
}
