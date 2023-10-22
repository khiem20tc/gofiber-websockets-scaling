package websocket

import (
	"gofiber-ws/redis"
	"log"
	"sync"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

var UpgradeWebsocket = func(c *fiber.Ctx) error {
	// IsWebSocketUpgrade returns true if the client
	// requested upgrade to the WebSocket protocol.
	if websocket.IsWebSocketUpgrade(c) {
		c.Locals("allowed", true)
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

var HandleWebSocket = websocket.New(func(c *websocket.Conn) {

	// c.Locals is added to the *websocket.Conn
	log.Println(c.Locals("allowed"))  // true
	log.Println(c.Params("id"))       // 123
	log.Println(c.Query("v"))         // 1.0
	log.Println(c.Cookies("session")) // ""

	// DIRECT SOCKET
	socket_id := c.Params("session")
	pubsub := redis.Client.Subscribe(redis.Context, socket_id)
	defer pubsub.Close()

	// BROADCAST SOCKET
	pubsubBroadcast := redis.Client.Subscribe(redis.Context, "broadcast")
	defer pubsubBroadcast.Close()

	// websocket.Conn bindings https://pkg.go.dev/github.com/fasthttp/websocket?tab=doc#pkg-index

	const msgType = websocket.TextMessage //TextMessage type

	wg := sync.WaitGroup{}
	wg.Add(3)

	server2clientDirect := func() {
		defer wg.Done()

		for {
			// DIRECT SERVER TO CLIENT
			pubsubmsg, err := pubsub.ReceiveMessage(redis.Context)
			if err != nil {
				panic(err)
			}

			msg := []byte(pubsubmsg.Payload)

			err = c.WriteMessage(msgType, msg)
			if err != nil {
				log.Println("write:", err)
				break
			}

		}
	}

	server2clientBroadcast := func() {
		defer wg.Done()

		for {
			// BROADCAST SERVER TO CLIENT
			pubsubmsgBroadcast, err := pubsubBroadcast.ReceiveMessage(redis.Context)
			if err != nil {
				panic(err)
			}

			msgBroadcast := []byte(pubsubmsgBroadcast.Payload)

			err = c.WriteMessage(msgType, msgBroadcast)
			if err != nil {
				log.Println("write:", err)
				break
			}
		}
	}

	client2server := func() {
		defer wg.Done()

		for {
			// ignore msgType
			_, msg, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				break
			}

			// Direct client to server
			if err := redis.Client.Publish(redis.Context, socket_id, msg).Err(); err != nil {
				panic(err)
			}

			// Broadcast client to server
			// if err := client.Publish("broadcast", msg).Err(); err != nil {
			// 	panic(err)
			// }

		}
	}

	go client2server()
	go server2clientDirect()
	go server2clientBroadcast()

	wg.Wait()

})
