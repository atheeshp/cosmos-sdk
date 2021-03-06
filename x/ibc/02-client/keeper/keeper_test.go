package keeper_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/keeper"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

const (
	testClientID  = "gaia"
	testClientID2 = "ethbridge"
	testClientID3 = "ethermint"

	testClientHeight = 5

	trustingPeriod time.Duration = time.Hour * 24 * 7 * 2
	ubdPeriod      time.Duration = time.Hour * 24 * 7 * 3
)

type KeeperTestSuite struct {
	suite.Suite

	cdc            *codec.Codec
	ctx            sdk.Context
	keeper         *keeper.Keeper
	consensusState ibctmtypes.ConsensusState
	header         ibctmtypes.Header
	valSet         *tmtypes.ValidatorSet
	privVal        tmtypes.PrivValidator
	now            time.Time
}

func (suite *KeeperTestSuite) SetupTest() {
	isCheckTx := false
	suite.now = time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)
	now2 := suite.now.Add(time.Hour)
	app := simapp.Setup(isCheckTx)

	suite.cdc = app.Codec()
	suite.ctx = app.BaseApp.NewContext(isCheckTx, abci.Header{Height: testClientHeight, ChainID: testClientID, Time: now2})
	suite.keeper = &app.IBCKeeper.ClientKeeper
	suite.privVal = tmtypes.NewMockPV()
	validator := tmtypes.NewValidator(suite.privVal.GetPubKey(), 1)
	suite.valSet = tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})
	suite.header = ibctmtypes.CreateTestHeader(testClientID, testClientHeight, now2, suite.valSet, []tmtypes.PrivValidator{suite.privVal})
	suite.consensusState = ibctmtypes.ConsensusState{
		Height:       testClientHeight,
		Timestamp:    suite.now,
		Root:         commitmenttypes.NewMerkleRoot([]byte("hash")),
		ValidatorSet: suite.valSet,
	}

	var validators staking.Validators
	for i := 1; i < 11; i++ {
		privVal := tmtypes.NewMockPV()
		pk := privVal.GetPubKey()
		val := staking.NewValidator(sdk.ValAddress(pk.Address()), pk, staking.Description{})
		val.Status = sdk.Bonded
		val.Tokens = sdk.NewInt(rand.Int63())
		validators = append(validators, val)

		app.StakingKeeper.SetHistoricalInfo(suite.ctx, int64(i), staking.NewHistoricalInfo(suite.ctx.BlockHeader(), validators))
	}
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) TestSetClientState() {
	clientState := ibctmtypes.NewClientState(testClientID, trustingPeriod, ubdPeriod, ibctmtypes.Header{})
	suite.keeper.SetClientState(suite.ctx, clientState)

	retrievedState, found := suite.keeper.GetClientState(suite.ctx, testClientID)
	suite.Require().True(found, "GetClientState failed")
	suite.Require().Equal(clientState, retrievedState, "Client states are not equal")
}

func (suite *KeeperTestSuite) TestSetClientType() {
	suite.keeper.SetClientType(suite.ctx, testClientID, exported.Tendermint)
	clientType, found := suite.keeper.GetClientType(suite.ctx, testClientID)

	suite.Require().True(found, "GetClientType failed")
	suite.Require().Equal(exported.Tendermint, clientType, "ClientTypes not stored correctly")
}

func (suite *KeeperTestSuite) TestSetClientConsensusState() {
	suite.keeper.SetClientConsensusState(suite.ctx, testClientID, testClientHeight, suite.consensusState)

	retrievedConsState, found := suite.keeper.GetClientConsensusState(suite.ctx, testClientID, testClientHeight)
	suite.Require().True(found, "GetConsensusState failed")

	tmConsState, ok := retrievedConsState.(ibctmtypes.ConsensusState)
	// recalculate cached totalVotingPower field for equality check
	tmConsState.ValidatorSet.TotalVotingPower()
	suite.Require().True(ok)
	suite.Require().Equal(suite.consensusState, tmConsState, "ConsensusState not stored correctly")
}

