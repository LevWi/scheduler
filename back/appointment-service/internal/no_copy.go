package common

// See sync.Mutex
type NoCopy struct{}

func (*NoCopy) Lock()   {}
func (*NoCopy) Unlock() {}
