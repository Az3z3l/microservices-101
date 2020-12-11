from flask import Flask, render_template, request
from flask_mail import Mail, Message
import os

# instantiate flask app
app = Flask(__name__)

# set configuration and instantiate mail
mail_settings = {
    "MAIL_SERVER": 'smtp.office365.com',
    "MAIL_PORT": 587,
    "MAIL_USE_TLS": True,
    "MAIL_USE_SSL": False,
    "MAIL_USERNAME": "yaswantr@am.students.amrita.edu",
    "MAIL_PASSWORD": os.environ.get('outlookpwd'),

}
app.config.update(mail_settings)
mail = Mail(app)

# create message

@app.route("/", methods=["POST"]) 
def index(): 
    email = request.form.get('email')
    name = request.form.get("name")
    msg = Message(  
                    subject="FileUpload Service - Registration Successful",
                    sender ='yaswantr@am.students.amrita.edu', 
                    recipients = [email] 
                ) 
    msg.html = render_template('template.html', username=name)

    mail.send(msg)
    return 'Sent'

if __name__ == '__main__': 
   app.run(port=80, host='0.0.0.0') 