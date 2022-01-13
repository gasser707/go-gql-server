# Shotify

#### Shotify is my submission for the Shopify Backend Developer & Production Engineering Intern Challenge. It is built with the following technologies: 
- Golang 
- GraphQL
- MySQL
- Google Kubernetes Engine
- Google Cloud Storage
- Github Actions
- testify

#### I chose these technologies after analyzing Shopify's development stack. Even though Shopify's backend is built with Ruby, I chose Go as it is a simple and fast language.
>“Go will be the server language of the future.” — Tobias Lütke, Shopify CEO

### How to test it?
#### I didn't want the tester to go through a lot of environment configurations and set up just to run the app, so I deployed it to Google Kubernetes Engine, you can test the live version at https://www.go-shotify.xyz


### what are the available features? 
- CRUD operations on items 
- **Uploading an image for an item, generating its thumbnail, and uploading both images to GCS** <- _My chosen extra feature_
- Creating several items at the same time
- Deleting several items at the same time
- CRUD operations on manufacturers of items

#### Here are examples of the GraphQL api usage:
###### Get A list of items:
```
query{
  products{
    id
    name
    imageUrl
    thumbnailUrl
    createdAt
    updatedAt
    labels
    price
    discountPercent
  }
}
```
You can also view the information of the manufacturer of the item
```
query{
  products{
    id
    name
    imageUrl
    thumbnailUrl
    price
    discountPercent
    manufacturer{
      id
      name
      joinedAt
      bio
    }
  }
}
```
###### Create one or several item in one shot:
make sure that a manufacturer with the manufacturerId you are inserting exists before associating the item to the manufacturer

```
mutation($file1: Upload, $file2: Upload){
  createProducts(input:
  	[
      {
        name:"test",
        description: "test",
        price:25.5,
        discountPercent:0,
        manufacturerId:1, 
        labels: ["hi", "hi2"],
        image: $file1,
      },
        {
        name:"test",
        description: "test",
        price:25.5,
        discountPercent:0,
        manufacturerId:1,
        labels: ["hi", "hi2"],
        image: $file2,
      }
  ])
    {
      id
      name
      imageUrl
      thumbnailUrl
      manufacturer{
        id
        name
        joinedAt
      }
    }
}
```
###### I deployed a little Node.js express app that has a GraphQL IDE that makes uploading images a lot easier than using the normal playground or using multipart requests using something like Postman. Here I am showing how you would send the two images from the previous mutation with your request:

![Altair image upload](https://media.giphy.com/media/w7swOLCNBmasJNWpLQ/giphy.gif)

###### Update an item:

```
mutation {
  updateProduct(input: {
    id: 32,
    name: "updated",
    description:"description update",
    labels:  ["laptops","electronics"],
    price: 2220,
    discountPercent:0,
    manufacturerId:1    
  }){
    id,
    name, description, labels, price
  }
}
```
###### Delete one or several items by their ids in one shot:
```
mutation{
  deleteProducts(input:[1, 2])
}
```

###### Get A list of manufacturers:
```
query{
  manufacturers{
    id
    name
    bio
    joinedAt
  }
}
```
###### Create a manufacturer:

```
mutation {
  createManufacturer(input: { name: "2 Man", bio: "bio2" }) {
    id
  }
}

```
###### Update a manufacturer:

```
mutation{
  updateManufacturer(input: {
    Id:2,
    name:"updated manufacturer",
    bio: "updated bio"
  }){
    name
    bio
    id
  }
}
```
###### Delete a manufacturer (if they don't have any associated products):
```
mutation{
  deleteManufacturer(input: 1)
}
```
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

### TLS Encryprition
Any production App needs to have HTTPS and I obtained a TLS certificate for my cluster using **cert-manager** and **Let's Encrypt** for the **Nginx Ingress** reverse proxy.
  
  
