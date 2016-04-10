package core

import (
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/event"
)

func TestCreation(t *testing.T) {
	var (
		db, _              = ethdb.NewMemDatabase()
		key1, _            = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
		addr1              = crypto.PubkeyToAddress(key1.PublicKey)
		addr1Nonce  uint64 = 0
		key2, _            = crypto.HexToECDSA("c71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
		addr2              = crypto.PubkeyToAddress(key2.PublicKey)
		addr2Nonce  uint64 = 0
		key3, _            = crypto.HexToECDSA("d71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
		addr3              = crypto.PubkeyToAddress(key3.PublicKey)
		addr3Nonce  uint64 = 0
		code               = common.Hex2Bytes(MakerCode)
		chainConfig        = &core.ChainConfig{HomesteadBlock: new(big.Int)}
		//gasLimit          = big.NewInt(260000)
		makerAddress = crypto.CreateAddress(addr1, 0)
	)
	defer db.Close()

	genesis := core.WriteGenesisBlockForTesting(db,
		core.GenesisAccount{addr1, big.NewInt(1000000), nil},
		core.GenesisAccount{addr2, big.NewInt(1000000), nil},
		core.GenesisAccount{addr3, big.NewInt(1000000), nil},
	)

	evmux := &event.TypeMux{}
	blockchain, err := core.NewBlockChain(db, chainConfig, core.FakePow{}, evmux)
	if err != nil {
		t.Fatal(err)
	}

	makerStart := blockchain.CurrentBlock().NumberU64() + 1
	t.Logf("initial round: maker genesis starts at %d\n", makerStart)
	chain, _ := core.GenerateChain(genesis, db, 1, func(i int, gen *core.BlockGen) {
		gen.Bc = blockchain
		tx, _ := types.NewContractCreation(gen.TxNonce(addr1), new(big.Int), big.NewInt(1000000), new(big.Int), code).SignECDSA(key1)
		addr1Nonce++

		gen.AddTx(tx)
	})
	if i, err := blockchain.InsertChain(chain); err != nil {
		t.Fatalf("insert error (block %d): %v\n", i, err)
		return
	}
	t.Logf("inserted block %x\n", chain[len(chain)-1].Hash())
	t.Logf("maker addr:%x\n", makerAddress)

	maker := NewBlockMaker(chainConfig, makerAddress, blockchain, db, &event.TypeMux{})

	t.Log(maker.abi.Methods["start"])

	var contractStart *big.Int
	maker.abi.Unpack(&contractStart, "start", maker.call("start"))

	if contractStart.Uint64() != makerStart {
		t.Fatalf("maker genesis start did not match contract. Expected %d, got %d", makerStart, contractStart)
	}

	parentHash := findDecendant(maker.CanonHash(), blockchain).Hash()

	vote1, _ := maker.Vote(parentHash, addr2Nonce, key2)
	addr2Nonce++
	block, _ := maker.Create(types.Transactions{vote1})
	chain1 := types.Blocks{block}
	if i, err := blockchain.InsertChain(chain1); err != nil {
		t.Fatalf("insert error (block %d): %v\n", i, err)

		return
	}

	vote2, _ := maker.Vote(parentHash, addr1Nonce, key1)
	addr1Nonce++
	block, _ = maker.Create(types.Transactions{vote2})
	chain2 := types.Blocks{block}
	if i, err := blockchain.InsertChain(chain2); err != nil {
		t.Fatalf("insert error (block %d): %v\n", i, err)
		return
	}

	t.Logf("voting on hash (2x): %x\n", chain1[0].Hash())
	vote3, _ := maker.Vote(chain1[0].Hash(), addr1Nonce, key1)
	vote4, _ := maker.Vote(chain1[0].Hash(), addr2Nonce, key2)
	winnerHash := chain1[0].Hash()

	t.Logf("voting on hash (1x): %x\n", chain2[0].Hash())
	vote5, _ := maker.Vote(chain2[0].Hash(), addr3Nonce, key3)
	addr1Nonce++
	addr2Nonce++
	addr3Nonce++

	block, _ = maker.Create(types.Transactions{vote3, vote4, vote5})
	if i, err := blockchain.InsertChain(types.Blocks{block}); err != nil {
		t.Fatalf("insert error (block %d): %v\n", i, err)
		return
	}

	// manual verification of the correct block
	block, _ = maker.Create(nil)
	if winnerHash != block.ParentHash() {
		t.Errorf("expected %x to be canonical, got %x", winnerHash, block.ParentHash())
	}
	time.Sleep(2 * time.Second)
}
