package types

const (
	// ModuleName defines the module name
	ModuleName = "mining"
	// StoreKey defines the primary module store key
	StoreKey = ModuleName
	// RouterKey defines the module's message routing key
	RouterKey = ModuleName
	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

// KVStore key prefixes
var (
	JobKeyPrefix             = []byte{0x01}
	CheckpointKeyPrefix      = []byte{0x02}
	LastCheckpointIDKey      = []byte{0x03}
	ShareKeyPrefix           = []byte{0x04}
	ValidatorRecordKeyPrefix = []byte{0x05}
	ParamsKey                = []byte{0x06}
	ValidatorRewardPoolKey   = []byte{0x07}

	// Collaborative mining prefixes
	WorkSubmissionKeyPrefix = []byte{0x10}
	WorkShareKeyPrefix      = []byte{0x11}
	BonusShareKeyPrefix     = []byte{0x12}
	EpochKeyPrefix          = []byte{0x13}
)
