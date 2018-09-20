package blockchain

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	mrand "math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	//"github.com/davecgh/go-spew/spew"
	libp2p "github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p-host"
	net "github.com/libp2p/go-libp2p-net"
	ma "github.com/multiformats/go-multiaddr"
)

var WalletSuffix string

// Block represents each 'item' in the blockchain
type Block struct {
	Index     int `json:"index"`
	Timestamp string `json:"timestamp"`
	Result       int `json:"result"`
	Hash      string `json:"hash"`
	PrevHash  string `json:"prevhash"`
	Proof        uint64           `json:"proof"`
	Transactions []Transaction `json:"transactions"`
}


type Transaction struct {
	Amount    uint64    `json:"amount"`
	Recipient string `json:"recipient"`
	Sender    string `json:"sender"`
	Data      []byte `json:"data"`
}

type TxPool struct {
	AllTx     []Transaction
}

func NewTxPool() *TxPool {
	return &TxPool{
		AllTx:   make([]Transaction, 0),
	}
}


func (p *TxPool)Clear() bool {
	if len(p.AllTx) == 0 {
		return true
	}
	p.AllTx = make([]Transaction, 0)
	return true
}

// Blockchain is a series of validated Blocks
type Blockchain struct {
	Blocks []Block
	TxPool *TxPool
}

func (t *Blockchain) NewTransaction(sender string, recipient string, amount uint64, data []byte) *Transaction {
	transaction := new(Transaction)
	transaction.Sender = sender
	transaction.Recipient = recipient
	transaction.Amount = amount
	transaction.Data = data

	return transaction
}

func (t *Blockchain)AddTxPool(tx *Transaction) int {
	t.TxPool.AllTx = append(t.TxPool.AllTx, *tx)
	return len(t.TxPool.AllTx)
}

func (t *Blockchain) LastBlock() Block {
	return t.Blocks[len(t.Blocks)-1]
}

var BlockchainInstance Blockchain = Blockchain{
	TxPool : NewTxPool(),
}

var mutex = &sync.Mutex{}


func Lock(){
	mutex.Lock()
}

func UnLock(){
	mutex.Unlock()
}

// makeBasicHost creates a LibP2P host with a random peer ID listening on the
// given multiaddress. It will use secio if secio is true.
func MakeBasicHost(listenPort int, secio bool, randseed int64) (host.Host, error) {

	// If the seed is zero, use real cryptographic randomness. Otherwise, use a
	// deterministic randomness source to make generated keys stay the same
	// across multiple runs
	var r io.Reader
	if randseed == 0 {
		r = rand.Reader
	} else {
		r = mrand.New(mrand.NewSource(randseed))
	}

	// Generate a key pair for this host. We will use it
	// to obtain a valid host ID.
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		return nil, err
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", listenPort)),
		libp2p.Identity(priv),
	}

	basicHost, err := libp2p.New(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	// Build host multiaddress
	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", basicHost.ID().Pretty()))

	// Now we can build a full multiaddress to reach this host
	// by encapsulating both addresses:
	addr := basicHost.Addrs()[0]
	fullAddr := addr.Encapsulate(hostAddr)
	log.Printf("I am %s\n", fullAddr)
	if secio {
		log.Printf("Now run \"go run main.go -l %d -d %s -secio\" on a different terminal\n", listenPort+2, fullAddr)
	} else {
		log.Printf("Now run \"go run main.go -l %d -d %s\" on a different terminal\n", listenPort+2, fullAddr)
	}

	return basicHost, nil
}

func HandleStream(s net.Stream) {

	log.Println("Got a new stream!")

	// Create a buffer stream for non blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	go ReadData(rw)
	go WriteData(rw)

	// stream 's' will stay open until you close it (or the other side closes it).
}

func ReadData(rw *bufio.ReadWriter) {

	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		if str == "" {
			return
		}
		if str != "\n" {

			chain := make([]Block, 0)
			if err := json.Unmarshal([]byte(str), &chain); err != nil {
				log.Fatal(err)
			}

			mutex.Lock()
			if len(chain) > len(BlockchainInstance.Blocks) {
				BlockchainInstance.Blocks = chain
				bytes, err := json.MarshalIndent(BlockchainInstance.Blocks, "", "  ")
				if err != nil {

					log.Fatal(err)
				}
				// Green console color: 	\x1b[32m
				// Reset console color: 	\x1b[0m
				fmt.Printf("\x1b[32m%s\x1b[0m> ", string(bytes))
			}
			mutex.Unlock()
		}
	}
}

func WriteData(rw *bufio.ReadWriter) {

	go func() {
		for {
			time.Sleep(5 * time.Second)
			mutex.Lock()
			bytes, err := json.Marshal(BlockchainInstance.Blocks)
			if err != nil {
				log.Println(err)
			}
			mutex.Unlock()

			mutex.Lock()
			rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
			rw.Flush()
			mutex.Unlock()

		}
	}()

	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		sendData = strings.Replace(sendData, "\n", "", -1)
		_result, err := strconv.Atoi(sendData)
		if err != nil {
			log.Fatal(err)
		}
		newBlock := GenerateBlock(BlockchainInstance.Blocks[len(BlockchainInstance.Blocks)-1], _result)

		if len(BlockchainInstance.TxPool.AllTx) > 0 {
			// todo 添加账户系统后的转账操作，现不做任何操作，仅将未打包交易打包到块中
			newBlock.Transactions = BlockchainInstance.TxPool.AllTx

			BlockchainInstance.TxPool.Clear()
		}

		if IsBlockValid(newBlock, BlockchainInstance.Blocks[len(BlockchainInstance.Blocks)-1]) {
			mutex.Lock()
			BlockchainInstance.Blocks = append(BlockchainInstance.Blocks, newBlock)
			mutex.Unlock()
		}

		bytes, err := json.Marshal(BlockchainInstance.Blocks)
		if err != nil {
			log.Println(err)
		}

		//spew.Dump(BlockchainInstance.Blocks)

		mutex.Lock()
		rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
		rw.Flush()
		mutex.Unlock()
	}

}



// make sure block is valid by checking index, and comparing the hash of the previous block
func IsBlockValid(newBlock, oldBlock Block) bool {
	if oldBlock.Index+1 != newBlock.Index {
		return false
	}

	if oldBlock.Hash != newBlock.PrevHash {
		return false
	}

	if CalculateHash(newBlock) != newBlock.Hash {
		return false
	}

	return true
}

// SHA256 hashing
func CalculateHash(block Block) string {
	record := strconv.Itoa(block.Index) + block.Timestamp + strconv.Itoa(block.Result) + block.PrevHash
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

// create a new block using previous block's hash
func GenerateBlock(oldBlock Block, Result int) Block {

	var newBlock Block

	t := time.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.Result = Result
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Hash = CalculateHash(newBlock)

	return newBlock
}
