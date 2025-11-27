package docking

// Reward per ligand based on rotatable bonds (complexity)
// More bonds = more compute time = more reward

var BondMultiplier = map[int]float64{
	0: 0.5, 1: 0.6, 2: 0.7, 3: 0.8, 4: 0.9,
	5: 1.0, 6: 1.2, 7: 1.4, 8: 1.6, 9: 1.8,
	10: 2.0, 11: 2.3, 12: 2.6, 13: 3.0,
}

const BaseReward = 1000 // unexus per ligand

func GetReward(rotatableBonds int) int64 {
	mult, ok := BondMultiplier[rotatableBonds]
	if !ok {
		if rotatableBonds < 0 {
			mult = 0.5
		} else {
			mult = 3.0
		}
	}
	return int64(float64(BaseReward) * mult)
}
