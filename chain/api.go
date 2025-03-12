package chain

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/glossd/fetch"
)

func RunAPI(port int) {
	sm := &http.ServeMux{}

	sm.HandleFunc("GET /api/blocks", fetch.ToHandlerFunc(func(in fetch.Request[fetch.Empty]) ([]Block, error)  {
		sort := in.Parameters["sort"]
		limit, err := strconv.Atoi(in.Parameters["limit"])
		if err != nil {
			limit = 20
		}
		skip, _ := strconv.Atoi(in.Parameters["skip"])
		
		// order descending
		j := blockchain.Len() - skip
		i := j - limit
		if sort == "asc" {
			j = skip
			i = j + limit
		}
		return blockchain.LoadRangeSafe(i, j), nil
	}))

	sm.HandleFunc("GET /api/blocks/search", fetch.ToHandlerFunc(func(in fetch.Request[fetch.Empty]) (Block, error)  {
		b, ok := blockchain.Load(in.Parameters["hash"])
		if ok {
			return b, nil
		}

		i, err := strconv.Atoi(in.Parameters["index"])
		if err == nil {
			if i < 0 || i >= blockchain.Len() {
				return Block{}, &fetch.Error{Status: 400, Msg: "index out of bound"}
			}
			return blockchain.LoadIndex(i), nil
		}
		
		return Block{}, &fetch.Error{Status: 404, Msg: "block not found"}
	}))

	sm.HandleFunc("GET /api/blocks/last", fetch.ToHandlerFuncEmptyIn(func() (Block, error)  {
		return blockchain.Last(), nil
	}))

	sm.HandleFunc("POST /api/blocks", fetch.ToHandlerFuncEmptyOut(func(in Block) error {
		return Broadcast(in)
	}))


	
	sm.HandleFunc("GET /api/mempool", fetch.ToHandlerFunc(func(in fetch.Request[fetch.Empty]) ([]Transaction, error)  {
		limit, err := strconv.Atoi(in.Parameters["limit"])
		if err != nil {
			limit = 20
		}

		return Top(limit), nil
	}))

	sm.HandleFunc("POST /api/mempool", fetch.ToHandlerFuncEmptyOut(func(in Transaction) error {
		return Push(in)
	}))



	sm.HandleFunc("/api/difficulty/target/bits", fetch.ToHandlerFuncEmptyIn(func() (uint32, error)  {
		return GetDiffuctlyTargetBits(), nil
	}))

	sm.HandleFunc("/api/reward", fetch.ToHandlerFuncEmptyIn(func() (uint64, error)  {
		return uint64(GetMinerReward()), nil
	}))

	http.ListenAndServe(fmt.Sprintf(":%d", port), sm)
}
