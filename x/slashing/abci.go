package slashing

import (
	"context"
	"errors"

	"cosmossdk.io/core/comet"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// BeginBlocker check for infraction evidence or downtime of validators
// on every begin block
func BeginBlocker(ctx context.Context, k keeper.Keeper) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, telemetry.Now(), telemetry.MetricKeyBeginBlocker)

	// Iterate over all the validators which *should* have signed this block
	// store whether or not they have actually signed it and slash/unbond any
	// which have missed too many blocks in a row (downtime slashing)
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	for _, voteInfo := range sdkCtx.VoteInfos() {
		err := k.HandleValidatorSignature(ctx, voteInfo.Validator.Address, voteInfo.Validator.Power, comet.BlockIDFlag(voteInfo.BlockIdFlag))
		if err != nil {
			if errors.Is(err, stakingtypes.ErrNoValidatorFound) {
				k.Logger(ctx).Error(
					"skipping slashing for validator missing from staking state",
					"cons_addr", sdk.ConsAddress(voteInfo.Validator.Address).String(),
					"power", voteInfo.Validator.Power,
					"block_id_flag", voteInfo.BlockIdFlag,
					"error", err,
				)
				continue
			}

			return err
		}
	}
	return nil
}
