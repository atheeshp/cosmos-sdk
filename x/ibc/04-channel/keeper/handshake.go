package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/capability"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	connectionexported "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"

	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	porttypes "github.com/cosmos/cosmos-sdk/x/ibc/05-port/types"
	commitmentexported "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

// CounterpartyHops returns the connection hops of the counterparty channel.
// The counterparty hops are stored in the inverse order as the channel's.
func (k Keeper) CounterpartyHops(ctx sdk.Context, ch types.Channel) ([]string, bool) {
	counterPartyHops := make([]string, len(ch.ConnectionHops))
	for i, hop := range ch.ConnectionHops {
		connection, found := k.connectionKeeper.GetConnection(ctx, hop)
		if !found {
			return []string{}, false
		}
		counterPartyHops[len(counterPartyHops)-1-i] = connection.GetCounterparty().GetConnectionID()
	}
	return counterPartyHops, true
}

// ChanOpenInit is called by a module to initiate a channel opening handshake with
// a module on another chain.
func (k Keeper) ChanOpenInit(
	ctx sdk.Context,
	order exported.Order,
	connectionHops []string,
	portID,
	channelID string,
	portCap *capability.Capability,
	counterparty types.Counterparty,
	version string,
) (*capability.Capability, error) {
	// channel identifier and connection hop length checked on msg.ValidateBasic()

	_, found := k.GetChannel(ctx, portID, channelID)
	if found {
		return nil, sdkerrors.Wrap(types.ErrChannelExists, channelID)
	}

	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, connectionHops[0])
	if !found {
		return nil, sdkerrors.Wrap(connection.ErrConnectionNotFound, connectionHops[0])
	}

	if connectionEnd.GetState() == connectionexported.UNINITIALIZED {
		return nil, sdkerrors.Wrap(
			connection.ErrInvalidConnectionState,
			"connection state cannot be UNINITIALIZED",
		)
	}

	if !k.portKeeper.Authenticate(ctx, portCap, portID) {
		return nil, sdkerrors.Wrap(porttypes.ErrInvalidPort, "caller does not own port capability")
	}

	channel := types.NewChannel(exported.INIT, order, counterparty, connectionHops, version)
	k.SetChannel(ctx, portID, channelID, channel)

	capKey, err := k.scopedKeeper.NewCapability(ctx, ibctypes.ChannelCapabilityPath(portID, channelID))
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalidChannelCapability, err.Error())
	}
	k.SetNextSequenceSend(ctx, portID, channelID, 1)
	k.SetNextSequenceRecv(ctx, portID, channelID, 1)

	return capKey, nil
}

// ChanOpenTry is called by a module to accept the first step of a channel opening
// handshake initiated by a module on another chain.
func (k Keeper) ChanOpenTry(
	ctx sdk.Context,
	order exported.Order,
	connectionHops []string,
	portID,
	channelID string,
	portCap *capability.Capability,
	counterparty types.Counterparty,
	version,
	counterpartyVersion string,
	proofInit commitmentexported.Proof,
	proofHeight uint64,
) (*capability.Capability, error) {
	// channel identifier and connection hop length checked on msg.ValidateBasic()

	previousChannel, found := k.GetChannel(ctx, portID, channelID)
	if found && !(previousChannel.State == exported.INIT &&
		previousChannel.Ordering == order &&
		previousChannel.Counterparty.PortID == counterparty.PortID &&
		previousChannel.Counterparty.ChannelID == counterparty.ChannelID &&
		previousChannel.ConnectionHops[0] == connectionHops[0] &&
		previousChannel.Version == version) {
		sdkerrors.Wrap(types.ErrInvalidChannel, "cannot relay connection attempt")
	}

	if !k.portKeeper.Authenticate(ctx, portCap, portID) {
		return nil, sdkerrors.Wrap(porttypes.ErrInvalidPort, "caller does not own port capability")
	}

	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, connectionHops[0])
	if !found {
		return nil, sdkerrors.Wrap(connection.ErrConnectionNotFound, connectionHops[0])
	}

	if connectionEnd.GetState() != connectionexported.OPEN {
		return nil, sdkerrors.Wrapf(
			connection.ErrInvalidConnectionState,
			"connection state is not OPEN (got %s)", connectionEnd.GetState().String(),
		)
	}

	// NOTE: this step has been switched with the one below to reverse the connection
	// hops
	channel := types.NewChannel(exported.TRYOPEN, order, counterparty, connectionHops, version)

	counterpartyHops, found := k.CounterpartyHops(ctx, channel)
	if !found {
		// should not reach here, connectionEnd was able to be retrieved above
		panic("cannot find connection")
	}

	// expectedCounterpaty is the counterparty of the counterparty's channel end
	// (i.e self)
	expectedCounterparty := types.NewCounterparty(portID, channelID)
	expectedChannel := types.NewChannel(
		exported.INIT, channel.Ordering, expectedCounterparty,
		counterpartyHops, counterpartyVersion,
	)

	if err := k.connectionKeeper.VerifyChannelState(
		ctx, connectionEnd, proofHeight, proofInit,
		counterparty.PortID, counterparty.ChannelID, expectedChannel,
	); err != nil {
		return nil, err
	}

	k.SetChannel(ctx, portID, channelID, channel)

	capKey, err := k.scopedKeeper.NewCapability(ctx, ibctypes.ChannelCapabilityPath(portID, channelID))
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrInvalidChannelCapability, err.Error())
	}
	k.SetNextSequenceSend(ctx, portID, channelID, 1)
	k.SetNextSequenceRecv(ctx, portID, channelID, 1)

	return capKey, nil
}

