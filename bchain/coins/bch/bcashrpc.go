package bch

import (
	"blockbook/bchain"
	"blockbook/bchain/coins/btc"
	"encoding/hex"
	"encoding/json"

	"github.com/cpacia/bchutil"
	"github.com/golang/glog"
	"github.com/juju/errors"
)

// BCashRPC is an interface to JSON-RPC bitcoind service.
type BCashRPC struct {
	*btc.BitcoinRPC
}

// NewBCashRPC returns new BCashRPC instance.
func NewBCashRPC(config json.RawMessage, pushHandler func(bchain.NotificationType)) (bchain.BlockChain, error) {
	b, err := btc.NewBitcoinRPC(config, pushHandler)
	if err != nil {
		return nil, err
	}

	s := &BCashRPC{
		b.(*btc.BitcoinRPC),
	}

	return s, nil
}

// Initialize initializes BCashRPC instance.
func (b *BCashRPC) Initialize() error {
	chainName, err := b.GetChainInfoAndInitializeMempool(b)
	if err != nil {
		return err
	}

	params := GetChainParams(chainName)

	// always create parser
	b.Parser, err = NewBCashParser(params, b.ChainConfig)

	if err != nil {
		return err
	}

	// parameters for getInfo request
	if params.Net == bchutil.MainnetMagic {
		b.Testnet = false
		b.Network = "livenet"
	} else {
		b.Testnet = true
		b.Network = "testnet"
	}

	glog.Info("rpc: block chain ", params.Name)

	return nil
}

// getblock

type cmdGetBlock struct {
	Method string `json:"method"`
	Params struct {
		BlockHash string `json:"blockhash"`
		Verbose   bool   `json:"verbose"`
	} `json:"params"`
}

// estimatesmartfee

type cmdEstimateSmartFee struct {
	Method string `json:"method"`
	Params struct {
		Blocks int `json:"nblocks"`
	} `json:"params"`
}

// GetBlock returns block with given hash.
func (b *BCashRPC) GetBlock(hash string, height uint32) (*bchain.Block, error) {
	var err error
	if hash == "" && height > 0 {
		hash, err = b.GetBlockHash(height)
		if err != nil {
			return nil, err
		}
	}
	header, err := b.GetBlockHeader(hash)
	if err != nil {
		return nil, err
	}
	data, err := b.GetBlockRaw(hash)
	if err != nil {
		return nil, err
	}
	block, err := b.Parser.ParseBlock(data)
	if err != nil {
		return nil, errors.Annotatef(err, "hash %v", hash)
	}
	block.BlockHeader = *header
	return block, nil
}

// GetBlockRaw returns block with given hash as bytes.
func (b *BCashRPC) GetBlockRaw(hash string) ([]byte, error) {
	glog.V(1).Info("rpc: getblock (verbose=0) ", hash)

	res := btc.ResGetBlockRaw{}
	req := cmdGetBlock{Method: "getblock"}
	req.Params.BlockHash = hash
	req.Params.Verbose = false
	err := b.Call(&req, &res)

	if err != nil {
		return nil, errors.Annotatef(err, "hash %v", hash)
	}
	if res.Error != nil {
		if isErrBlockNotFound(res.Error) {
			return nil, bchain.ErrBlockNotFound
		}
		return nil, errors.Annotatef(res.Error, "hash %v", hash)
	}
	return hex.DecodeString(res.Result)
}

// GetBlockFull returns block with given hash.
func (b *BCashRPC) GetBlockFull(hash string) (*bchain.Block, error) {
	return nil, errors.New("Not implemented")
}

// EstimateSmartFee returns fee estimation.
func (b *BCashRPC) EstimateSmartFee(blocks int, conservative bool) (float64, error) {
	glog.V(1).Info("rpc: estimatesmartfee ", blocks)

	res := btc.ResEstimateSmartFee{}
	req := cmdEstimateSmartFee{Method: "estimatesmartfee"}
	req.Params.Blocks = blocks
	// conservative param is omitted
	err := b.Call(&req, &res)

	if err != nil {
		return 0, err
	}
	if res.Error != nil {
		return 0, res.Error
	}
	return res.Result.Feerate, nil
}

func isErrBlockNotFound(err *bchain.RPCError) bool {
	return err.Message == "Block not found" ||
		err.Message == "Block height out of range"
}
