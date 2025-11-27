package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
)

func main() {
	blockHash := "F5DF714BC6960DE05739F86AC33074AF9AAAD7054FC71B14B8A0790A890D6EFF"
	hashBytes, _ := hex.DecodeString(blockHash)

	var seed int64
	for i := 0; i < 8 && i < len(hashBytes); i++ {
		seed = (seed << 8) | int64(hashBytes[i])
	}
	rng := rand.New(rand.NewSource(seed))

	size := int64(64)
	numCouplings := size * 4

	couplings := make([][3]int64, numCouplings)
	for i := int64(0); i < numCouplings; i++ {
		spin1 := rng.Int63n(size)
		spin2 := rng.Int63n(size)
		for spin2 == spin1 {
			spin2 = rng.Int63n(size)
		}
		strength := rng.Int63n(21) - 10
		couplings[i] = [3]int64{spin1, spin2, strength}
	}

	problemData := fmt.Sprintf("%v", couplings)
	problemHashBytes := sha256.Sum256([]byte(problemData))
	problemHash := hex.EncodeToString(problemHashBytes[:])
	threshold := -size / 2

	fmt.Println("=== ISING PROBLEM ===")
	fmt.Printf("Size: %d, Threshold: %d\n", size, threshold)
	fmt.Printf("Problem Hash: %s\n", problemHash[:32])

	// Solve with simulated annealing
	spins := make([]int, size)
	for i := range spins {
		if rng.Float64() < 0.5 {
			spins[i] = 1
		} else {
			spins[i] = -1
		}
	}

	calcEnergy := func() int64 {
		var e int64
		for _, c := range couplings {
			e += c[2] * int64(spins[c[0]]) * int64(spins[c[1]])
		}
		return e
	}

	energy := calcEnergy()
	bestEnergy := energy
	bestSpins := make([]int, size)
	copy(bestSpins, spins)

	temp := 10.0
	for iter := 0; iter < 100000; iter++ {
		i := rng.Intn(int(size))
		spins[i] *= -1
		newEnergy := calcEnergy()
		delta := newEnergy - energy

		if delta < 0 || rng.Float64() < temp/(temp+float64(delta)) {
			energy = newEnergy
			if energy < bestEnergy {
				bestEnergy = energy
				copy(bestSpins, spins)
			}
		} else {
			spins[i] *= -1
		}
		temp *= 0.99995
	}

	solutionData := fmt.Sprintf("%v", bestSpins)
	solutionHashBytes := sha256.Sum256([]byte(solutionData))
	solutionHash := hex.EncodeToString(solutionHashBytes[:])

	fmt.Println("\n=== SOLUTION ===")
	fmt.Printf("Best Energy: %d (threshold: %d)\n", bestEnergy, threshold)
	fmt.Printf("Meets Threshold: %v\n", bestEnergy <= threshold)
	fmt.Printf("Solution Hash: %s\n", solutionHash[:32])
}
