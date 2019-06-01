package bot

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"strings"
	"sync"
	"time"
)

const ApiUrl = "https://api.telegram.org/bot%s/%s"

var httpClient = http.Client{}

type TgConnection struct {
	token string
}

func (tgConn *TgConnection) sendQuery(command string, queryMethod string, contentType string, body *bytes.Buffer) string {

	switch queryMethod {
	case "GET":
		cmd := fmt.Sprintf(ApiUrl, tgConn.token, command)
		resp, err := httpClient.Get(cmd)
		if err != nil {
			fmt.Println(err)
		}
		return tgConn.processResp(resp)
	case "POST":
		cmd := fmt.Sprintf(ApiUrl, tgConn.token, command)
		req, _ := http.NewRequest("POST", cmd, body)
		req.Header.Set("Content-Type", contentType)
		resp, err := httpClient.Do(req)
		if err != nil {
			fmt.Println(err)
		}
		return tgConn.processResp(resp)
	}
	return "1"
}

func (tgConn *TgConnection) processResp(resp *http.Response) string {
	body, _ := ioutil.ReadAll(resp.Body)
	strBody := string(body)
	fmt.Println(strBody)
	return strBody
}

type Bot struct {
	connection TgConnection
	commands   map[string]Command
	offset     int
	limit      int
	mx         sync.Mutex
}

func NewBot(token string) *Bot {
	conn := TgConnection{token: token}
	bot := &Bot{connection: conn,
		commands: make(map[string]Command),
		offset:   0,
		limit:    10,
	}
	go bot.Connect()
	return bot
}

func (bot *Bot) Connect() {
	for {
		resp := bot.getUpdates()
		bot.receiveMsg(resp)
		d, _ := time.ParseDuration("1s")
		time.Sleep(d)
	}
}

func (bot *Bot) SendMessage(chatId int, msg string) {
	msg = strings.Replace(msg, "\n", "%0D%0A", -1)
	cmd := fmt.Sprintf("sendMessage?chat_id=%d&text=%s", chatId, msg)
	bot.connection.sendQuery(cmd, "GET", "", nil)
}

func (bot *Bot) SendPhoto(chatId int, img image.Image) {

	var buf = new(bytes.Buffer)
	w := bufio.NewWriter(buf)
	err := jpeg.Encode(w, img, nil)
	if err != nil {
		fmt.Println(err)
	}

	formBuf := &bytes.Buffer{}
	m_w := multipart.NewWriter(formBuf)
	fmt.Println(m_w.FormDataContentType())
	file, err := m_w.CreateFormFile("photo", "")
	if err != nil {
		fmt.Println(err)
	}
	_, _ = file.Write(buf.Bytes())

	err = m_w.Close()
	if err != nil {
		fmt.Println(err)
	}
	cmd := fmt.Sprintf("sendPhoto?chat_id=%d", chatId)
	bot.connection.sendQuery(cmd, "POST", m_w.FormDataContentType(), formBuf)
}

func (bot *Bot) GetMe(c chan string) {
	resp := bot.connection.sendQuery("getMe", "GET", "", nil)
	c <- resp
}

func (bot *Bot) getUpdates() *Updates {
	cmd := fmt.Sprintf("getUpdates?limit=%d&offset=%d", bot.limit, bot.offset)
	resp := bot.connection.sendQuery(cmd, "GET", "", nil)
	data := &Updates{}
	_ = json.Unmarshal([]byte(resp), data)
	//c <- data
	return data
}

func (bot *Bot) receiveMsg(updates *Updates) {
	for _, elem := range updates.Result {
		cmd := elem.Message.Text
		log.Println(elem.UpdateId)
		bot.offset = elem.UpdateId + 1

		cb, ok := bot.commands[cmd]
		if ok {
			cb.Callback(&elem)
		} else {
			bot.DefaultCallback(&elem)
		}
	}
}

func (bot *Bot) AddCommand(command string, cb func(result *Result)) {
	cmd := Command{Command: command, Callback: cb}
	bot.commands[command] = cmd
}

func (bot *Bot) DefaultCallback(result *Result) {
	fmt.Println(result.Message.From.Id)
	bot.SendMessage(result.Message.From.Id, result.Message.Text)
}

func (bot *Bot) Start() {
	log.Println(fmt.Sprintf("TG BOT STARTED WITH token = %s", bot.connection.token))
	select {}
}
