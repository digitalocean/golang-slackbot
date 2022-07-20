package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

// Request takes in the StatusCode and Content from other functions to display to the user's slack.
type Request struct {
	// StatusCode is the http code that will be returned back to the user.
	StatusCode int `json:"statusCode"`
	// Content will contain the presigned url, error messages, or success messages.
	Content string `json:"body"`
}

// Response returns back the http code, type of data, and the presigned url to the user.
type Response struct {
	// StatusCode is the http code that will be returned back to the user.
	StatusCode int `json:"statusCode,omitempty"`
	// Headers is the information about the type of data being returned back.
	Headers map[string]string `json:"headers,omitempty"`
	// Body will contain the success or failed messages.
	Body string `json:"body,omitempty"`
}

// Email is the json struct of the email message that contains the email addresses, subject, and message.
type Email struct {
	// From is the email address of the user.
	From string
	// To is the email address that the user wants to send to.
	To string
	// Subject is the subject of the email that the user wants to send.
	Subject string
	// Message is the body of the email that the user wants to send.
	Message string
}

// Sms is the json struct of the sms message that contains the phone numbers and message.
type Sms struct {
	// Number is the user's phone number that they are sending from.
	Number string
	// To is the phone number they are sending a message to.
	To string
	// Message is the body of the message that the user wants to send.
	Message string
}

// Url is the json struct of the presigned url function that contains the filename, type, and duration.
type Url struct {
	// Filename is the name of the file that will be uploaded or downloaded.
	Filename string
	// Type is a presigned request type to "GET" or "PUT" an object.
	Type string
	// Duration is the duration in which the presigned url will last.
	Duration string
}

var (
	auth_tok, app_tok, channelid, url string
	// ErrRequest will return an error if the request was not successful.
	ErrRequest = errors.New("request was not successful")
	// ErrClosed will return an error if the channel provided is closed.
	ErrClosed = errors.New("channel closed")
)

func init() {
	auth_tok = os.Getenv("AUTH_TOKEN")
	if auth_tok == "" {
		panic("no auth_tok provided")
	}
	app_tok = os.Getenv("APP_TOKEN")
	if app_tok == "" {
		panic("no app_tok provided")
	}
	channelid = os.Getenv("CHANNEL_ID")
	if channelid == "" {
		panic("no channelid provided")
	}
	url = os.Getenv("URL")
	if url == "" {
		panic("no url provided")
	}
}

// Main configures a client with socketmode and slacks API using the token, app token, and channelid that the slack bot will be in,
// handles the different slash commands, and returns back a slack attachment with the body returned by the different functions.
func Main(in Request) {
	api := slack.New(auth_tok, slack.OptionDebug(true), slack.OptionAppLevelToken(app_tok))
	client := socketmode.New(
		api,
		socketmode.OptionDebug(true),
	)

	c, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func(c context.Context, api *slack.Client, client *socketmode.Client) (*Response, error) {
		for {
			select {
			case <-c.Done():
				return &Response{StatusCode: http.StatusBadRequest}, ErrClosed
			case event := <-client.Events:
				switch event.Type {
				case socketmode.EventTypeSlashCommand:
					command, ok := event.Data.(slack.SlashCommand)
					if !ok {
						continue
					}
					client.Ack(*event.Request)
					var err error
					switch command.Command {
					case "/emails":
						err = handleEmail(in, command, *api)
					case "/sms":
						err = handleSMS(in, command, *api)
					case "/url":
						err = handleURL(in, command, *api)
					}
					if err != nil {
						return &Response{StatusCode: http.StatusBadRequest}, err
					}
					return &Response{
						StatusCode: http.StatusOK,
						Body:       "success",
					}, nil
				}
			}
		}
	}(c, api, client)
	client.Run()
}

func handleEmail(in Request, command slack.SlashCommand, api slack.Client) error {
	params := &slack.Msg{Text: command.Text}
	str := strings.Split(params.Text, " ")
	temp := deleteEmpty(str)
	from, to, subject, content := temp[0], temp[1], temp[2], temp[3:]
	contentstr := strings.Join(content, " ")
	url = (fmt.Sprintf("%s/sendgrid-email", url))

	payload := Email{
		From:    from,
		To:      to,
		Subject: subject,
		Message: contentstr,
	}
	json, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(json))
	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return ErrRequest
	}
	createMessage(in, api)
	return nil
}

func handleSMS(in Request, command slack.SlashCommand, api slack.Client) error {
	params := &slack.Msg{Text: command.Text}
	str := strings.Split(params.Text, " ")
	temp := deleteEmpty(str)
	from, to, msg := temp[0], temp[1], temp[2:]
	msgstr := strings.Join(msg, " ")
	url = (fmt.Sprintf("%s/twilio-sms", url))

	payload := Sms{
		Number:  from,
		To:      to,
		Message: msgstr,
	}
	json, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(json))
	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return ErrRequest
	}
	createMessage(in, api)
	return nil
}

func handleURL(in Request, command slack.SlashCommand, api slack.Client) error {
	params := &slack.Msg{Text: command.Text}
	str := strings.Split(params.Text, " ")
	temp := deleteEmpty(str)
	filename, request, duration := temp[0], temp[1], temp[2]
	url = (fmt.Sprintf("%s/presigned-url", url))

	payload := Url{
		Filename: filename,
		Type:     request,
		Duration: duration,
	}
	json, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(json))
	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return ErrRequest
	}
	createMessage(in, api)
	return nil
}

func deleteEmpty(s []string) []string {
	var temp []string
	for _, str := range s {
		if str != "" {
			temp = append(temp, str)
		}
	}
	return temp
}

func createMessage(in Request, api slack.Client) {
	code := strconv.Itoa(in.StatusCode)
	attachment := slack.Attachment{
		Color: "#0069ff",
		Fields: []slack.AttachmentField{
			{
				Title: (fmt.Sprintf("Response: %s", code)),
				Value: in.Content,
			},
		},
		Footer: "DigitalOcean" + " | " + time.Now().Format("01-02-2006 3:4:5 MST"),
	}
	_, timestamp, err := api.PostMessage(
		channelid,
		slack.MsgOptionAttachments(attachment),
		slack.MsgOptionAsUser(true),
	)
	if err != nil {
		log.Fatalf("%s\n", err)
	}
	log.Printf("Message successfully sent at %s\n", timestamp)
}
