package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/gofiber/websocket/v2"
)

var (
	adminMsgChan = make(chan string)
	clientConns  = make(map[*websocket.Conn]bool)
)

func Routes() error {
	eng := html.New("./static", ".html")

	app := fiber.New(fiber.Config{
		Views: eng,
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("client", fiber.Map{
			"message": "This is a message",
		})
	})

	app.Use("/ws", websocket.New(func(c *websocket.Conn) {
		// Add the client connection to the map
		clientConns[c] = true

		for {
			ty, msg, err := c.ReadMessage()
			if err != nil {
				delete(clientConns, c)
				break
			}
			// Broadcast the message to all connected clients
			for client := range clientConns {
				client.WriteMessage(ty, msg)
			}
		}
	}))

	// admin route
	app.Get("/admin", func(c *fiber.Ctx) error {
		return c.Render("admin", fiber.Map{})
	})

	app.Post("/admin/post", func(c *fiber.Ctx) error {
		msg := c.FormValue("message")

		// for client := range clientConns {
		// 	if err := client.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
		// 		return err
		// 	}
		// }

		adminMsgChan <- msg

		return c.JSON(fiber.Map{"message": "Message sent successfully"})
	})

	go func() {
		for {
			// Continuously listen for admin messages
			msg := <-adminMsgChan
			// Broadcast the admin message to all connected clients
			for client := range clientConns {
				if err := client.WriteMessage(websocket.TextMessage, []byte("[ADMIN] "+msg)); err != nil {
					continue
				}
			}
		}
	}()

	return app.Listen(":8080")
}

func main() {
	Routes()
}
