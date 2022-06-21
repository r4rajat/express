<h1 align="center">Express - Expose Ingress Custom k8s Controller</h1>

---


## 📝 Table of Contents

- [About](#about)
- [Getting Started](#getting_started)
- [Running the Code](#run)
- [Authors](#authors)
- [Acknowledgments](#acknowledgement)

## 🧐 About <a name = "about"></a>

The Express custom kubernetes controller is written primarily in go lang. This controller explicitly keeps a watch on newly created Deployments in all Namespaces,<br>
And as soon as a new Deployment is created, our controller will create a Service and an Ingress for it and expose it to the outer world.

## 🏁 Getting Started <a name = "getting_started"></a>

These instructions will get you the project up and running on your local machine for development and testing purposes. See [Running the Code](#run) for notes on how to deploy the project on a Local System or on a Kubernetes Server.

### Prerequisites

To run the Express Controller on Local System, first we need to install following Software Dependencies.

- [Go](https://go.dev/dl/)
- [Docker](https://docs.docker.com/get-docker/)
- [Minikube](https://minikube.sigs.k8s.io/docs/start/)

Once above Dependencies are installed we can move with [further steps](#installing)

### Installing <a name = "installing"></a>

A step by step series of examples that tell you how to get a development env running.


#### Step 1: Running a 2 Node Mock Kubernetes Server Locally using minikube
```
minikube start --nodes 2 -p k8s-cluster
```

#### Step 2: Enabling ingress in minikube
```
minikube addons enable ingress -p k8s-cluster
```

#### Step 3: Setting Up Environmental Variables

Set up the Environmental variables according to your needs. The Application will run with defaults as mentioned in the following table

| Environmental Variable | Usage                               | Default Values |
|------------------------|-------------------------------------|----------------|
| EXPRESS_QUEUE          | Queue for holding interface objects | EXPRESS        |


## 🔧 Running the Code <a name = "run"></a>

To Run the Express Controller on local machine, Open a terminal in the Project and run following command
```
go build
```
```
./express
```

---

To Run the Express Controller inside k8s cluster, follow the below steps

1. Create the Role having permissions to list, create, edit and delete the resources in all namespaces
```
kubectl create role allowAll --resource deployments,pods,services,ingresses --verb list,create,delete
```
2. Create Role Binding of above role with the default service account
```
kubectl create rolebinding allowAll --role allowAll --serviceaccount default:default
```
3. Create Namespace
```
kubectl create namespace express
```
4. Deploy the Controller with the Manifest File
```
kubectl create -f express.yaml
```

## ✍️ Authors <a name = "authors"></a>

- [@r4rajat](https://github.com/r4rajat) - Implementation

## 🎉 Acknowledgements <a name = "acknowledgement"></a>

- References
    - https://pkg.go.dev/k8s.io/client-go
    - https://pkg.go.dev/k8s.io/apimachinery
    - https://pkg.go.dev/github.com/mitchellh/go-homedir
    - https://kubernetes.github.io/ingress-nginx/
