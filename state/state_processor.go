package state

type Processor struct {
	window *window
}

// later add processor config
func NewProcessor(state *StateDB, winSize int) *Processor {
	return &Processor{
		window: NewWindow(state, winSize),
	}
}

func (p *Processor) ApplyTransactions(txes []*Transaction) []*Receipt {
	receipts := make([]*Receipt, 0, len(txes))
	for i := 0; i < len(txes); i += p.window.size {
		start := i
		end := start + p.window.size
		if end >= len(txes) {
			end = len(txes) - 1
		}

		batch := p.window.process(txes[start:end])
		receipts = append(receipts, batch...)
	}
	return receipts
}

// handle deploy seperately, recreate execution window etc..
func (p *Processor) Deploy() {

}
