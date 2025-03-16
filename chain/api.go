package chain

import (
	"fmt"
	"math"
	"net/http"
	"slices"
	"strconv"

	"github.com/glossd/fetch"
)

func RunAPI(port int) {
	sm := &http.ServeMux{}

	sm.HandleFunc("GET /api/blocks", fetch.ToHandlerFunc(func(in fetch.RequestEmpty) ([]Block, error) {
		sort := in.Parameters["sort"]
		limit, err := strconv.Atoi(in.Parameters["limit"])
		if err != nil {
			limit = 20
		}
		if limit == -1 {
			limit = math.MaxInt
		}
		skip, _ := strconv.Atoi(in.Parameters["skip"])

		if sort == "asc" {
			j := skip
			i := j + limit
			return blockchain.LoadRangeSafe(i, j), nil
		} else {
			// descending
			j := blockchain.Len() - skip
			i := j - limit
			res := blockchain.LoadRangeSafe(i, j)
			slices.Reverse(res)
			return res, nil
		}
	}))

	sm.HandleFunc("GET /api/blocks/search", fetch.ToHandlerFunc(func(in fetch.RequestEmpty) (Block, error) {
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

	sm.HandleFunc("GET /api/blocks/last", fetch.ToHandlerFuncEmptyIn(func() (Block, error) {
		return blockchain.Last(), nil
	}))

	sm.HandleFunc("POST /api/blocks", fetch.ToHandlerFuncEmptyOut(func(in Block) error {
		return Broadcast(in)
	}))

	sm.HandleFunc("GET /api/mempool", fetch.ToHandlerFunc(func(in fetch.Request[fetch.Empty]) ([]Transaction, error) {
		limit, err := strconv.Atoi(in.Parameters["limit"])
		if err != nil {
			limit = 20
		}
		if limit == -1 {
			return Top(math.MaxInt), nil
		}

		return Top(limit), nil
	}))

	sm.HandleFunc("POST /api/mempool", fetch.ToHandlerFuncEmptyOut(func(in Transaction) error {
		return Push(in)
	}))

	sm.HandleFunc("/api/difficulty/target/bits", fetch.ToHandlerFuncEmptyIn(func() (uint32, error) {
		return GetDiffuctlyTargetBits(), nil
	}))

	sm.HandleFunc("/api/reward", fetch.ToHandlerFuncEmptyIn(func() (uint64, error) {
		return uint64(GetMinerReward()), nil
	}))

	http.ListenAndServe(fmt.Sprintf(":%d", port), sm)
}
