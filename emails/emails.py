# export API_KEY=XXXXXXXX

import os
import sys
from sendgrid import SendGridAPIClient
from sendgrid.helpers.mail import Mail

def main():
    key = os.getenv('API_KEY')
    args = sys.argv[1:]
    user_from = args[0]
    user_to = args[1]
    user_subject = args[2]
    content = ' '.join(args[3:])

    if not user_from:
        print("No user email given.")

    if not user_to:
        print("No to email given.")

    if not user_subject:
        print("No subject given.")

    if not content:
        print("No message given.")
    
    message = Mail(
        from_email = user_from,
        to_emails = user_to,
        subject = user_subject,
        html_content = content)

    # Sending email via Sendgrid API and informing user if email successfully sent or not.
    sg = SendGridAPIClient(key)
    response = sg.send(message)
    print(response.status_code)
    if response.status_code != 202:
        print("Email failed to send.\n")
        print("Please check if your email is authenticated and the email you are sending to is correct.")


if __name__ == "__main__":
    main()