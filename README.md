# Composed Function: Go "Slack Bot"

## Introduction: Composed Function

This repository contains a slack bot written in Golang. The slack bot uses socketmode and runs indefinitely when hosted somewhere, like Docker & DigitalOcean's Droplets. The slack bot takes in 3 slash commmands: presigned url, twilio sms, and sendgrid email, which are all explained individually below. It will return back a slack attachment to the channel including either a presigned url, success message, or an error message. 

### Introduction: 3 Sample Functions

This repository contains a sample presigned URL function written in Golang. You are able to choose to get a presigned URL to upload a file to a DigitalOcean Space or to download a file from a DigitalOcean Space. Given the filename, request (`GET` or `PUT`), and duration (how many seconds/minutes/hours you want the url to be valid for), a presigned url gets returned to you or else an error message.

This repository contains a sample Twilio SMS function written in Python. You are able to send an sms using only Twilio verified phone numbers. Given 2 phone numbers and a message, a successful message gets returned or else an error message.

This repository contains a sample Sendgrid Emails function written in Python. You are able to send an email using Sendgrid's API to email addresses with or without DMARCS. Given 2 email addresses, a subject, and a body, a successful message gets returned or an else an error message.

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

### Presigned URL

* You need a DigitalOcean account. If you don't already have one, you can sign up at [https://cloud.digitalocean.com/registrations/new](https://cloud.digitalocean.com/registrations/new).
* You need a DigitalOcean Space. If you don't have one, you can create one at https://www.digitalocean.com/products/spaces.
* You need to add your `SPACES_KEY`, `SPACES_SECRET`, `BUCKET`, and `REGION` to the `.env` file to connect to Spaces API as well as your bucket.

### Sendgrid Emails

* You need a SendGrid account. If you don't already have one, you can sign up at https://signup.sendgrid.com/.
* You need to create a SendGrid API key with Full Access to connect to your sendgrid account. You can learn more about it at https://docs.sendgrid.com/ui/account-and-settings/api-keys.
* You need to add your `API_KEY` to the `.env` file to connect to the SendGrid API.
* To send emails with valid email addresses, you have to set up sender authentication. You can learn more at https://docs.sendgrid.com/glossary/sender-authentication.
* To send emails to email address with DMARCS, you have to set up domain authentication. You can learn more at https://docs.sendgrid.com/ui/account-and-settings/how-to-set-up-domain-authentication.

### Twilio SMS

* You need a Twilio account. If you don't already have one, you can sign up at https://www.twilio.com/try-twilio.
* To send and receive sms with Twilio, you need a Twilio virtual phone number. You can learn more at https://www.twilio.com/docs/phone-numbers.
* The phone number you are sending a message to also has to be a twilio verified phone number.
* You need to add your `TWILIO_ACCOUNT_SID` and `TWILIO_AUTH_TOKEN` to the `.env` file to connect to the Twilio API.

## Hosting and Deploying

* I've added a simple DockerFile if you'd like to create a docker container.
* You can host your app on a DigitalOcean Droplet. More information can be found at https://www.digitalocean.com/products/droplets.

## Using the Slash Commands

NOTE: If you do not use the correct number of arguments, slackbot will give you the following error and shut down socketmode:
```bash
/command failed with the error "dispatch_failed"
```
Please be sure to enter the correct number of arguements for each slash command.

### To send a SMS

```bash
/sms [Twilio number] [Receivers number] [Message]
```

### To send an email

```bash
/emails [Your email] [Receivers email] [Subject] [Message]
```

### To get a presigned url

```bash
/url [Filename] [Request] [Duration]
```

### To Upload or Download a file using curl in terminal:
```
curl -X PUT -d 'The contents of the file.' "{url}"
```