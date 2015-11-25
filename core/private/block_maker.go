package core

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/logger/glog"
)

type BlockMaker struct {
	chainConfig *core.ChainConfig

	blockchain *core.BlockChain
	mux        *event.TypeMux
	db         ethdb.Database

	address     common.Address
	abi         abi.ABI
	key         *ecdsa.PrivateKey
	selfAddress common.Address

	quit chan struct{} // quit chan
}

func NewBlockMaker(chainConfig *core.ChainConfig, addr common.Address, bc *core.BlockChain, db ethdb.Database, mux *event.TypeMux) *BlockMaker {
	abi, err := abi.JSON(strings.NewReader(definition))
	if err != nil {
		panic(err)
	}

	key, _ := crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")

	bm := &BlockMaker{
		chainConfig: chainConfig,
		blockchain:  bc,
		db:          db,
		address:     addr,
		abi:         abi,
		key:         key,
		selfAddress: crypto.PubkeyToAddress(key.PublicKey),
		mux:         mux,
		quit:        make(chan struct{}),
	}
	go bm.update()

	return bm
}

const blockTime = 5 * time.Second

func (bm *BlockMaker) update() {
	eventSub := bm.mux.Subscribe(core.ChainHeadEvent{})
	eventCh := eventSub.Chan()
	for {
		select {
		case event, ok := <-eventCh:
			if !ok {
				// Event subscription closed, set the channel to nil to stop spinning
				eventCh = nil
				continue
			}

			switch ev := event.Data.(type) {
			case core.ChainHeadEvent:
				fmt.Println(ev)
			}
		case <-bm.quit:
			return
		}
	}
}

func (bm *BlockMaker) Stop() {
	close(bm.quit)
}

func (bm *BlockMaker) createHeader() (*types.Block, *types.Header) {
	canonHash := common.BytesToHash(bm.call("getCanonHash"))
	parent := findDecendant(canonHash, bm.blockchain)

	tstamp := time.Now().Unix()
	if parent.Time().Int64() >= tstamp {
		tstamp++
	}

	return parent, &types.Header{
		ParentHash: parent.Hash(),
		Number:     new(big.Int).Add(parent.Number(), common.Big1),
		Difficulty: core.CalcDifficulty(bm.chainConfig, uint64(tstamp), parent.Time().Uint64(), parent.Number(), parent.Difficulty()),
		GasLimit:   core.CalcGasLimit(parent),
		GasUsed:    new(big.Int),
		Time:       big.NewInt(tstamp),
	}
}

func (bm *BlockMaker) Create(txs types.Transactions) (*types.Block, *state.StateDB) {
	parent, header := bm.createHeader()

	gp := new(core.GasPool).AddGas(header.GasLimit)
	statedb, _ := state.New(parent.Root(), bm.db)
	var receipts types.Receipts

	for i, tx := range txs {
		snap := statedb.Copy()
		receipt, _, _, err := core.ApplyTransaction(bm.chainConfig, bm.blockchain, gp, statedb, header, tx, header.GasUsed, bm.chainConfig.VmConfig)
		if err != nil {
			switch {
			case core.IsGasLimitErr(err):
				from, _ := tx.From()
				glog.Infof("Gas limit reached for (%x) in this block. Continue to try smaller txs\n", from)
			case err != nil:
				glog.Infof("TX (%x) failed, will be removed: %v\n", tx.Hash().Bytes()[:4], err)
			}
			statedb.Set(snap)

			txs = txs[:i]
			break
		}
		receipts = append(receipts, receipt)
	}
	core.AccumulateRewards(statedb, header, nil)
	header.Root = statedb.IntermediateRoot()

	return types.NewBlock(header, txs, nil, receipts), statedb
}

func (bm *BlockMaker) CanonHash() common.Hash {
	return common.BytesToHash(bm.call("getCanonHash"))
}

func (bm *BlockMaker) Vote(hash common.Hash, nonce uint64, key *ecdsa.PrivateKey) (*types.Transaction, error) {
	vote, err := bm.abi.Pack("vote", hash)
	if err != nil {
		return nil, err
	}
	return types.NewTransaction(nonce, bm.address, new(big.Int), big.NewInt(200000), new(big.Int), vote).SignECDSA(key)
}

func (bm *BlockMaker) call(method string, args ...interface{}) []byte {
	input, err := bm.abi.Pack(method, args...)
	if err != nil {
		panic(err)
	}
	return bm.execute(input)
}

func (bm *BlockMaker) execute(input []byte) []byte {
	header := bm.blockchain.CurrentHeader()
	gasLimit := big.NewInt(3141592)
	statedb, _ := state.New(header.Root, bm.db)
	tx, _ := types.NewTransaction(statedb.GetNonce(bm.selfAddress), bm.address, new(big.Int), gasLimit, new(big.Int), input).SignECDSA(bm.key)
	env := core.NewEnv(statedb, bm.chainConfig, bm.blockchain, tx, header, bm.chainConfig.VmConfig)

	ret, _, _ := core.ApplyMessage(env, tx, new(core.GasPool).AddGas(gasLimit))
	return ret
}

