import os
from sendgrid import SendGridAPIClient
from sendgrid.helpers.mail import Mail

def sendEmail(user_from, user_to, user_subject, content):
    '''
    Takes in the email address, subject, and message to send an email using SendGrid, 
    returns a string letting the user know if the email sent or failed to send.

        Parameters:
            args: Contains the from email address, to email address, subject and message to send

        Returns:
            string: Success or failed string f the message sent successfully or if an error happened
    '''
    key = os.getenv('API_KEY')

    if not user_from or not user_to or not user_subject or not content:
        return "failed"
    
    message = Mail(
        from_email = user_from,
        to_emails = user_to,
        subject = user_subject,
        html_content = content)

    sg = SendGridAPIClient(key)
    response = sg.send(message)
    if response.status_code != 202:
        return "failed"
    return "success"