# Ultron Algorithm

Designing an algorithm for matching Kubernetes Pods to Nodes that factors in resource requests, limits, disk types, network types, node prices, and compute durability (spot vs. durable) can be quite complex. The goal is to optimize for both the workloads needs and cost-efficiency, all while adhering to constraints like available resources, labels, and node properties.

## Objectives

- Minimize cost by prioritizing cheaper Nodes (spot instances over durable if possible).
- Avoid over-subscription of resources (respect limits).
- Prioritize Pods with specific needs (e.g., high-priority Pods, Pods requiring certain disk types or network types).

## Constraints

- The Pods resource requests (CPU, memory, storage) must fit within the Nodes available resources.
- Disk type labels and network type labels on the Pod must match the Nodes attributes.
- The Nodes available resources should be able to satisfy the Pods limits (not just requests).

## Synopsis

This algorithm balances the constraints of resource requests/limits, disk/network types, and pricing (spot vs. durable nodes). It can be optimized by adjusting the weights based on specific cluster policies (e.g., cost optimization vs. performance). This allows Kubernetes to efficiently schedule Pods while adhering to workload requirements and cluster cost constraints. For each Pod, score each Node using a weighted scoring system. The higher the score, the better suited the Node is for the Pod.

### Required inputs

The algorithm would need the following inputs (and more) to calculate the most suitable Node for a given Pod:

- Pod.RequestedCPU: The CPU required by the Pod.
- Pod.RequestedMemory: The memory required by the Pod.
- Pod.RequestedStorage: The storage required by the Pod.
- Pod.RequestedDiskType: The disk type required by the Pod.
- Pod.RequestedNetworkType: The network type required by the Pod.
- Pod.LimitCPU: The maximum CPU the Pod can consume (ensures the Node can handle peaks).
- Pod.LimitMemory: The maximum RAM the Pod can consume (ensures the Node can handle peaks).
- Pod.Priority: Indicates the importance of the workload. High-priority Pods (e.g., latency-sensitive) should get more reliable resources (durable Nodes), while low-priority Pods (e.g., batch jobs) can use spot instances.
- Node.Type: Spot or durable. Spot Nodes are cheaper but risk preemption, while durable Nodes are more stable.
- Node.Price: The hourly cost of using the Node, with spot instances being cheaper but less reliable.
- Node.MedianPrice: The median price for the node over a defined time period (e.g., last week/month). This smooths out price fluctuations, making it a better indicator of long-term cost.
- Node.AvailableCPU: Available CPU that the Node can provide.
- Node.AvailableMemory: Available memory that the Node can provide.
- Node.AvailableStorage: Available storage that the Node can provide.
- Node.TotalCPU: The total CPU capacity the Node can provide.
- Node.TotalMemory: The total memory capacity the Node can provide.
- Node.TotalStorage: The total storage capacity the Node can provide.
- Node.DiskType: Type of storage offered by the Node (e.g., SSD or HDD). Pods requesting specific disk types (e.g., SSD) must match the Node’s disk type.
- Node.NetworkType: Network characteristics of the Node (e.g., low-latency or high-bandwidth). Pods with specific network requirements should be matched to suitable Nodes.
- Node.InterruptionRate: The likelihood of a node being preempted. Critical workloads can penalize Nodes with higher interruption rates.

#### Additional inputs

If available, data on Node performance, historical interruptions, and Pod scheduling efficiency can be fed into the algorithm to make better scheduling decisions over time.

### Scoring components

The scoring components of the algorithm are designed to evaluate how well available Nodes matches the requirements of a Pod, using several key factors. Each factor contributes to a Nodes overall score, guiding the decision-making process. The algorithm considers resource utilization (CPU, memory), disk and network types, Node pricing, stability (e.g., spot instance risks), and the priority of the workload. By assigning a weighted score to each Node based on these factors, the algorithm identifies the Node that offers the best balance between resource fit, cost efficiency, and workload needs, ensuring optimal scheduling.

