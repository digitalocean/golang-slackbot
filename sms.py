import os
from twilio.rest import Client
from twilio.base.exceptions import TwilioRestException

def valid_number(number, client):
    '''
    Returns true if the number given is a valid twilio phone number.

        Parameters:
            number: Phone number provided to validate
            client: Connects to Twilio API

        Returns:
            boolean: True if valid phone number, false if invalid phone number
    '''
    try: 
        client.lookups.phone_numbers(number).fetch(type = "carrier")
        return True
    except TwilioRestException as e:
        return False

def sendSMS(number, user_to, body):
    '''
    Takes in the phone numbers and message to send an sms using Twilio, 
    returns success or failed letting the user know if the message sent or failed to send.

        Parameters:
            args: Contains the from phone number, to phone number, and message to send

        Returns:
            string: Success or failed string f the message sent successfully or if an error happened
    '''
    sid = os.getenv('TWILIO_ACCOUNT_SID')
    token = os.getenv('TWILIO_AUTH_TOKEN')

    if not number or not user_to or not body:
        return "failed"

    client = Client(sid, token)
    if valid_number(number, client) and valid_number(user_to, client):
        client.messages.create(
            body = body,
            from_ = number,
            to = user_to
        )
        return "success"
    else:
        return "failed"