// ChanOpenAck is called by the handshake-originating module to acknowledge the
// acceptance of the initial request by the counterparty module on the other chain.
func (k Keeper) ChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID string,
	chanCap *capability.Capability,
	counterpartyVersion string,
	proofTry commitmentexported.Proof,
	proofHeight uint64,
) error {
	channel, found := k.GetChannel(ctx, portID, channelID)
	if !found {
		return sdkerrors.Wrap(types.ErrChannelNotFound, channelID)
	}

	if !(channel.State == exported.INIT || channel.State == exported.TRYOPEN) {
		return sdkerrors.Wrapf(
			types.ErrInvalidChannelState,
			"channel state should be INIT or TRYOPEN (got %s)", channel.State.String(),
		)
	}

	if !k.scopedKeeper.AuthenticateCapability(ctx, chanCap, ibctypes.ChannelCapabilityPath(portID, channelID)) {
		return sdkerrors.Wrap(types.ErrChannelCapabilityNotFound, "caller does not own capability for channel")
	}

	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return sdkerrors.Wrap(connection.ErrConnectionNotFound, channel.ConnectionHops[0])
	}

	if connectionEnd.GetState() != connectionexported.OPEN {
		return sdkerrors.Wrapf(
			connection.ErrInvalidConnectionState,
			"connection state is not OPEN (got %s)", connectionEnd.GetState().String(),
		)
	}

	counterpartyHops, found := k.CounterpartyHops(ctx, channel)
	if !found {
		// should not reach here, connectionEnd was able to be retrieved above
		panic("cannot find connection")
	}

	// counterparty of the counterparty channel end (i.e self)
	counterparty := types.NewCounterparty(portID, channelID)
	expectedChannel := types.NewChannel(
		exported.TRYOPEN, channel.Ordering, counterparty,
		counterpartyHops, counterpartyVersion,
	)

	if err := k.connectionKeeper.VerifyChannelState(
		ctx, connectionEnd, proofHeight, proofTry,
		channel.Counterparty.PortID, channel.Counterparty.ChannelID,
		expectedChannel,
	); err != nil {
		return err
	}

	channel.State = exported.OPEN
	channel.Version = counterpartyVersion
	k.SetChannel(ctx, portID, channelID, channel)

	return nil
}

// ChanOpenConfirm is called by the counterparty module to close their end of the
//  channel, since the other end has been closed.
func (k Keeper) ChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
	chanCap *capability.Capability,
	proofAck commitmentexported.Proof,
	proofHeight uint64,
) error {
	channel, found := k.GetChannel(ctx, portID, channelID)
	if !found {
		return sdkerrors.Wrap(types.ErrChannelNotFound, channelID)
	}

	if channel.State != exported.TRYOPEN {
		return sdkerrors.Wrapf(
			types.ErrInvalidChannelState,
			"channel state is not TRYOPEN (got %s)", channel.State.String(),
		)
	}

	if !k.scopedKeeper.AuthenticateCapability(ctx, chanCap, ibctypes.ChannelCapabilityPath(portID, channelID)) {
		return sdkerrors.Wrap(types.ErrChannelCapabilityNotFound, "caller does not own capability for channel")
	}

	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return sdkerrors.Wrap(connection.ErrConnectionNotFound, channel.ConnectionHops[0])
	}

	if connectionEnd.GetState() != connectionexported.OPEN {
		return sdkerrors.Wrapf(
			connection.ErrInvalidConnectionState,
			"connection state is not OPEN (got %s)", connectionEnd.GetState().String(),
		)
	}

	counterpartyHops, found := k.CounterpartyHops(ctx, channel)
	if !found {
		// Should not reach here, connectionEnd was able to be retrieved above
		panic("cannot find connection")
	}

	counterparty := types.NewCounterparty(portID, channelID)
	expectedChannel := types.NewChannel(
		exported.OPEN, channel.Ordering, counterparty,
		counterpartyHops, channel.Version,
	)

	if err := k.connectionKeeper.VerifyChannelState(
		ctx, connectionEnd, proofHeight, proofAck,
		channel.Counterparty.PortID, channel.Counterparty.ChannelID,
		expectedChannel,
	); err != nil {
		return err
	}

	channel.State = exported.OPEN
	k.SetChannel(ctx, portID, channelID, channel)

	return nil
}

