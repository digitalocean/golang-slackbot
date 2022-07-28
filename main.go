package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
	"golang.org/x/sync/errgroup"
)

// SlackRequest takes in the StatusCode and Content from other functions to display to the user's slack.
type SlackRequest struct {
	// StatusCode is the http code that will be returned back to the user.
	StatusCode int `json:"statusCode"`
	// Content will contain the presigned url, error messages, or success messages.
	Content string `json:"body"`
}

// SlackResponse returns back the http code, type of data, and the presigned url to the user.
type SlackResponse struct {
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
	From string `json:"from"`
	// To is the email address that the user wants to send to.
	To string `json:"to"`
	// Subject is the subject of the email that the user wants to send.
	Subject string `json:"subject"`
	// Message is the body of the email that the user wants to send.
	Message string `json:"message"`
}

// Sms is the json struct of the sms message that contains the phone numbers and message.
type Sms struct {
	// Number is the user's phone number that they are sending from.
	Number string `json:"number"`
	// To is the phone number they are sending a message to.
	To string `json:"to"`
	// Message is the body of the message that the user wants to send.
	Message string `json:"message"`
}

// Url is the json struct of the presigned url function that contains the filename, type, and duration.
type Url struct {
	// Filename is the name of the file that will be uploaded or downloaded.
	Filename string `json:"filename"`
	// Type is a presigned request type to "GET" or "PUT" an object.
	Type string `json:"type"`
	// Duration is the duration in which the presigned url will last.
	Duration string `json:"duration"`
}

// funcResponse takes in the StatusCode and Content from other functions to display to the user's slack.
type funcResponse struct {
	// StatusCode is the http code that will be returned back to the user.
	StatusCode int `json:"statusCode"`
	// Content will contain the presigned url, error messages, or success messages.
	Body string `json:"body"`
}

var (
	authToken, appToken, channelid, url string
	// ErrNotEnoughArgs will return an error if the user does not provide the right number of arguments.
	ErrNotEnoughArgs = errors.New("not enough arguments provided")
)

const (
	// EmailsCommand is the slash command to send an email.
	EmailsCommand = "/emails"
	// SmsCommand is the slash command to send sms.
	SmsCommand = "/sms"
	// UrlCommand is the slash command to get a presigned url.
	UrlCommand = "/url"
)

func init() {
	authToken = os.Getenv("AUTH_TOKEN")
	if authToken == "" {
		panic("no authToken provided")
	}
	appToken = os.Getenv("APP_TOKEN")
	if appToken == "" {
		panic("no appToken provided")
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

// main configures a client with socketmode and slacks API using the token, app token, and channelid that the slack bot will be in,
// handles the different slash commands, and returns back a slack attachment with the body returned by the different functions.
func main() {
	var eg errgroup.Group
	api := slack.New(authToken, slack.OptionDebug(true), slack.OptionAppLevelToken(appToken))
	client := socketmode.New(
		api,
		socketmode.OptionDebug(true),
	)

	c, cancel := context.WithCancel(context.Background())
	defer cancel()

	eg.Go(func() error {
		for {
			select {
			case <-c.Done():
				return nil
			case event := <-client.Events:
				switch event.Type {
				case socketmode.EventTypeSlashCommand:
					command, ok := event.Data.(slack.SlashCommand)
					if !ok {
						continue
					}
					client.Ack(*event.Request)
					var (
						err          error
						slackRequest *SlackRequest
					)
					switch command.Command {
					case EmailsCommand:
						emailResponse, err := handleEmail(command)
						if err != nil {
							fmt.Fprintf(os.Stderr, "error handling email response: %s\n", err.Error())
						}
						slackRequest = &SlackRequest{
							StatusCode: emailResponse.StatusCode,
							Content:    emailResponse.Body,
						}
					case SmsCommand:
						smsResponse, err := handleSMS(command)
						if err != nil {
							fmt.Fprintf(os.Stderr, "error handling sms response: %s\n", err.Error())
						}
						slackRequest = &SlackRequest{
							StatusCode: smsResponse.StatusCode,
							Content:    smsResponse.Body,
						}
					case UrlCommand:
						urlResponse, err := handleURL(command)
						if err != nil {
							fmt.Fprintf(os.Stderr, "error handling url response: %s\n", err.Error())
						}
						slackRequest = &SlackRequest{
							StatusCode: urlResponse.StatusCode,
							Content:    urlResponse.Body,
						}
					default:
						slackRequest = &SlackRequest{
							StatusCode: 404,
							Content:    "command not found",
						}
					}
					err = makeRequest(slackRequest, api)
					if err != nil {
						fmt.Fprintf(os.Stderr, "error sending slack attachment: %s\n", err.Error())
					}
				}
			}
		}
	})
	eg.Go(func() error {
		return client.Run()
	})

	err := eg.Wait()
	if err != nil {
		log.Fatal(err)
	}
}

func handleEmail(command slack.SlashCommand) (*funcResponse, error) {
	params := &slack.Msg{Text: command.Text}
	str := strings.Split(params.Text, " ")
	temp := deleteEmpty(str)
	if len(temp) < 4 {
		resp := &funcResponse{
			StatusCode: http.StatusBadRequest,
		}
		return resp, ErrNotEnoughArgs
	}
	from, to, subject, content := temp[0], temp[1], temp[2], temp[3:]
	contentstr := strings.Join(content, " ")
	emailUrl := fmt.Sprintf("%s/sendgrid-email/sample/emails", url)

	payload := Email{
		From:    from,
		To:      to,
		Subject: subject,
		Message: contentstr,
	}
	json, err := json.Marshal(payload)
	if err != nil {
		resp := &funcResponse{
			StatusCode: http.StatusInternalServerError,
		}
		return resp, err
	}
	req, err := http.NewRequest(http.MethodPost, emailUrl, bytes.NewBuffer(json))
	if err != nil {
		resp := &funcResponse{
			StatusCode: http.StatusInternalServerError,
		}
		return resp, err
	}
	req.Header.Add("content-type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		resp := &funcResponse{
			StatusCode: http.StatusInternalServerError,
		}
		return resp, err
	}
	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		resp := &funcResponse{
			StatusCode: http.StatusInternalServerError,
		}
		return resp, err
	}
	resp := &funcResponse{
		StatusCode: res.StatusCode,
		Body:       string(bytes),
	}
	defer res.Body.Close()
	return resp, nil
}

