package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/bloXroute-Labs/serum-api/bxserum/transaction"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	solanarpc "github.com/gagliardetto/solana-go/rpc"
	sendandconfirmtransaction "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
	solanaws "github.com/gagliardetto/solana-go/rpc/ws"
	"log"
	"os"
)

const (
	rpcEndpoint      = solanarpc.MainNetBeta_RPC
	wsEndpoint       = solanarpc.MainNetBeta_WS
	recipientAddress = "FmZ9kC8bRVsFTgAWrXUyGHp3dN3HtMxJmoi2ijdaYGwi"
)

type txConfirmation struct {
	TxHash string `json:"txHash"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rpcClient := solanarpc.New(rpcEndpoint)
	wsClient, err := solanaws.Connect(ctx, wsEndpoint)
	if err != nil {
		log.Fatal(err)
	}

	recentBlockhash, err := rpcClient.GetRecentBlockhash(ctx, solanarpc.CommitmentFinalized)
	if err != nil {
		log.Fatal(err)
	}

	privateKey := os.Getenv("PRIVATE_KEY")
	if privateKey == "" {
		log.Fatal("env variable `PRIVATE_KEY` not set")
	}

	unsignedTx, err := unsignedTransaction(privateKey, recentBlockhash)
	if err != nil {
		log.Fatal(err)
	}
	unsignedTxBytes, err := partialMarshal(unsignedTx)
	unsignedTxBase64 := base64.StdEncoding.EncodeToString(unsignedTxBytes)

	signedTx, err := transaction.SignTx(unsignedTxBase64)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("tx message signed, ready to send tx")

	signature, err := sendAndConfirmTx(context.Background(), signedTx, rpcClient, wsClient)
	if err != nil {
		log.Fatalf("transaction not sent successfully: %v", err)
	}

	fmt.Printf("tx %s sent and confirmed successfully\n", signature.String())
}

func unsignedTransaction(privateKey string, recentBlockHash *solanarpc.GetRecentBlockhashResult) (*solana.Transaction, error) {
	pKey, err := solana.PrivateKeyFromBase58(privateKey)
	if err != nil {
		return nil, err
	}
	recipient := solana.MustPublicKeyFromBase58(recipientAddress)

	return solana.NewTransaction([]solana.Instruction{
		system.NewTransferInstruction(1, pKey.PublicKey(), recipient).Build(),
	}, recentBlockHash.Value.Blockhash)
}

func partialMarshal(tx *solana.Transaction) ([]byte, error) {
	messageBytes, err := tx.Message.MarshalBinary()
	if err != nil {
		return nil, err
	}

	var signatureCount []byte
	bin.EncodeCompactU16Length(&signatureCount, len(tx.Signatures))

	output := make([]byte, 0, len(signatureCount)+len(signatureCount)*64+len(messageBytes))
	output = append(output, signatureCount...) // signatureCount | signatures | message
	for _, sig := range tx.Signatures {
		output = append(output, sig[:]...)
	}
	output = append(output, messageBytes...)

	return output, nil
}

func sendAndConfirmTx(ctx context.Context, txBase64 string, rpcClient *solanarpc.Client, wsClient *solanaws.Client) (solana.Signature, error) {
	txBytes, err := solanarpc.DataBytesOrJSONFromBase64(txBase64)
	if err != nil {
		return solana.Signature{}, err
	}

	tx, err := solanarpc.TransactionWithMeta{Transaction: txBytes}.GetTransaction()
	if err != nil {
		return solana.Signature{}, err
	}

	return sendandconfirmtransaction.SendAndConfirmTransaction(ctx, rpcClient, wsClient, tx)
}