package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
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

type Message struct{
	ID string
	Chat string
	Time string
	Name string
	Text string
}

var clients = make(map[string]map[*websocket.Conn]bool)

var Msg = make(chan Message)

var MessageArchieve = (map[string][]Message{})

var GeneralPass string

//хендлер який використовується для введення коду перед видаленням всіх повідомлень
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

//хендлер для очищення чату який видаляє всі повідомлення чату з архіву повідомлень, а потім відправляє клієнту і вигляді html пусте вікно для повідолмень
func clear(c echo.Context) error {
	mu := sync.Mutex{}
	pass := c.FormValue("pass")
	if pass == GeneralPass {
		mu.Lock()
		//deleting MessageArchieve (all chats)
		for chat, _ := range MessageArchieve{
			delete(MessageArchieve, chat)
		}
		mu.Unlock()

		mu.Lock()

		//sends to all clients empty chat field
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

func deleteMsg(c echo.Context) error {
	chat := c.Param("chat")
	msg := c.Param("msg")
	mu := sync.Mutex{}
	pass := c.FormValue("pass")
	//TODO: implement pass
	if pass == "" {
		mu.Lock()
		//deleting MessageArchieve (all chats)
		for msgIdx, message := range MessageArchieve[chat]{
			if message.ID == msg {
				MessageArchieve[chat] = append(MessageArchieve[chat][:msgIdx], MessageArchieve[chat][msgIdx+1:]...)
			}
		}
		mu.Unlock()

		mu.Lock()

		//sends to all clients empty chat field
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

		mu.Lock()
		for client, admin := range clients[chat] {
			for _, msg := range MessageArchieve[chat] {
				var err error
				if admin {
					err = client.WriteMessage(websocket.TextMessage, []byte(addHTMLwithDeleteButton(msg)))
				} else {
					err = client.WriteMessage(websocket.TextMessage, []byte(addHTML(msg)))
				}
				if err != nil {
					log.Printf("error: %v", err)
					client.Close()
				}
			}
		}
		mu.Unlock()

		return c.HTML(200, "Успішно!")
	} else {
		return c.HTML(200, "Не успішно!")
	}
}

//хендлер для створення WS-зв'язку та відправлення клієнту всіх попередніх звернень
func handleConnections(c echo.Context) error {
	id := c.Param("id")
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
		clients[id][ws] = false
		mu.Unlock()
	}
	return nil
}

func handleConnectionsAdmin(c echo.Context) error {
	id := c.Param("id")
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
			err := ws.WriteMessage(websocket.TextMessage, []byte(addHTMLwithDeleteButton(msg)))
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

//Функція яка запускається як горутина, отримує всі повідомлення, складає їх в архів повідомлень та надсилає всім учасникам відповідного чату
func handleMessages() {
	var messageNumber int
	mu := sync.Mutex{}
	for msg := range Msg {
		mu.Lock()
		msg.Text = wrapURLsWithAnchorTags(msg.Text)
		msg.ID = strconv.Itoa(messageNumber)
		messageNumber++
		MessageArchieve[msg.Chat] = append(MessageArchieve[msg.Chat], msg)
		mu.Unlock()
		// Send it out to every client that is currently connected
		mu.Lock()
		for client, admin := range clients[msg.Chat] {
			var err error
			if admin {
				err = client.WriteMessage(websocket.TextMessage, []byte(addHTMLwithDeleteButton(msg)))
			} else {
				err = client.WriteMessage(websocket.TextMessage, []byte(addHTML(msg)))
			}
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients[msg.Chat], client)
			}
		}
		mu.Unlock()
	}
}

//хендлер для отримання нових повідомлень
func recieve(c echo.Context) error {
	id := c.Param("id")
	name := c.FormValue("name")
	text := c.FormValue("text")

	currentTime := time.Now().Format("15:04")

	Msg <- Message{Chat: id, Time: currentTime, Name: name, Text: text}

	//html який повертається і замінює форму відправки повідомлення
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

//Функція, яка з типу повідомлення створює html-розмітку, яка містить текст повідомлення, автора та час його створення
func addHTML(message Message) string {
	return fmt.Sprintf("<div id=\"notifications\" hx-swap-oob=\"afterbegin\" style=\"word-wrap: break-word;\"><p style=\"word-wrap: break-word;\"><small>%s</small><br><b>%s</b>: %s</p></div>", message.Time, message.Name, message.Text)
}

func addHTMLwithDeleteButton(message Message) string {
	return fmt.Sprintf("<div id=\"notifications\" hx-swap-oob=\"afterbegin\" style=\"word-wrap: break-word;\"><p style=\"word-wrap: break-word;\"><small>%s <a style=\"color:red;\" hx-post=\"https://livechatextension.pp.ua/delete/%s/%s\">del</a></small><br><b>%s</b>: %s </p></div>", message.Time, message.Chat, message.ID, message.Name, message.Text)
}

//Функція для трансформації тексту у клікабельні посилання шляхом додавання html-тегу <a>
func wrapURLsWithAnchorTags(text string) string {
    // Регулярний вираз для виявлення URL
    urlPattern := `https?://[^\s]+`
    re := regexp.MustCompile(urlPattern)

    // Заміна URL на HTML-тег <a>
    result := re.ReplaceAllStringFunc(text, func(url string) string {
        return fmt.Sprintf("<a href=\"%s\" target=\"blank\">%s</a>", url, url)
    })

    return result
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
	e.POST("/delete/:chat/:msg", deleteMsg)
	e.GET("/ws/:id", handleConnections)
	e.GET("/wsadmin/:id", handleConnectionsAdmin)
	e.POST("/send/:id", recieve)
	e.POST("/clear", clear)
	go handleMessages()
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", port)))
}