func findDecendant(hash common.Hash, blockchain *core.BlockChain) *types.Block {
	block := blockchain.GetBlock(hash)
	// get next in line
	return blockchain.GetBlockByNumber(block.NumberU64() + 1)
}

const definition = `[{"constant":true,"inputs":[{"name":"p","type":"uint256"},{"name":"n","type":"uint256"}],"name":"getEntry","outputs":[{"name":"","type":"bytes32"}],"type":"function"},{"constant":false,"inputs":[{"name":"hash","type":"bytes32"}],"name":"vote","outputs":[],"type":"function"},{"constant":true,"inputs":[{"name":"","type":"address"}],"name":"canVote","outputs":[{"name":"","type":"bool"}],"type":"function"},{"constant":true,"inputs":[],"name":"start","outputs":[{"name":"","type":"uint256"}],"type":"function"},{"constant":true,"inputs":[],"name":"getSize","outputs":[{"name":"","type":"uint256"}],"type":"function"},{"constant":false,"inputs":[{"name":"addr","type":"address"}],"name":"addVoter","outputs":[],"type":"function"},{"constant":true,"inputs":[],"name":"getCanonHash","outputs":[{"name":"","type":"bytes32"}],"type":"function"},{"inputs":[],"type":"constructor"}]`

const MakerCode = `60606040525b600060006001600260005060003373ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548160ff021916908302179055504360016000508190555060006000506000600050805480919060010190908154818355818115116100f2576002028160020283600052602060002091820191016100f19190610096565b808211156100ed57600060018201600050805460008255906000526020600020908101906100e291906100c4565b808211156100de57600081815060009055506001016100c4565b5090565b5b5050600101610096565b5090565b5b505050815481101561000257906000526020600020906002020160005b50915060014303409050816000016000506000828152602001908152602001600020600081815054809291906001019190505550816001016000508054806001018281815481835581811511610197578183600052602060002091820191016101969190610178565b808211156101925760008181506000905550600101610178565b5090565b5b5050509190906000526020600020900160005b83909190915055505b5050610543806101c36000396000f36060604052361561007f576000357c01000000000000000000000000000000000000000000000000000000009004806398ba676d14610081578063a69beaba146100b6578063adfaa72e146100ce578063be9a6555146100fa578063de8fa4311461011d578063f4ab9adf14610140578063f8d11a57146101585761007f565b005b6100a060048080359060200190919080359060200190919050506104d0565b6040518082815260200191505060405180910390f35b6100cc600480803590602001909190505061017b565b005b6100e4600480803590602001909190505061040a565b6040518082815260200191505060405180910390f35b6101076004805050610526565b6040518082815260200191505060405180910390f35b61012a60048050506104bb565b6040518082815260200191505060405180910390f35b610156600480803590602001909190505061042f565b005b610165600480505061031d565b6040518082815260200191505060405180910390f35b6000600061018761052f565b915081600060005080549050111515610236576000600050805480919060010190908154818355818115116102315760020281600202836000526020600020918201910161023091906101d5565b8082111561022c57600060018201600050805460008255906000526020600020908101906102219190610203565b8082111561021d5760008181506000905550600101610203565b5090565b5b50506001016101d5565b5090565b5b505050505b600060005082815481101561000257906000526020600020906002020160005b50905060008160000160005060008581526020019081526020016000206000505414156102ed578060010160005080548060010182818154818355818115116102d1578183600052602060002091820191016102d091906102b2565b808211156102cc57600081815060009055506001016102b2565b5090565b5b5050509190906000526020600020900160005b85909190915055505b8060000160005060008481526020019081526020016000206000818150548092919060010191905055505b505050565b60006000600060006000600050600160006000508054905003815481101561000257906000526020600020906002020160005b509250600090505b82600101600050805490508110156103fc578260000160005060008460010160005083815481101561000257906000526020600020900160005b50548152602001908152602001600020600050548360000160005060008481526020019081526020016000206000505410156103ee578260010160005081815481101561000257906000526020600020900160005b5054915081505b5b8080600101915050610358565b819350610404565b50505090565b600260005060205280600052604060002060009150909054906101000a900460ff1681565b600260005060003373ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900460ff161515610474576104b8565b6001600260005060008373ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548160ff021916908302179055505b50565b600060006000508054905090506104cd565b90565b60006000600060005084815481101561000257906000526020600020906002020160005b5090508060010160005083815481101561000257906000526020600020900160005b5054915061051f565b5092915050565b60016000505481565b600060016000505443039050610540565b9056`
