/*
Package messenger sends messages to Discord channel
*/
package messenger

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"encoding/json"

	"github.com/golang/glog"
	"github.com/livepeer/stream-tester/model"
	"github.com/patrickmn/go-cache"
)

var (
	webhookURL    string
	userName      string
	usersToNotify string
	debounceCache *cache.Cache = cache.New(5*time.Minute, 30*time.Minute)
	msgCh         chan string
	msgQueue      []string
)

type discordMessage struct {
	Content  string `json:"content,omitempty"`
	UserName string `json:"username,omitempty"`
}

// Init ...
func Init(WebhookURL, UserName, UsersToNotify string) {
	webhookURL = WebhookURL
	userName = UserName
	usersToNotify = UsersToNotify
	if WebhookURL != "" {
		msgCh = make(chan string, 64)
		go sendLoop()
	}
}

func sendLoop() {
	for {
		msg := <-msgCh
		postMessage(msg)
	}
}

// SendFatalMessage send message to Discord channel
// and automatically mentiones UsersToNotify in the message
func SendFatalMessage(msg string) {
	glog.Error(msg)
	if usersToNotify != "" {
		msg = usersToNotify + ": " + msg
	}
	sendMessage(msg)
}

// SendMessage send message to Discord channel
func SendMessage(msg string) {
	if msg == "" {
		return
	}
	glog.Info(msg)
	sendMessage(msg)
}

// SendCodeMessage send message to Discord channel, wrapping it as three ticks
func SendCodeMessage(msg string) {
	if msg == "" {
		return
	}
	glog.Info(msg)
	sendMessage("```\n" + msg + "```")
}

// SendMessageDebounced send message to Discord channel
func SendMessageDebounced(msg string) {
	glog.Info(msg)
	if _, has := debounceCache.Get(msg); !has {
		sendMessage(msg)
		debounceCache.SetDefault(msg, true)
	}
}

func sendMessage(msg string) {
	if webhookURL == "" || msgCh == nil {
		return
	}
	msgCh <- msg
}

func postMessage(msg string) http.Header {
	if webhookURL == "" {
		return nil
	}
	if len(msg) > 2000 {
		for {
			l := 1980
			if l > len(msg) {
				l = len(msg)
			}
			msg1 := msg[:l]
			msg = msg[l:]
			SendMessage(msg1)
			if len(msg) == 0 {
				break
			}
			time.Sleep(time.Second)
		}
		return nil
	}
	dm := &discordMessage{
		Content:  msg,
		UserName: userName,
	}
	data, _ := json.Marshal(dm)
	var body io.Reader
	body = bytes.NewReader(data)
	// resp, err := http.Post(webhookURL, "application/json", body)
	req, _ := http.NewRequest("POST", webhookURL, body)
	req.Header.Add("User-Agent", "stream-tester/"+model.Version)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-RateLimit-Precision", "millisecond")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		glog.Errorf("error posting to Discord err=%v", err)
		return nil
	}
	b, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	glog.Infof("Discord response headers")
	for k, v := range resp.Header {
		glog.Infof("%s: %+v", k, v)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		glog.Errorf("status error posting to Discord status=%s body: %s", resp.Status, string(b))
	}
	return resp.Header
}
