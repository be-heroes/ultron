# emma-ultron

The emma-ultron is a critical component of Project Ultron, designed to enhance the automation and intelligence of workload management within Kubernetes clusters. At its core, the webhook server is responsible for dynamically and automatically assigning labels to pods as they are provisioned. This auto-labeling mechanism enables Kubernetes to make more informed decisions about pod scheduling, resource allocation, and workload optimization. By applying labels based on a variety of important criteria, the webhookserver ensures that workloads are efficiently distributed across the cluster.

The labels applied by the emma-ultron-webhookserver can include a wide range of information:

- **Pod workload type**: Labels can categorize workloads as batch processing, real-time applications or AI/ML jobs, allowing for better differentiation and handling of diverse workload types on available nodes.

- **Resource requirements**: Labels reflecting the specific CPU, memory or other resource needs of a pod can guide more accurate scheduling and ensure that pods are placed on nodes with appropriate available resources.

- **Priority and custom labels**: Users can define custom labels to express specific organizational or operational priorities, such as higher-priority workloads or specialized workflows, enabling fine-tuned control over workload handling.

- **Node-specific labels**: These labels are applied to optimize pod placement based on node characteristics, ensuring that pods are scheduled on nodes that can provide the best performance or cost efficiency for the given workload.

This automatic labeling system is crucial for Kubernetes clusters that rely on autoscaling mechanisms like karpenter or cluster autoscaler, where nodes and pods are frequently provisioned or de-provisioned. By seamlessly integrating into these dynamic environments, the emma-ultron plays a key role in optimizing resource usage and improving overall cluster performance.

## Prerequisites

- Go 1.23 or higher
- Docker (if you want to run the application in a container)

## Environment Variables

The application requires the following environment variables to be set:

- `EMMA_CLIENT_ID`: Your Emma API client ID
- `EMMA_CLIENT_SECRET`: Your Emma API client secret

## Installation

### Clone the repository

```sh
git clone https://github.com/emma-community/emma-ultron-webhookserver
cd emma-ultron-webhookserver
```

### Set up environment variables

```sh
export EMMA_CLIENT_ID=your_client_id
export EMMA_CLIENT_SECRET=your_client_secret
```

### Build the application

```sh
go build -o main main.go
```

### Run the application

```sh
./main
```

## Docker

To build and run the application using Docker.

### Build the Docker image

```sh
docker build -t emma-ultron:latest .
```

### Run the Docker container

```sh
docker run -e EMMA_CLIENT_ID=your_client_id -e EMMA_CLIENT_SECRET=your_client_secret emma-ultron:latest
```

## Additional links

- [Project Ultron => Abstract](https://github.com/emma-community/emma-ultron-operator/blob/main/docs/ultron_abstract.md)
- [Project Ultron => Algorithm](https://github.com/emma-community/emma-ultron-operator/blob/main/docs/ultron_algorithm.md)
- [Project Ultron => WebHookServer Sequence Diagram](https://github.com/emma-community/emma-ultron-operator/blob/main/docs/ultron.png)
