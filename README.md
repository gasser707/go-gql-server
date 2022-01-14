# Shotify

#### Shotify is an image sharing and selling application. It is a self-learning project to familiarize my-self with the Go programing language through an advanced project. It is built with the following technologies: 

- Golang 
- GraphQL
- MySQL
- Google Kubernetes Engine
- Google Cloud Storage
- Github Actions
- testify

### What are the available features? 

#### Authentication
- Secure authentication with JWT tokens, refresh tokens, cookies encrypted with [gorilla/securecookie](https://github.com/gorilla/securecookie) and CSRF tokens.
- Session storage using Redis.
- Secure password reset by emailing eset link.
- Email verification on signup by sending an account confirmation email.
#### Images
- CRUD operations on items 
- Creating several images at the same time by concurrency using **Go Channels and Routines**
- Deleting several images at the same time
- Buying images 
- Selling images
- Discounting images
- Archiving images
- Setting images as private
- Saving images to Google Cloud Storage
- Autogenerating labels or tags for images by using Google Cloud Vision
- Searching for an image by an image.
- Powerful image search that lets users search for images by several filters such as:
    * id
    * userId
    * title
    * labels (with option to match images that have **all** labels sent )
    * private (in images a user owns)
    * forSale 
    * priceLimit 
    * archived(in images a user owns)
    * discountPercentLimit
    * Search by uploading another image.

#### Resource protection

 User can only use update and delete operations on images they own, and they can search or filter images that aren't archived or private unless they previously bought them when they were public.

#### Emails

Users are sent confirmation emails to confirm their emails are valid, they can't access resources validating their emails. 
In dev environment. Emails are sent to a mailhog server that I set up as port of the docker-compose and the kubernetes cluster. In prod environment, they are sent using SendGrid.

### How to test it?

This project was built with a kubernetes development workflow using [Skaffold](https://skaffold.dev/).
To test it locally you can do so using either Kubernetes or Docker-Compose. You will also need a GCS bucket and 
your own GCP service account credentials as json. provide the path to your keys json and bucket name as environment variables. See the [.env.sample](./backend/.env.sample)

- Testing using skaffold:
    * Fill in the values of the [.env.sample file](./backend/.env.sample) and rename it to `.env`
    * Run `skaffold dev` and eveything should work. Provided you have kubernetes installed.
    * Run `kubectl get ingress` to see where what is the ip-address of your ingress. in your `etc/hosts` file on your system,
    add the the lines `<ingress ip>  shotify.com` at the bottom.

- Testing using docker-compose:
    * Run `docker-compose up`. The database with its correct schema are already mounted in a virtual volume so don't worry about migrations.


Look at the [schema](./backend/graphql/schemas) to see how to test the api.

# Design Discussion

###### This project was built with the following important design principles in mind:
- The clean architecture design pattern by [Robert C. Martin](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html):
  * Separating the application into layers such that every layer can use only the next layer below it but not the layer above it.
  * Structuring the layers so that they are technology and framework agnostic, for example when uploading an image to a cloud storage provider this is kept vendor agnostic in the following way:
 ```
      type StorageOperatorInterface interface {
      UploadImage(img io.Reader, imgName string, productId string) (url string, err error)
      DeleteImage(path string) error
    }

    type GcsClient struct {
      client *gcs.Client
    }

    type storageOperator struct {
      storageClient StorageOperatorInterface
    }

    type ProductsService struct {
      repo            repos.ProductsRepoInterface
      storageOperator StorageOperatorInterface
      thumbnailMaker  image.ThumbnailMakerInterface
    }
 ```
`GcsClient` just needs to implement the interface so it is used by the items or products service, this allows us to easily change storage providers without touching code in the service layer

- The GRASP patterns:
  * Having low coupling between different services but high cohesion within every service.
  * Creating pure fabrication objects to remove direct dependencies between services.
- Testability:
  * Using interfaces and adaptors to facilitate mocking of objects while performing unit tests.
- Scalability:
  * Achieved through separation of concerns that make it easy to remove a service from the application and use it as a microservice of its own with as little as possible refactoring and configuration.

### DataLoader and the N+1 problem
You might be wondering why I created the manufacturers entity, this is done to have another service to demonstrate how the application structure would scale when adding new services or features to the app. Also, I wanted to have GraphQL types referencing each other to demonstrate implementation of **DataLoader** to solve the ***GraphQL N+1 problem***

### Testing
Testing is a very important part of any production-ready application. I created unit tests using [Testify](https://github.com/stretchr/testify) and [mockery](https://github.com/vektra/mockery)

### CI/CD
Any production-ready application must have a CI/CD pipeline that ensures all tests pass before merging code to the main branch and ensure continuous deployment of the application. I used Github Actions for this project you can see my workflows in the [.github/wokflows](./.github/workflows) folder. The deployment is made to GKE.

### Monitoring 
- I have deployed the Prometheus-Grafana-AlertManager monitoring stack to see our clusters resources you can see the files in [infra/monitoring](./infra/monitoring)
  ![Grafana](https://i.imgur.com/2xLOoch.png)
  ![Grafana](https://i.imgur.com/gjZMWnj.png)
  
  
- I have also configured AlertManager to send notifications about incidents to a slack channel
 ![slack channel](https://i.imgur.com/L3a1pMz.png)
 
 
- Also the AlertManager WatchDog was configured to send health checks to [https://healthchecks.io/](https://healthchecks.io/) instead of the slack channel to avoid spamming the channel with alive ping notification.
  ![Health-check](https://i.imgur.com/mpmnkYd.png)
  
  
- Finally, Important incidents are also configured to be sent by text to those concerned by SMS using Zenduty.
 ![zenduty text](https://i.imgur.com/P3ofwzX.jpeg)
  
### A few Kubernetes notes
I deployed MySQL as a Replicated Statefulset application with a primary node and two secondary nodes. The two secondary nodes have Xtrabackup sidecars on them that clone the data from the primary node. Generally, I gave all the deployments **Burstable** Quality of Service (QoS) as this is just a demo.