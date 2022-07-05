# export TWILIO_ACCOUNT_SID=XXXXXXXX && export TWILIO_AUTH_TOKEN=XXXXXXXXXXXXX

import os
import sys
from twilio.rest import Client
from twilio.base.exceptions import TwilioRestException

# Validates the phone numbers since only twilio validated numbers are allowed to send sms via twilio.
def valid_number(number, client):
    try: 
        client.lookups.phone_numbers(number).fetch(type = "carrier")
        return True
    except TwilioRestException as e:
        if e == 21606:
            return False
        else:
            raise e


def main():
    # Enter the account sid and auth token from users account.
    sid = os.getenv('TWILIO_ACCOUNT_SID')
    token = os.getenv('TWILIO_AUTH_TOKEN')
    args = sys.argv[1:]
    number = args[0]
    user_to = args[1]
    body = ' '.join(args[2:])

    if not number:
        print("Your phone number is required.")
    
    if not user_to:
        print("The receiver's phone number is required.")

    if not body:
        print("A message is required.")

    # Connecting to Twilio's client to send sms.
    client = Client(sid, token)
    if (valid_number(number, client) == True) & (valid_number(user_to, client) == True):
        client.messages.create(
            body = body,
            from_ = number,
            to = user_to
        )
        print("Message sent.")
    else:
        print("The phone numbers provided are not twilio verified numbers.")


if __name__ == "__main__":
    main()
