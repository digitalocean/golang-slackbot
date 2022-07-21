# Composed Function: Go "Slack Bot"

## Introduction: Composed Function

This repository contains a slack bot written in Golang. The slack bot uses socketmode and will send back a slack attachment to the channel your bot is in with the status code and body of the returned response. You can deploy it on DigitalOcean's App Platform as a Serverless Function component. Documentation is available at https://docs.digitalocean.com/products/functions.

The slack bot takes in 3 slash commmands: presigned url, twilio sms, and sendgrid email. It will return back a slack attachment to the channel including the status code and body, which contains either a success message, error message, or presigned url. 

## Requirements

### Slack Bot

* You need to create a slack app and enable SocketMode. You can learn more at https://api.slack.com/apis/connections/socket. 
* I've attached a `manifest.yml` file if you'd like to set up the app with manifests. The manifest file contains the scopes of the bot you will be adding to allow the bot to read and write to the channel as well as accept slash commands.
* You need to add your `AUTH_TOKEN`, `APP_TOKEN`, and `CHANNEL_ID` to the `.env` file to connect to the Slack API as well as your channel that you will be adding the slack bot to.

NOTE: If you decide not to use the `manifest.yml` to create an app, here are some things you need to make sure are activated for your app.
- Activate Socket Mode
- Activate Interactivity & Shortcuts
- Activate Incoming Webhooks
- Add Slash Commands
    - /emails
        - usage hint: [Your email] [Receiver's email] [Subject] [Message]
    - /sms
        - usage hint: [Twilio number] [Receiver's number] [Message]
    - /url
        - usage hint: [Filename] [Request] [Duration]
- Bot Token Scopes
    - `app_mentions:read` : view messages that directly mention your app
    - `chat:write` : send messages via your app
    - `commands` : using slash commands
    - `incoming-webhook` : post attachments & messages to channels


## Using the Slash Commands in Slack

NOTE: If you do not use the correct number of arguments, slackbot will give you the following error and shut down socketmode:
```bash
/command failed with the error "dispatch_failed"
```
Please be sure to enter the correct number of arguements for each slash command.

### To send a SMS using the Twilio API function

```bash
/sms [Twilio number] [Receivers number] [Message]
```

### To send an email using the SendGrid API Function

```bash
/emails [Your email] [Receivers email] [Subject] [Message]
```

### To get a presigned url from DigitalOcean's using the presigned URL function

```bash
/url [Filename] [Request] [Duration]
```

### To Upload or Download a file using curl in terminal:
```
curl -X PUT -d 'The contents of the file.' "{url}"
```