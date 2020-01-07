// Copyright 2015 The go-ethereum Authors
// Copyright 2019 The Ether-1 Development Team
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package nodeprotocol

import (
	"fmt"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
)

type NodeValidations struct {
	Id          []byte            `json:"id"`
	Validations [][]byte          `json:"validations"`
}

func CheckNextRewardedNode(nodeId string, address common.Address) bool {
	selfNodeKey := ActiveNode().Server().Config.PrivateKey
	selfNodeId :=  GetNodeId(ActiveNode().Server().Self())
	log.Info("Retrieving Node Key", "Key", selfNodeKey)
	if nodeId == selfNodeId {
		return true
	}
	return false
}

func SendValidationTx(nodeId string, nodeIp string, nodeAddress common.Address) {

}

func CheckValidNodeProtocolTx(state *state.StateDB, currentBlock *types.Block, from common.Address, to *common.Address, data []byte) bool {
	if currentBlock.Header().Number.Int64() >= params.NodeProtocolBlock {
		log.Warn("Verifying Validity of Node Protocol Tx", "To", to, "From", from, "Number", currentBlock.NumberU64())
		for _, nodeType := range params.NodeTypes {
			if *to == nodeType.TxAddress {
				/*if CheckNodeCandidate(state, from) {
					log.Warn("Node Protocol Tx Validation Complete", "Valid", "True")
					return true*/
				if from == common.HexToAddress("0x96216849c49358B10257cb55b28eA603c874b05E") { // for testing
					log.Warn("Node Protocol Tx Validation Complete (Test/Debug)", "Valid", "True")
					return true
				}
			}
		}
	}
	log.Error("Node Protocol Tx Validation Complete", "Valid", "False")
	return false
}

// SignNodeProtocolValidation is used to respond to a peer/next node's validation request
// A signed validation using enode private key signals an unequivocal validation of activity
func SignNodeProtocolValidation(privateKey *ecdsa.PrivateKey, nodeId []byte) []byte {
	hash := crypto.Keccak256(nodeId)
        signedValidation, err := crypto.Sign(hash, privateKey)
        if err != nil {
		log.Error("Error", "Error", err)
        }
	return signedValidation
}

// ValidateNodeProtocolSignature is used to verify validation signatures when a node validation tx
// is recevied to decentrally validate a nodes activity
func ValidateNodeProtocolSignature(nodeId []byte, signedValidation []byte) bool {
	recoveredPub, err := crypto.Ecrecover(crypto.Keccak256(nodeId), signedValidation)
	if err != nil {
		log.Error("Error", "Error", err)
	}
	recoveredId, _ := crypto.UnmarshalPubkey(recoveredPub)
	recoveredIdString := fmt.Sprintf("%x", crypto.FromECDSAPub(recoveredId)[1:])
	recoveredAddr := crypto.PubkeyToAddress(*recoveredId)

	log.Info("Recovered Address", "Address", recoveredId)
	fmt.Println("Recovered ID: " + recoveredIdString)
	fmt.Println("Recovered Address: " + recoveredAddr.String())
	return false
}

func SendSignedNodeProtocolTx(privateKey *ecdsa.PrivateKey, validations NodeValidations) *types.Transaction {
	client, err := ethclient.Dial("/home/nucleos/.xerom/geth.ipc")
	if err != nil {
		log.Error("Error", "Error", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Error("Error", "Error", "cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	from := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), from)
	if err != nil {
		log.Error("Error", "Error", err)
	}

	value := big.NewInt(0)
	gasLimit := uint64(8000000)
	gasPrice := big.NewInt(0)
	to := common.HexToAddress("0x0000000000000000000000000000000000001000")
	data, err := json.Marshal(validations)
	if err != nil {
		log.Error("Error", "Error", err)
	}

	tx := types.NewTransaction(nonce, to, value, gasLimit, gasPrice, data)
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Error("Error", "Error", err)
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Error("Error", "Error", err)
	}

	//ts := types.Transactions{signedTx}
	//rawTxBytes := ts.GetRlp(0)
	//rawTxHex := hex.EncodeToString(rawTxBytes)

	//fmt.Printf(rawTxHex) // f86...772

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Error("Error", "Error", err)
	}
	//fmt.Printf("\nTx Sent: %s", signedTx.Hash().Hex())
	return signedTx
}
