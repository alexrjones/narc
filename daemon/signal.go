package daemon

type Signal uint8

const (
	SignalTerm Signal = iota
	SignalHup
)

type SignalPacket struct {
	Signal                 Signal
	LastActivityName       string
	LastActivityIgnoreIdle bool
}
