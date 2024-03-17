package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

type ChatMessage struct {
	ChatText string `json:"chat_message"`
}

var clients = make(map[*websocket.Conn]bool)

var Msg = make(chan string)

var MessageArchieve []string

var GeneralPass string

func home(c echo.Context) error {
	return c.HTML(200, `
	<form method="POST" action="/clear">
	<div class="container">
	<div class="col">
	<input type="text" name="pass" id="pass" class="form-control" placeholder="Password" required><br>
	</div>
	<button type="submit" class="btn btn-primary btn-user btn-block">
	>>
	</button>
	</div>
	</form>`)
}

func clear(c echo.Context) error {
	mu := sync.Mutex{}
	pass := c.FormValue("pass")
	if pass == GeneralPass {
		mu.Lock()
		MessageArchieve = []string{}
		mu.Unlock()
		fmt.Println("Chat cleared")
		return c.HTML(200, "Успішно!")
	} else {
		return c.HTML(200, "Не успішно!")
	}
}

func handleConnections(c echo.Context) error {
	mu := sync.Mutex{}
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	defer ws.Close()

	mu.Lock()
	for _, msg := range MessageArchieve {
		err := ws.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				log.Printf("error: %v", err)
				ws.Close()
			}
	}
	mu.Unlock()

	mu.Lock()
	clients[ws] = true
	mu.Unlock()

	for {
		time.Sleep(30 *time.Second)
		mu.Lock()
		available, ok := clients[ws]
		mu.Unlock()
		if !available || !ok {
			return nil
		}
	}
}

func handleMessages() {
	mu := sync.Mutex{}
	for msg := range Msg{
		mu.Lock()
		MessageArchieve = append(MessageArchieve, msg)
		mu.Unlock()
		// Send it out to every client that is currently connected
		mu.Lock()
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
		mu.Unlock()
	}
}

func recieve(c echo.Context) error {
	name := c.FormValue("name")
	text := c.FormValue("text")
	
	Msg <- fmt.Sprintf("<div id=\"notifications\" hx-swap-oob=\"afterbegin\"><p><b>%s</b>: %s</p></div>", name, text)

	return c.HTML(200, fmt.Sprintf(`
        <div class="input-group mb-3">
            <input type="text" name="name" id="name" class="form-control col-3" placeholder="Name" value="%s" required><br>
            <input type="text" name="text" id="text" class="form-control" placeholder="Text" required><br>
            <button type="submit" class="btn btn-outline-secondary">
                >>
            </button>
        </div>
	`, name))
}

func main() {
	GeneralPass = os.Getenv("PASS")
	port := os.Getenv("PORT")
	fmt.Printf("Port: %s. Pass: %s\n", port, GeneralPass)
	e := echo.New()
	e.Use(middleware.CORS())
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Static("/", "../public")
	e.GET("/", home)
	e.GET("/ws", handleConnections)
	e.POST("/send", recieve)
	e.POST("/clear", clear)
	go handleMessages()
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", port)))
}