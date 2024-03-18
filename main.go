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

type Message struct{
	Chat string
	Time string
	Name string
	Text string
}

var clients = make(map[string]map[*websocket.Conn]bool)

var Msg = make(chan Message)

var MessageArchieve = (map[string][]Message{})

var GeneralPass string

func home(c echo.Context) error {
	return c.HTML(200, `
	<form method="POST" action="/clear">
	<div class="container">
	<div class="col">
	<input type="text" name="pass" id="pass" class="form-control" placeholder="Password" required><br>
	</div>
	<button type="submit" class="btn btn-primary btn-user btn-block">
	Очистити чат
	</button>
	</div>
	</form>`)
}

func clear(c echo.Context) error {
	mu := sync.Mutex{}
	pass := c.FormValue("pass")
	if pass == GeneralPass {
		mu.Lock()
		for chat, _ := range MessageArchieve{
			delete(MessageArchieve, chat)
		}
		mu.Unlock()
		fmt.Println("Chat cleared")

		mu.Lock()
		for chat, chatClients := range clients {
			for client := range chatClients {
				err := client.WriteMessage(websocket.TextMessage, []byte(`<div id="chat_room" hx-swap-oob="morphdom"><div class="col fixed-height-block p-3" id="notifications"></div></div>`))
				if err != nil {
					log.Printf("error: %v", err)
					client.Close()
					delete(clients[chat], client)
				}
			}
		}
		mu.Unlock()

		return c.HTML(200, "Успішно!")
	} else {
		return c.HTML(200, "Не успішно!")
	}
}

func handleConnections(c echo.Context) error {
	id := c.Param("id")
	fmt.Println("::::", id)
	mu := sync.Mutex{}
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	chatClients, ok := clients[id]

	if !ok {
		clients[id] = make(map[*websocket.Conn]bool)
	}

	_, ok = chatClients[ws]

	if !ok {
		mu.Lock()
		for _, msg := range MessageArchieve[id] {
			err := ws.WriteMessage(websocket.TextMessage, []byte(addHTML(msg)))
			if err != nil {
				log.Printf("error: %v", err)
				ws.Close()
			}
		}
		clients[id][ws] = true
		mu.Unlock()
	}
	return nil
}

func handleMessages() {
	mu := sync.Mutex{}
	for msg := range Msg {
		mu.Lock()
		MessageArchieve[msg.Chat] = append(MessageArchieve[msg.Chat], msg)
		mu.Unlock()
		// Send it out to every client that is currently connected
		mu.Lock()
		for client := range clients[msg.Chat] {
			err := client.WriteMessage(websocket.TextMessage, []byte(addHTML(msg)))
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients[msg.Chat], client)
			}
		}
		mu.Unlock()
	}
}

func recieve(c echo.Context) error {
	id := c.Param("id")
	name := c.FormValue("name")
	text := c.FormValue("text")

	currentTime := time.Now().Format("15:04:05")

	Msg <- Message{Chat: id, Time: currentTime, Name: name, Text: text}

	return c.HTML(200, fmt.Sprintf(`
	<div class="input-group input-group-sm border-right-0">
	<input type="text" name="name" id="name" class="form-control col-1 border-right-0" placeholder="Ваше ім'я" value="%s" hidden required>
</div>
<div class="input-group input-group-sm border-right-0">
	<textarea type="text" name="text" id="text" class="form-control border-right-0" placeholder="напишіть щось..." required></textarea><br>
</div>
<div class="d-grid gap-2 d-md-flex justify-content-md-end p-1">
	<button type="submit" class="btn btn-sm border-right-0 btn-primary">
		<span class="icon text-white">
			<i class="fa fa-paper-plane"> Відправити</i>
		</span>
	</button>
</div>
	`, name))
}

func addHTML(message Message) string {
	return fmt.Sprintf("<div id=\"notifications\" hx-swap-oob=\"afterbegin\"><p><small>%s</small><br><b>%s</b>: %s</p></div>", message.Time, message.Name, message.Text)
}

func main() {
	GeneralPass = os.Getenv("PASS")
	port := os.Getenv("PORT")
	fmt.Printf("Port: %s. Pass: %s\n", port, GeneralPass)
	e := echo.New()
	e.Use(middleware.CORS())
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Static("/", "./static")
	e.GET("/", home)
	e.GET("/ws/:id", handleConnections)
	e.POST("/send/:id", recieve)
	e.POST("/clear", clear)
	go handleMessages()
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", port)))
}
