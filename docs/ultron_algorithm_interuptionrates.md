# Ultron Algorithm - Interuption Rates

Calculating interruption rates for Spot instances across all major cloud providers (such as AWS, Google Cloud Platform, and Microsoft Azure) involves tracking the likelihood or frequency at which a Spot instance is terminated due to factors like price increases or insufficient capacity. Here's a breakdown of how you would approach calculating or estimating these interruption rates:

## Obtain Provider-Specific Interruption Data

Each cloud provider handles Spot or Preemptible instances differently, and they may publish data or metrics related to instance interruptions. You will want to gather this data for each cloud provider:

### AWS Spot Instances

- AWS Spot Interruption Data: AWS provides some statistics on Spot instance interruptions via the AWS Status page or through their Spot Instance Advisor, which includes historical data on interruption rates for different instance types and regions.

- Metrics: Track metrics such as SpotInstanceInterruptionTerminated in AWS CloudWatch to understand how frequently your instances are being interrupted.

- Official Data Sources: AWS provides aggregated historical interruption rates, which can give you a good sense of how likely interruptions are depending on instance types and availability zones.

### Google Cloud Preemptible VMs

- Preemptible Instance Data: Google Cloud's Preemptible VMs can be interrupted after running for up to 24 hours, but the precise interruption rate may vary based on demand and resource availability.

- Metrics: You can monitor metrics such as compute.googleapis.com/instance/preempted_count via Google Cloud Monitoring to track how often preemption occurs for your instances.

- Official Data Sources: Google provides high-level insights, but interruption rates are more likely to be inferred from your instance’s preemption patterns.

### Microsoft Azure Spot VMs

- Azure Spot VM Data: Azure Spot VMs are similar to AWS Spot instances and Google Preemptible VMs. They can be evicted if there’s demand for standard pricing or capacity is low.

- Metrics: You can use Azure Monitor to track the eviction rate via Percentage VM Evictions or similar metrics provided by Azure.

- Official Data Sources: Azure also publishes data and tips for managing Spot VMs based on historical trends in different regions.


## Calculate/Estimate Interruption Rates

To calculate or estimate interruption rates for Spot instances, you would:

### Track Instance Termination Events

Use cloud provider monitoring tools (CloudWatch for AWS, Google Cloud Monitoring, Azure Monitor) to gather data on how often instances are interrupted.
Keep track of metrics such as termination or preemption counts.

### Analyze by Instance Type, Region, and Time Frame

Different instance types (e.g., general-purpose, compute-optimized) and regions have varying interruption rates.
Analyze your data to calculate the rate of interruptions for specific instances within a time frame, e.g., daily, weekly, or monthly.

### Historical Data & Provider Documentation

Utilize the historical data or recommendations from Spot Instance Advisors (AWS), Google Cloud’s Preemptible VM documentation, and Azure Spot VM pricing guides. These provide estimates of interruption rates.

### Calculate Interruption Rate

The basic formula to calculate interruption rate is:

Interruption Rate = (Number of Interrupted Instances / Total Running Spot Instances) × 100

This can be calculated for a specific instance type, region, or time period.
