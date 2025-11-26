package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrInvalidJob         = errorsmod.Register(ModuleName, 1, "invalid job")
	ErrJobNotFound        = errorsmod.Register(ModuleName, 2, "job not found")
	ErrInvalidProof       = errorsmod.Register(ModuleName, 3, "invalid proof")
	ErrProofVerification  = errorsmod.Register(ModuleName, 4, "proof verification failed")
	ErrInsufficientReward = errorsmod.Register(ModuleName, 5, "insufficient reward")
	ErrJobExpired         = errorsmod.Register(ModuleName, 6, "job expired")
	ErrJobNotActive       = errorsmod.Register(ModuleName, 7, "job not active")
	ErrUnauthorized       = errorsmod.Register(ModuleName, 8, "unauthorized")
	ErrInvalidMiner       = errorsmod.Register(ModuleName, 9, "invalid miner")
	ErrNoShares           = errorsmod.Register(ModuleName, 10, "no shares to claim")
	ErrAlreadyClaimed     = errorsmod.Register(ModuleName, 11, "rewards already claimed")
	ErrCheckpointNotFound = errorsmod.Register(ModuleName, 12, "checkpoint not found")
	ErrValidatorNotFound  = errorsmod.Register(ModuleName, 13, "validator not found")
	ErrInvalidParams      = errorsmod.Register(ModuleName, 14, "invalid params")
	ErrCannotCancel       = errorsmod.Register(ModuleName, 15, "cannot cancel job")
)
