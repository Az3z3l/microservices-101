# File Upload Service

A fileupload-as-a-service that uses microservices to run built as a part of Cloud-Computing course. 
<br>

### To deploy as docker-compose: 
 - Create a file `secrets/vars.env` with your mongo password and email creds
 - docker-compose up
<br>

### To deploy as kubernetes in gcp:
 - Refer [HowTo.md](./HowTo.md)
<br>


### Tech-Stack Used:
- Golang
- Python
- Nginx(LoadBalancing and ProxyPass)
- HTML and Vannila JS
<br>

### MicroServices present:
- Login
    - Golang 
    - Checks users presence and sets JWT-Token for authentication
- Registration
    - Golang
    - Registers a new user
- Mailer
    - Python3
    - Sends mail to user on successful registration
- FileManager
    - Golang
    - Upload, Delete, Download, Get status of files of current user
    - The details of current user is gathereed from the JWT-Token set during login
- Nginx
    - Handles sending requests to appropriate containers
    - Display frontend
<br>

Deployed at: http://34.71.80.196:1337/ 
