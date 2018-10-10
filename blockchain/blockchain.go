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
	"encoding/gob"
	//"github.com/davecgh/go-spew/spew"
	"path/filepath"
)

var WalletSuffix string
var DataFileName string = "chainData.txt"
const difficulty = 1

// Block represents each 'item' in the blockchain
type Block struct {
	Index     int `json:"index"`
	Timestamp string `json:"timestamp"`
	Result       int `json:"result"`
	Hash      string `json:"hash"`
	PrevHash  string `json:"prevhash"`
	Proof        uint64           `json:"proof"`
	Transactions []Transaction `json:"transactions"`
	Accounts   map[string]Account  `json:"accounts"`
	Difficulty int
	Nonce      string
	Validator string
}


type Account struct {
	Balance uint64 `json:"balance"`
	State   uint64 `json:"state"`
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
	DataDir string
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

func (t *Blockchain) GetBalance(address string) uint64 {
	accounts := t.LastBlock().Accounts
	if value, ok := accounts[address]; ok {
		return value.Balance
	}
	return 0
}


func (t *Blockchain)PackageTx(newBlock *Block) {
	(*newBlock).Transactions = t.TxPool.AllTx
	AccountsMap := t.LastBlock().Accounts
	for k1, v1 := range AccountsMap {
		fmt.Println(k1, "--", v1)
	}

	unusedTx := make([]Transaction,0)

	for _, v := range t.TxPool.AllTx{
		if value, ok := AccountsMap[v.Sender]; ok {
			if value.Balance < v.Amount{
				unusedTx = append(unusedTx, v)
				continue
			}
			value.Balance -= v.Amount
			value.State += 1
			AccountsMap[v.Sender] = value
		}

		if value, ok := AccountsMap[v.Recipient]; ok {
			value.Balance += v.Amount
			AccountsMap[v.Recipient] = value
		}else {

			newAccount := new(Account)
			newAccount.Balance = v.Amount
			newAccount.State = 0
			AccountsMap[v.Recipient] = *newAccount
		}
	}

    t.TxPool.Clear()
    //余额不够的交易放回交易池
    if len(unusedTx) > 0 {
		for _, v := range unusedTx{
			t.AddTxPool(&v)
		}
	}

	(*newBlock).Accounts = AccountsMap
}


func (t *Blockchain)WriteDate2File() {
	if t.DataDir == "" {
		return
	}

	joinPath := filepath.Join(t.DataDir, DataFileName)

	file,err := os.OpenFile(joinPath,os.O_WRONLY|os.O_CREATE,0755)  //以写方式打开文件
	if err != nil {
		log.Println("can't write data to file, open file fail err:",err)
		return
	}
	defer file.Close()
	enc := gob.NewEncoder(file)

	if err := enc.Encode(t); err != nil {
		log.Fatal("encode error:", err)
	}

	fmt.Println()
	fmt.Printf("\n%sfile:%s\n>", "已配置数据存储目录，写入当前数据到存储目录中.", filepath.Join(t.DataDir,DataFileName))
}

func (t *Blockchain)ReadDataFromFile() {
	if t.DataDir == "" {
		return
	}

	joinPath := filepath.Join(t.DataDir, DataFileName)

	if !IsExist(joinPath) {
		return
	}

	file,err := os.Open(joinPath)  //以写方式打开文件
	if err != nil {
		log.Println("can't read data from file, open file fail err:",err)
		return
	}
	defer file.Close()
	dec := gob.NewDecoder(file)

	var blockchainInstanceFromFile Blockchain
	if err := dec.Decode(&blockchainInstanceFromFile); err != nil {
		log.Fatal("decode error:", err)
	}

	BlockchainInstance = blockchainInstanceFromFile
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
func MakeBasicHost(listenPort int, secio bool, randseed int64, initAccount string) (host.Host, error) {

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
		if initAccount != "" {
			log.Printf("Now run \"go run main.go \x1b[32m -c chain -l %d -a %s -d %s -secio\x1b[0m\" on a different terminal\n", listenPort+2, initAccount, fullAddr)
		}else {
			log.Printf("Now run \"go run main.go \x1b[32m -c chain -l %d -d %s -secio\x1b[0m\" on a different terminal\n", listenPort+2, fullAddr)
		}
	} else {
		if initAccount != "" {
			log.Printf("Now run \"go run main.go \x1b[32m -c chain -l %d -a %s -d %s\x1b[0m\" on a different terminal\n", listenPort+2, initAccount, fullAddr)
		}else {
			log.Printf("Now run \"go run main.go \x1b[32m -c chain -l %d -d %s\x1b[0m\" on a different terminal\n", listenPort+2, fullAddr)
		}
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

				BlockchainInstance.WriteDate2File()
			}
			mutex.Unlock()
		}
	}
}

