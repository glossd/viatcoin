package chain

import (
	"encoding/json"
	"fmt"
	"github.com/glossd/memmq"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{} // use default options

func serverStream(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	go func() {
		for {
			_, _, err := c.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					fmt.Printf("Error WS ReadMessage: %v\n", err)
				}
				break
			}
		}
	}()
	go func() {
		memmq.Subscribe("blocks", func(msg interface{}) bool {
			b, ok := msg.(Block)
			if !ok {
				return true
			}
			err := c.WriteJSON(b)
			if err != nil {
				fmt.Printf("Error WS Write: %s\n", err)
			}
			return true
		})
		// todo mempool updates
	}()
}

func Stream(apiURL string, process func(Block)) error {
	c, _, err := websocket.DefaultDialer.Dial(apiURL+"/api/stream", nil)
	if err != nil {
		return fmt.Errorf("dial: %s", err)
	}
	defer c.Close()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			return fmt.Errorf("read: %s", err)
		}
		var b Block
		err = json.Unmarshal(message, &b)
		if err != nil {
			return fmt.Errorf("unmarshal: %s", err)
		}
		process(b)
	}
}