#### ResourceFitScore (CPU/Memory)

Score how well the Pod’s resource requests/limits fit within the available resources on the Node.

```plaintext
ResourceFitScore = (Node.AvailableCPU - Pod.RequestedCPU) / Node.TotalCPU + (Node.AvailableMemory - Pod.RequestedMemory) / Node.TotalMemory
```

With regards to ranking, higher is better. Nodes that leave minimal excess resources after the Pod is scheduled should be preferred (i.e., efficient packing).

#### DiskTypeScore

If the Pod requests a specific disk type (e.g., SSD or HDD), assign a higher score to Nodes that offer the requested disk type. Binary score: 1 if the disk type matches, 0 otherwise.

```plaintext
DiskTypeScore = Node.DiskType == Pod.RequestedDiskType ? 1 : 0
```

#### NetworkTypeScore

Similarly, if the Pod requests specific network characteristics (e.g., low-latency), score the Nodes based on matching network attributes. Binary score: 1 if the network type matches, 0 otherwise.

```plaintext
NetworkTypeScore = Node.NetworkType == Pod.RequestedNetworkType ? 1 : 0
```

#### PriceScore (Spot vs Durable)

Prefer spot instances if the workload is fault-tolerant (e.g., batch jobs). Durable compute should be preferred for critical workloads. Use a normalized cost score based on node price. The cheaper the Node, the higher the score.

```plaintext
PriceScore = 1 - (Node.MedianPrice / Node.Price)
```

The lower the spot price, the higher the score.

#### NodeStabilityScore (Spot Instances)

Spot instances may be terminated, so a risk factor can be introduced for spot nodes. For critical workloads, add a penalty to spot instances to reduce their overall score. Risk factor score: For spot instances, apply a penalty based on historical interruption rates or likelihood of preemption.

```plaintext
NodeStabilityScore =  Node.InterruptionRate * (Node.Price / Node.MedianPrice)
```

Higher interruption rates result in lower scores, especially for high-cost spot instances.

#### WorkloadPriorityScore

Assign weights based on workload priority. High-priority workloads (latency-sensitive) might be biased toward durable nodes, while batch jobs can favor spot instances. Higher-priority Pods may receive a higher score for durable nodes.

```plaintext
WorkloadPriorityScore = Pod.Priority == HighPriority ? 1 : 0
```

### Score calculation

For each Node, calculate a total score:

```plaintext
Total Score = `α` * ResourceFitScore + 
              `β` * DiskTypeScore + 
              `γ` * NetworkTypeScore + 
              `δ` * PriceScore - 
              `ε` * NodeStabilityScore + 
              `ζ` * WorkloadPriorityScore
```

Where: α, β, γ, δ, ε, ζ are weights that adjust the importance of each factor. These can be tuned based on the specific workload or cluster requirements.

### Node Selection

- Filter Nodes: First, filter out any Nodes that cannot satisfy the basic constraints (e.g., insufficient CPU/memory, mismatched disk or network types).

- Sort Nodes: Sort the remaining Nodes based on their calculated total score.

- Select the Best Node: Assign the Pod to the Node with the highest score.

- No match found: If no Node satisfies the constraints (e.g., resource exhaustion), the Pod can be placed in a pending state or simply left unscheduled.

### Customization

This algorithm can be customized based on:

- **ClusterPolicies**: By adjusting the weights (α, β, γ, etc.), the algorithm can prioritize cost-efficiency (spot instances) or performance (high-priority workloads on durable instances).

- **WorkloadRequirements**: Different workloads (e.g., batch vs. latency-critical) can be handled by tuning the weighting system and introducing additional constraints (e.g., setting minimum availability requirements for critical apps).

This flexible design enables fine-tuning for a variety of workload needs, from high-performance latency-sensitive applications to fault-tolerant batch processing tasks.