func WriteData(rw *bufio.ReadWriter) {
	//pos 内容
	//go pos()
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

		address := GenPosAddress()
		newBlock := GenerateBlock(BlockchainInstance.Blocks[len(BlockchainInstance.Blocks)-1], _result,address)

		if len(BlockchainInstance.TxPool.AllTx) > 0 {
			BlockchainInstance.PackageTx(&newBlock)
		}else {
			newBlock.Accounts = BlockchainInstance.LastBlock().Accounts
			newBlock.Transactions = make([]Transaction, 0)
		}

		if IsBlockValid(newBlock, BlockchainInstance.Blocks[len(BlockchainInstance.Blocks)-1]) {
			//pow 内容
			mutex.Lock()
			BlockchainInstance.Blocks = append(BlockchainInstance.Blocks, newBlock)
			mutex.Unlock()
			//pos 内容
			//candidateBlocks <- newBlock
		}
		//pos 内容
		//<-hasbeenValid

		bytes, err := json.Marshal(BlockchainInstance.Blocks)
		if err != nil {
			log.Println(err)
		}

		BlockchainInstance.WriteDate2File()

		b, err := json.MarshalIndent(BlockchainInstance.Blocks, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		// Green console color: 	\x1b[32m
		// Reset console color: 	\x1b[0m
		fmt.Printf("\x1b[32m%s\x1b[0m ", string(b))

		mutex.Lock()
		rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
		rw.Flush()
		mutex.Unlock()
	}

}

func GenPosAddress() string {
	var address string
	t := time.Now()
	address = SHA256Hasing(t.String())
	//@todo
	validators[address] = 500
	fmt.Println(validators)
	return address
}



// make sure block is valid by checking index, and comparing the hash of the previous block
func IsBlockValid(newBlock, oldBlock Block) bool {
	return true
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
func GenerateBlock(oldBlock Block, Result int,address string) Block {
	var newBlock Block

	t := time.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.Result = Result
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Difficulty = difficulty
	newBlock.Validator = address
	//pow 内容
	for i := 0; ; i++ {
		hex := fmt.Sprintf("%x", i)
		newBlock.Nonce = hex
		if !isHashValid(calculateHash(newBlock), newBlock.Difficulty) {
			fmt.Println(calculateHash(newBlock), " do more work!")
			//time.Sleep(time.Second)
			continue
		} else {
			fmt.Println(calculateHash(newBlock), " work done!")
			newBlock.Hash = calculateHash(newBlock)
			break
		}

	}
	return newBlock
}

func isHashValid(hash string, difficulty int) bool {
	prefix := strings.Repeat("0", difficulty)
	return strings.HasPrefix(hash, prefix)
}

// SHA256 hasing
func calculateHash(block Block) string {
	record := strconv.Itoa(block.Index) + block.Timestamp + strconv.Itoa(block.Result) + block.PrevHash + block.Nonce
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

func IsExist(file string) bool {
	fi, e := os.Stat(file)
	if e != nil {
		return os.IsExist(e)
	}
	return !fi.IsDir()
}