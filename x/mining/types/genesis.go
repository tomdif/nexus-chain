package types

type GenesisState struct {
	Params      Params       `json:"params"`
	Jobs        []Job        `json:"jobs"`
	Checkpoints []Checkpoint `json:"checkpoints"`
}

func (gs *GenesisState) Reset()         { *gs = GenesisState{} }
func (gs *GenesisState) String() string { return "" }
func (gs *GenesisState) ProtoMessage()  {}

func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:      DefaultParams(),
		Jobs:        []Job{},
		Checkpoints: []Checkpoint{},
	}
}

func (gs GenesisState) Validate() error {
	return gs.Params.Validate()
}