func (suite KeeperTestSuite) TestGetAllClients() {
	expClients := []exported.ClientState{
		ibctmtypes.NewClientState(testClientID2, trustingPeriod, ubdPeriod, ibctmtypes.Header{}),
		ibctmtypes.NewClientState(testClientID3, trustingPeriod, ubdPeriod, ibctmtypes.Header{}),
		ibctmtypes.NewClientState(testClientID, trustingPeriod, ubdPeriod, ibctmtypes.Header{}),
	}

	for i := range expClients {
		suite.keeper.SetClientState(suite.ctx, expClients[i])
	}

	clients := suite.keeper.GetAllClients(suite.ctx)
	suite.Require().Len(clients, len(expClients))
	suite.Require().Equal(expClients, clients)
}

func (suite KeeperTestSuite) TestGetConsensusState() {
	suite.ctx = suite.ctx.WithBlockHeight(10)
	cases := []struct {
		name    string
		height  uint64
		expPass bool
	}{
		{"zero height", 0, false},
		{"height > latest height", uint64(suite.ctx.BlockHeight()) + 1, false},
		{"latest height - 1", uint64(suite.ctx.BlockHeight()) - 1, true},
		{"latest height", uint64(suite.ctx.BlockHeight()), true},
	}

	for i, tc := range cases {
		tc := tc
		cs, found := suite.keeper.GetSelfConsensusState(suite.ctx, tc.height)
		if tc.expPass {
			suite.Require().True(found, "Case %d should have passed: %s", i, tc.name)
			suite.Require().NotNil(cs, "Case %d should have passed: %s", i, tc.name)
		} else {
			suite.Require().False(found, "Case %d should have failed: %s", i, tc.name)
			suite.Require().Nil(cs, "Case %d should have failed: %s", i, tc.name)
		}
	}
}

func (suite KeeperTestSuite) TestConsensusStateHelpers() {
	// initial setup
	clientState, _ := ibctmtypes.Initialize(testClientID, trustingPeriod, ubdPeriod, suite.header)
	suite.keeper.SetClientState(suite.ctx, clientState)
	suite.keeper.SetClientConsensusState(suite.ctx, testClientID, testClientHeight, suite.consensusState)

	nextState := ibctmtypes.ConsensusState{
		Height:       testClientHeight + 5,
		Timestamp:    suite.now,
		Root:         commitmenttypes.NewMerkleRoot([]byte("next")),
		ValidatorSet: suite.valSet,
	}

	header := ibctmtypes.CreateTestHeader(testClientID, testClientHeight+5, suite.header.Time.Add(time.Minute), suite.valSet, []tmtypes.PrivValidator{suite.privVal})

	// mock update functionality
	clientState.LastHeader = header
	suite.keeper.SetClientConsensusState(suite.ctx, testClientID, testClientHeight+5, nextState)
	suite.keeper.SetClientState(suite.ctx, clientState)

	latest, ok := suite.keeper.GetLatestClientConsensusState(suite.ctx, testClientID)
	// recalculate cached totalVotingPower for equality check
	latest.(ibctmtypes.ConsensusState).ValidatorSet.TotalVotingPower()
	suite.Require().True(ok)
	suite.Require().Equal(nextState, latest, "Latest client not returned correctly")

	// Should return existing consensusState at latestClientHeight
	lte, ok := suite.keeper.GetClientConsensusStateLTE(suite.ctx, testClientID, testClientHeight+3)
	// recalculate cached totalVotingPower for equality check
	lte.(ibctmtypes.ConsensusState).ValidatorSet.TotalVotingPower()
	suite.Require().True(ok)
	suite.Require().Equal(suite.consensusState, lte, "LTE helper function did not return latest client state below height: %d", testClientHeight+3)
}
