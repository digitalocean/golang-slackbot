package main

import (
	"context"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"

	u "slackbot/url"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

	auth_tok := os.Getenv("AUTH_TOKEN")
	app_tok := os.Getenv("APP_TOKEN")
	channelid := os.Getenv("CHANNEL_ID")

	api := slack.New(auth_tok, slack.OptionDebug(true), slack.OptionAppLevelToken(app_tok))
	client := socketmode.New(
		api,
		socketmode.OptionDebug(true),
	)

	c, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func(c context.Context, api *slack.Client, client *socketmode.Client) {
		for {
			select {
			case <-c.Done():
				return
			case event := <-client.Events:
				switch event.Type {
				case socketmode.EventTypeSlashCommand:
					command, ok := event.Data.(slack.SlashCommand)
					if !ok {
						continue
					}
					client.Ack(*event.Request)
					err := handleSlashCommand(command, api, channelid)
					if err != nil {
						log.Fatal(err)
					}
				}
			}
		}
	}(c, api, client)
	client.Run()
}

func handleSlashCommand(command slack.SlashCommand, api *slack.Client, channelid string) error {
	switch command.Command {
	case "/emails":
		return handleEmail(command, *api, channelid)
	case "/sms":
		return handleSMS(command, *api, channelid)
	case "/url":
		return handleURL(command, *api, channelid)
	}
	return nil
}

func handleEmail(command slack.SlashCommand, api slack.Client, channelid string) error {
	params := &slack.Msg{Text: command.Text}
	str := strings.Split(params.Text, " ")
	temp := deleteEmpty(str)
	from, to, subject, content := temp[0], temp[1], temp[2], temp[3:]
	contentstr := strings.Join(content, " ")

	cmd := exec.Command("python3", "-c", "import emails; print(emails.sendEmail('"+from+"', '"+to+"', '"+subject+"', '"+contentstr+"'))")
	out, err := cmd.CombinedOutput()
	output := string(out)
	if err != nil {
		output = "failed"
	}
	createMessage(output, api, channelid)
	return nil
}

func handleSMS(command slack.SlashCommand, api slack.Client, channelid string) error {
	params := &slack.Msg{Text: command.Text}
	str := strings.Split(params.Text, " ")
	temp := deleteEmpty(str)
	number, to, msg := temp[0], temp[1], temp[2:]
	msgstr := strings.Join(msg, " ")

	cmd := exec.Command("python3", "-c", "import sms; print(sms.sendSMS('"+number+"', '"+to+"', '"+msgstr+"'))")
	out, err := cmd.CombinedOutput()
	output := string(out)
	if err != nil {
		output = "failed"
	}
	createMessage(output, api, channelid)
	return nil
}

func handleURL(command slack.SlashCommand, api slack.Client, channelid string) error {
	params := &slack.Msg{Text: command.Text}
	str := strings.Split(params.Text, " ")
	temp := deleteEmpty(str)
	filename, req, duration := temp[0], temp[1], temp[2]
	url := u.FindURL(filename, req, duration)
	createMessage(url, api, channelid)
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

func createMessage(out string, api slack.Client, channelid string) {
	if strings.TrimRight(out, "\n") == "success" {
		attachment := slack.Attachment{
			Color: "#93c572",
			Fields: []slack.AttachmentField{
				{
					Title: "MESSAGE SENT!",
				},
			},
			Footer: "DigitalOcean" + " | " + time.Now().Format("01-02-2006 15:04:05 MST"),
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
	} else if strings.TrimRight(out, "\n") == "failed" {
		attachment := slack.Attachment{
			Color: "#ba2507",
			Fields: []slack.AttachmentField{
				{
					Title: "ERROR!",
				},
			},
			Footer: "DigitalOcean" + " | " + time.Now().Format("01-02-2006 15:04:05 MST"),
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
	} else {
		attachment := slack.Attachment{
			Color: "#0069ff",
			Fields: []slack.AttachmentField{
				{
					Title: "Presigned url: ",
					Value: out,
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
}