func handleSMS(command slack.SlashCommand) (*funcResponse, error) {
	params := &slack.Msg{Text: command.Text}
	str := strings.Split(params.Text, " ")
	temp := deleteEmpty(str)
	if len(temp) < 3 {
		resp := &funcResponse{
			StatusCode: http.StatusBadRequest,
		}
		return resp, ErrNotEnoughArgs
	}
	from, to, msg := temp[0], temp[1], temp[2:]
	msgstr := strings.Join(msg, " ")
	smsUrl := fmt.Sprintf("%s/twilio-sms/sample/sms", url)

	payload := Sms{
		Number:  from,
		To:      to,
		Message: msgstr,
	}
	json, err := json.Marshal(payload)
	if err != nil {
		resp := &funcResponse{
			StatusCode: http.StatusInternalServerError,
		}
		return resp, err
	}
	req, err := http.NewRequest(http.MethodPost, smsUrl, bytes.NewBuffer(json))
	if err != nil {
		resp := &funcResponse{
			StatusCode: http.StatusInternalServerError,
		}
		return resp, err
	}
	req.Header.Add("content-type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		resp := &funcResponse{
			StatusCode: http.StatusInternalServerError,
		}
		return resp, err
	}
	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		resp := &funcResponse{
			StatusCode: http.StatusInternalServerError,
		}
		return resp, err
	}
	resp := &funcResponse{
		StatusCode: res.StatusCode,
		Body:       string(bytes),
	}
	defer res.Body.Close()
	return resp, nil
}

func handleURL(command slack.SlashCommand) (*funcResponse, error) {
	params := &slack.Msg{Text: command.Text}
	str := strings.Split(params.Text, " ")
	temp := deleteEmpty(str)
	if len(temp) < 3 {
		resp := &funcResponse{
			StatusCode: http.StatusBadRequest,
		}
		return resp, ErrNotEnoughArgs
	}
	filename, request, duration := temp[0], temp[1], temp[2]
	preUrl := fmt.Sprintf("%s/presigned-url/presign/url", url)

	payload := Url{
		Filename: filename,
		Type:     request,
		Duration: duration,
	}
	json, err := json.Marshal(payload)
	if err != nil {
		resp := &funcResponse{
			StatusCode: http.StatusInternalServerError,
		}
		return resp, err
	}
	req, err := http.NewRequest(http.MethodPost, preUrl, bytes.NewBuffer(json))
	if err != nil {
		resp := &funcResponse{
			StatusCode: http.StatusInternalServerError,
		}
		return resp, err
	}
	req.Header.Add("content-type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		resp := &funcResponse{
			StatusCode: http.StatusInternalServerError,
		}
		return resp, err
	}
	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		resp := &funcResponse{
			StatusCode: http.StatusInternalServerError,
		}
		return resp, err
	}
	resp := &funcResponse{
		StatusCode: res.StatusCode,
		Body:       string(bytes),
	}
	defer res.Body.Close()
	return resp, nil
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

func makeRequest(in *SlackRequest, api *slack.Client) error {
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
	_, _, err := api.PostMessage(
		channelid,
		slack.MsgOptionAttachments(attachment),
		slack.MsgOptionAsUser(true),
	)
	if err != nil {
		return err
	}
	return nil
}
