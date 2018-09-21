package rpc

import (
	"io"
	"time"
	"strconv"
	"net/http"
	"encoding/json"

	"Course/blockchain"


	"github.com/gorilla/mux"
	"github.com/davecgh/go-spew/spew"

	"Course/wallet"
	"fmt"
)

func makeMuxRouter() http.Handler {
	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/", handleGetBlockchain).Methods("GET")
	muxRouter.HandleFunc("/block", handleWriteBlock).Methods("POST")   // post请求： 本地产生一个块（若交易池中有交易，则打包进块），并将块信息广播到对端节点 e.g {"Msg": 123}
	muxRouter.HandleFunc("/txpool", handleWriteTx).Methods("POST")     // post请求： 向本地交易池中放入新的交易  e.g {"From":"0x1","To":"0x2","Value":10000,"Data":"message"}
	muxRouter.HandleFunc("/getbalance", handleBalance).Methods("POST") // 查询账户地址 { "Address":"0x12312132"}
	return muxRouter
}

func handleGetBlockchain(w http.ResponseWriter, r *http.Request) {
	bytes, err := json.MarshalIndent(blockchain.BlockchainInstance, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	io.WriteString(w, string(bytes))
}

type SendTxArgs struct {
	From     string
	To       string
	Value    uint64
	Data  	 string
}

func handleWriteTx(w http.ResponseWriter, r *http.Request) {
	var m SendTxArgs
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		respondWithJSON(w, r, http.StatusBadRequest, r.Body)
		return
	}
	defer r.Body.Close()

	validate := wallet.Validate_Address(m.From, m.To, m.Value,blockchain.WalletSuffix)
	if false == validate {
		fmt.Println("Parameter verification failed")
		respondWithJSON(w, r, http.StatusCreated, "Parameter verification failed")
		return
	}
	balance := blockchain.BlockchainInstance.GetBalance(m.From)
	if balance < m.Value {
		respondWithJSON(w, r, http.StatusCreated, "Insufficient balance")
		fmt.Println("Insufficient balance")
		return
	}
	tx := blockchain.BlockchainInstance.NewTransaction(m.From, m.To, m.Value, []byte(m.Data))
	blockchain.BlockchainInstance.AddTxPool(tx)

	respondWithJSON(w, r, http.StatusCreated, tx)

}

type Message struct {
	Msg int
}

func handleWriteBlock(w http.ResponseWriter, r *http.Request) {
	var m Message

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		respondWithJSON(w, r, http.StatusBadRequest, r.Body)
		return
	}
	defer r.Body.Close()

	newBlock := blockchain.GenerateBlock(blockchain.BlockchainInstance.Blocks[len(blockchain.BlockchainInstance.Blocks)-1], m.Msg)


	if len(blockchain.BlockchainInstance.TxPool.AllTx) > 0 {
		blockchain.BlockchainInstance.PackageTx(&newBlock)
	}else {
		newBlock.Accounts = blockchain.BlockchainInstance.LastBlock().Accounts
	}

	if blockchain.IsBlockValid(newBlock, blockchain.BlockchainInstance.Blocks[len(blockchain.BlockchainInstance.Blocks)-1]) {
		blockchain.Lock()
		blockchain.BlockchainInstance.Blocks = append(blockchain.BlockchainInstance.Blocks, newBlock)
		blockchain.UnLock()
		spew.Dump(blockchain.BlockchainInstance.Blocks)
	}

	respondWithJSON(w, r, http.StatusCreated, newBlock)

}

type RequestBalance struct {
	Address string
}

func handleBalance(w http.ResponseWriter, r *http.Request) {
	var rb RequestBalance
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&rb); err != nil {
		respondWithJSON(w, r, http.StatusBadRequest, r.Body)
		return
	}
	defer r.Body.Close()

	balance := blockchain.BlockchainInstance.GetBalance(rb.Address)
	respondWithJSON(w, r, http.StatusCreated, balance)
}

func respondWithJSON(w http.ResponseWriter, r *http.Request, code int, payload interface{}) {
	response, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("HTTP 500: Internal Server Error"))
		return
	}
	w.WriteHeader(code)
	w.Write(response)
}


func RunHttpServer(port int) error {
	mux := makeMuxRouter()
	listentPort := strconv.Itoa(port)
	fmt.Println("local http server listening on 127.0.0.1:"+listentPort)
	s := &http.Server{
		Addr:           ":" + listentPort,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := s.ListenAndServe(); err != nil {
		return err
	}

	return nil
}