// Closing Handshake
//
// This section defines the set of functions required to close a channel handshake
// as defined in https://github.com/cosmos/ics/tree/master/spec/ics-004-channel-and-packet-semantics#closing-handshake

// ChanCloseInit is called by either module to close their end of the channel. Once
// closed, channels cannot be reopened.
func (k Keeper) ChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
	chanCap *capability.Capability,
) error {
	if !k.scopedKeeper.AuthenticateCapability(ctx, chanCap, ibctypes.ChannelCapabilityPath(portID, channelID)) {
		return sdkerrors.Wrap(types.ErrChannelCapabilityNotFound, "caller does not own capability for channel")
	}

	channel, found := k.GetChannel(ctx, portID, channelID)
	if !found {
		return sdkerrors.Wrap(types.ErrChannelNotFound, channelID)
	}

	if channel.State == exported.CLOSED {
		return sdkerrors.Wrap(types.ErrInvalidChannelState, "channel is already CLOSED")
	}

	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return sdkerrors.Wrap(connection.ErrConnectionNotFound, channel.ConnectionHops[0])
	}

	if connectionEnd.GetState() != connectionexported.OPEN {
		return sdkerrors.Wrapf(
			connection.ErrInvalidConnectionState,
			"connection state is not OPEN (got %s)", connectionEnd.GetState().String(),
		)
	}

	channel.State = exported.CLOSED
	k.SetChannel(ctx, portID, channelID, channel)
	k.Logger(ctx).Info("channel close initialized: portID (%s), channelID (%s)", portID, channelID)
	return nil
}

// ChanCloseConfirm is called by the counterparty module to close their end of the
// channel, since the other end has been closed.
func (k Keeper) ChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
	chanCap *capability.Capability,
	proofInit commitmentexported.Proof,
	proofHeight uint64,
) error {
	if !k.scopedKeeper.AuthenticateCapability(ctx, chanCap, ibctypes.ChannelCapabilityPath(portID, channelID)) {
		return sdkerrors.Wrap(types.ErrChannelCapabilityNotFound, "caller does not own capability for channel")
	}

	channel, found := k.GetChannel(ctx, portID, channelID)
	if !found {
		return sdkerrors.Wrap(types.ErrChannelNotFound, channelID)
	}

	if channel.State == exported.CLOSED {
		return sdkerrors.Wrap(types.ErrInvalidChannelState, "channel is already CLOSED")
	}

	connectionEnd, found := k.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return sdkerrors.Wrap(connection.ErrConnectionNotFound, channel.ConnectionHops[0])
	}

	if connectionEnd.GetState() != connectionexported.OPEN {
		return sdkerrors.Wrapf(
			connection.ErrInvalidConnectionState,
			"connection state is not OPEN (got %s)", connectionEnd.GetState().String(),
		)
	}

	counterpartyHops, found := k.CounterpartyHops(ctx, channel)
	if !found {
		// Should not reach here, connectionEnd was able to be retrieved above
		panic("cannot find connection")
	}

	counterparty := types.NewCounterparty(portID, channelID)
	expectedChannel := types.NewChannel(
		exported.CLOSED, channel.Ordering, counterparty,
		counterpartyHops, channel.Version,
	)

	if err := k.connectionKeeper.VerifyChannelState(
		ctx, connectionEnd, proofHeight, proofInit,
		channel.Counterparty.PortID, channel.Counterparty.ChannelID,
		expectedChannel,
	); err != nil {
		return err
	}

	channel.State = exported.CLOSED
	k.SetChannel(ctx, portID, channelID, channel)

	return nil
}
