# How to build a microservice?

## Stuff you need to know before that

### Dockers
- [Dockers](https://www.docker.com/resources/what-container) are containers for running your application. 
<br>

### Kubernetes
- [Kubernetes](https://kubernetes.io/docs/concepts/overview/what-is-kubernetes/) is a container orchestration system. 
<br>

### Docker-compose
- [Docker Compose](https://docs.docker.com/compose/) is a tool for defining and running multi-container Docker applications
<br>

### Microservice
- [Microservice](https://microservices.io/) is a way of app development such that the functionalities of the platform needed are broken down to several components. 
<br>


This app was built using docker-compose and then converted to a kubernetes based app using [Kompose](https://kompose.io/)


<br>

## How this works?

Each component/service holds one functionality. This service is run inside a docker. We do this for all the components we have. Now, how to connect it? Here we use nginx and docker-compose to do the trick. Nginx is a load-balancer and we can use it to proxy all the requests based on the url to the respective docker containers. One other thing is that, we give each docker services a name in the docker-compose.yaml file. This name can be used to retrieve the internal ip of the docker as well. So using Nginx and the names of the services we gave in docker-compose.yml, we can successfully send requests to the respective apps. This way, even if the register functionality does not work due to some fault in the code, the user could login successfully. This wont be possible in a monolithic application. 

<br>

## Converting Docker-Compose to Kubernetes understandable file:

This is fairly easy compared to the previous tasks. All we need to do is convert the docker-compose.yaml to a file that kubernetes would understand. This can be done using a tool called [kompose](https://kompose.io/setup/). To use `kompose`, we first need to push the local docker files to a registry. Refer [this](https://rollout.io/blog/using-docker-push-to-publish-images-to-dockerhub/) to push the dockerfiles. Then replace `build` in docker-compose with `image` and specify the location to download. Now, run `kompose convert -f docker-compose.yaml -o kube.yaml`. This will create a kubernetes file. The last thing to do before deploying is to explose our Nginx port. Add the line `type: LoadBalancer` below `Name: Nginx` and `Port: 80`. Voila! you have an app ready to be deployed. 


## Deploying to GCP:

Now that we have the kube.yaml file ready, deployment should be a piece of cake. Open GCP console and get your kube.yaml file in there. Then go to Kubernetes Engine --> Clusters. Create a new cluster(give default values if you're not sure what's what). After it is initialised click on `connect` and then click `Run in Cloud Shell`. This should add a command on your gconsole. Press enter. Now the cluster is ready to run your app. To load the kube file, run `kubectl apply -f kube.yaml`. Now, run `kubectl get svc` till gcp gives up and tells your app's ip next to Nginx's service name. 

Aaaaaaaaand you've deployed a microservice to cloud \o/

<br><br>

#### Other amazing microservices:
- [microservice built mainly using python](https://github.com/Captain-K-101/MicroServices/)
- [microservice built mainly using nodejs](https://github.com/pranjalsingh008/imager)
 

<br><br>

#### Resources:
- https://www.docker.com/why-docker
- https://docs.docker.com/develop/develop-images/dockerfile_best-practices/
- https://docs.docker.com/compose/gettingstarted/
- https://docs.nginx.com/nginx/admin-guide/web-server/reverse-proxy/
- https://www.youtube.com/watch?v=tndzLznxq40
- https://www.nginx.com/blog/introduction-to-microservices/
- https://kubernetes.io/docs/tasks/configure-pod-container/translate-compose-kubernetes/
