package chain

type SyncData struct {
	Blocks  []Block
	Wallets map[string][]int64
	MemPool []Transaction
}

func SyncWith(apiUrl string) {
	// todo sync this node with a new one
}
