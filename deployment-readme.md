# Kubernetes AI Agent Operator – GKE Integration

## 1. Project Overview

The Kubernetes AI Agent Operator is intended to automate the deployment and lifecycle management of AI Agents in Google Kubernetes Engine (GKE).

The long-term goal is to provide a standard deployment mechanism where application developers only need to define an `AIAgent` Custom Resource (CR). The Operator handles the platform-specific deployment configuration, Helm values generation, GitHub integration, and deployment workflow through Argo CD.

Instead of every application team maintaining its own Helm charts and Kubernetes deployment manifests, the platform team provides reusable Helm templates, and the Operator generates the configuration required for each Agent.

### Main objectives

- Deploy AI Agents into GKE through a Kubernetes Operator.
- Generate Helm values automatically from an `AIAgent` Custom Resource.
- Support standard application and network configurations.
- Integrate with GitHub to publish generated Helm configuration.
- Integrate with Argo CD for GitOps-based deployment.
- Support environment-specific configuration such as namespace, DNS, region, and cluster.
- Standardize the deployment process across different AI Agents.

## 2. Target Architecture

The expected deployment architecture consists of three major stages:

1. Infrastructure provisioning using Terraform.
2. Helm configuration generation using the AI Agent Operator.
3. Application deployment using Argo CD.

```text
                  Application Developer
                           |
                           |
                     AIAgent CR
                           |
                           v
                Kubernetes API Server
                           |
                           v
                 AI Agent Operator
                           |
                           v
                  CR Specification
                           |
                           v
                 Helm Values Generator
                           |
              +------------+------------+
              |                         |
              v                         v
        Application Values        Network Values
              |                         |
              +------------+------------+
                           |
                           v
                     GitHub API
                           |
                    Platform Service ID
                           |
                           v
                   Git Repository
                           |
                           v
                        Argo CD
                           |
                           v
                    Helm Deployment
                           |
                           v
                      GKE Cluster
                           |
                           v
                    AI Agent Pod
```

## 3. Deployment Workflow

### Stage 1: Infrastructure Provisioning – Terraform

Terraform is responsible for provisioning the infrastructure required by the application.

Depending on the existing platform setup, this may include:

- GKE cluster and node pools.
- Kubernetes namespaces.
- Network infrastructure.
- Service accounts and IAM permissions.
- Artifact Registry access.
- Required network connectivity.
- Other platform-specific resources.

According to the project discussion, the Terraform configuration is already available.

The first activity is to understand the existing Terraform configuration and identify the resources and permissions available to the Operator.

**Expected output:** A provisioned GKE environment capable of running the AI Agent workloads.

### Stage 2: Helm Configuration Generation – Operator

This is the primary implementation responsibility of the AI Agent Operator.

The Operator receives an `AIAgent` Custom Resource and generates the corresponding Helm values using standard platform templates.

The generated configuration should cover:

**Application configuration**

- Container image.
- Image tag.
- Replica count.
- Container ports.
- Environment variables.
- CPU and memory requests and limits.
- Health and readiness probes.
- Model configuration.
- Secret references.

**Network configuration**

- Kubernetes Service.
- Service type.
- Ingress or Gateway configuration, depending on platform standards.
- DNS hostname.
- Network policies, if required.
- Application port and routing configuration.

**Environment configuration**

- GKE cluster.
- Cloud region.
- Namespace.
- Environment name.
- Deployment-specific variables.

The Operator should use reusable Helm templates maintained by the platform team rather than generating a completely different chart for every Agent.

**Expected output:** Agent-specific Helm values and the required chart configuration.

### Stage 3: GitHub Integration

After generating the Helm values, the Operator needs to publish the configuration to the designated Git repository.

The platform team is expected to provide a Service ID with the required GitHub permissions.

The Operator's GitHub integration should support:

1. Authenticate with GitHub using the approved Service ID mechanism.
2. Identify the target repository.
3. Create or update a branch.
4. Generate the required Helm values files.
5. Commit the generated configuration.
6. Create a Pull Request if required by the platform workflow.

The Git repository becomes the source of truth for the deployment configuration.

**Expected output:** Generated Helm configuration available in the designated Git repository.

### Stage 4: Argo CD Integration

Argo CD follows the GitOps deployment model.

It monitors the configured Git repository and deploys the desired application configuration into GKE.

The deployment process is:

1. Argo CD identifies changes in the Git repository.
2. Argo CD reads the Helm chart and values files.
3. Helm templates are rendered.
4. Kubernetes manifests are generated.
5. Argo CD applies the desired configuration to GKE.
6. The Agent workload is deployed.

If automated synchronization is enabled, Argo CD may deploy the changes without an explicit API call.

If the platform requires manual synchronization, the Operator may need to interact with the Argo CD API.

This behavior must be confirmed with the Cloud Engineering team.

**Expected output:** AI Agent deployed and running in the GKE cluster.

## 4. AIAgent Custom Resource Design

The `AIAgent` Custom Resource provides a declarative interface for application developers.

Developers should not need to know the internal Helm chart structure or GitHub repository layout.

### Example AIAgent YAML

The following is an illustrative API design. Final fields and naming conventions must be aligned with the platform's existing Helm templates.

```yaml
apiVersion: ai.example.com/v1alpha1
kind: AIAgent
metadata:
  name: devops-agent
  namespace: agents

spec:
  application:
    image: europe-west1-docker.pkg.dev/project/agents/devops-agent:v1
    replicas: 2
    port: 8000

  environment:
    cluster: gke-production
    region: europe-west1
    namespace: agents
    environment: production

  network:
    serviceType: ClusterIP
    dnsName: devops-agent.company.com

  model:
    provider: openai
    name: gpt-4.1-mini
    apiKeySecret: openai-api-key

  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 500m
      memory: 512Mi

  deployment:
    helm:
      repository: platform-agent-charts
      chart: standard-agent
      version: "1.0.0"
```

### CRD responsibilities

The CRD should capture the information required to generate the platform deployment configuration.

| Field | Purpose |
|---|---|
| application | Agent image and runtime configuration |
| environment | GKE cluster, region, namespace and environment |
| network | Service and DNS configuration |
| model | AI provider and model settings |
| resources | CPU and memory requirements |
| deployment | Helm chart and deployment metadata |

Sensitive information such as API keys should be referenced through Kubernetes Secrets rather than stored directly in the CR.

## 5. Helm Chart Strategy

A reusable Helm chart strategy is preferred over maintaining individual charts for every Agent.

### Proposed repository separation

The application repository should contain:

- Agent application code.
- Dockerfile.
- Application dependencies.
- AIAgent Custom Resource examples.
- Application-specific documentation.

The platform Helm repository should contain:

- Standard application Helm chart.
- Network Helm templates.
- Reusable Kubernetes templates.
- Default values.
- Environment-specific configuration conventions.

The Operator generates Agent-specific values and publishes them to the appropriate repository or configuration directory.

### Example platform repository structure

This is a proposed structure and should be adjusted to match the platform team's existing conventions.

```text
platform-agent-charts/
│
├── charts/
│   ├── standard-agent/
│   │   ├── Chart.yaml
│   │   ├── values.yaml
│   │   └── templates/
│   │       ├── deployment.yaml
│   │       ├── service.yaml
│   │       ├── configmap.yaml
│   │       └── ingress.yaml
│   │
│   └── network/
│       ├── Chart.yaml
│       ├── values.yaml
│       └── templates/
│           ├── service.yaml
│           └── networkpolicy.yaml
│
└── environments/
    ├── development/
    │   └── agents/
    ├── staging/
    │   └── agents/
    └── production/
        └── agents/
```

### Example generated values

```yaml
nameOverride: devops-agent

replicaCount: 2

image:
  repository: europe-west1-docker.pkg.dev/project/agents/devops-agent
  tag: v1
  pullPolicy: IfNotPresent

service:
  type: ClusterIP
  port: 8000

ingress:
  enabled: true
  hostname: devops-agent.company.com

resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 512Mi
```

The Operator should generate values according to the actual chart's `values.yaml` contract. The example above is not a replacement for the platform's approved chart schema.

## 6. Operator Internal Components

The Operator can be organized into separate Go packages to keep responsibilities independent.

```text
kubernetes-ai-agent-operator/
│
├── api/
│   └── v1alpha1/
│       ├── aiagent_types.go
│       └── groupversion_info.go
│
├── internal/
│   │
│   ├── controller/
│   │   └── aiagent_controller.go
│   │
│   ├── helm/
│   │   ├── generator.go
│   │   ├── values.go
│   │   └── templates/
│   │
│   ├── github/
│   │   └── client.go
│   │
│   ├── argocd/
│   │   └── client.go
│   │
│   └── config/
│       └── configuration.go
│
├── config/
│   ├── crd/
│   ├── manager/
│   └── rbac/
│
├── agent/
│   ├── app/
│   ├── requirements.txt
│   └── Dockerfile
│
├── manifests/
│   └── aiagent.yaml
│
├── Makefile
├── Dockerfile
└── README.md
```

### Component responsibilities

| Component | Responsibility |
|---|---|
| API/CRD | Defines the Agent configuration |
| Controller | Watches AIAgent resources and coordinates reconciliation |
| Helm Generator | Generates Helm values using standard templates |
| GitHub Client | Publishes configuration to GitHub |
| Argo CD Client | Interacts with Argo CD when explicit API integration is required |
| Configuration | Manages repository, chart and deployment settings |

## 7. Reconciliation Workflow

The Operator should use Kubernetes reconciliation to manage the desired deployment configuration.

A possible workflow:

```text
AIAgent Created or Updated
          |
          v
   Controller Reconcile
          |
          v
  Validate AIAgent Spec
          |
          v
  Generate Helm Values
          |
          v
  Validate Generated Values
          |
          v
  Publish to GitHub
          |
          v
  Create or Update Pull Request
          |
          v
  Wait for Git Configuration
          |
          v
  Argo CD Deployment
          |
          v
  Verify Deployment Status
          |
          v
  Update AIAgent Status
```

### Important reconciliation considerations

- Avoid generating duplicate Git commits on every reconciliation.
- Use deterministic file paths and generated values.
- Track the Git commit or configuration revision.
- Make GitHub operations idempotent.
- Handle API failures and retry safely.
- Track Pull Request status if approval is required.
- Do not report the Agent as running merely because its Helm values were generated.
- Distinguish configuration generation, Git publication, synchronization and workload readiness in the CR status.

### Suggested status phases

| Phase | Meaning |
|---|---|
| Pending | Agent resource received |
| Validating | Validating the configuration |
| Generating | Generating Helm values |
| Publishing | Publishing configuration to GitHub |
| AwaitingApproval | Pull Request requires approval |
| Deploying | Argo CD is synchronizing the configuration |
| Running | Agent workload is healthy in GKE |
| Failed | A deployment stage failed |

## 8. CI/CD Pipeline Changes

The existing CI pipeline uses the regular Kubernetes deployment methodology.

The target architecture changes the responsibility of the CI pipeline and the Operator.

### Existing workflow

```text
Source Code
    |
    v
CI Pipeline
    |
    v
Build Docker Image
    |
    v
Push Image to Registry
    |
    v
Deploy Kubernetes Resources
```

### Proposed workflow

```text
Source Code
    |
    v
CI Pipeline
    |
    v
Build and Test
    |
    v
Build Docker Image
    |
    v
Push Image to Artifact Registry
    |
    v
Update AIAgent Configuration
    |
    v
AI Agent Operator
    |
    v
Generate Helm Values
    |
    v
GitHub Repository
    |
    v
Argo CD
    |
    v
GKE Deployment
```

### Expected responsibility changes

**CI pipeline**

- Build the Agent application.
- Run tests and security checks.
- Build the container image.
- Push the image to the container registry.
- Provide the image reference to the deployment workflow.
- Update or submit the corresponding AIAgent configuration.

**AI Agent Operator**

- Read the AIAgent configuration.
- Generate Helm values.
- Publish deployment configuration to GitHub.
- Coordinate the deployment workflow.
- Track deployment progress.

**Argo CD**

- Monitor the configured Git repository.
- Render Helm templates.
- Synchronize Kubernetes resources.
- Report application health and synchronization status.

The exact mechanism for updating the AIAgent resource from CI needs to be agreed with the platform team.

## 9. GitHub API Integration

The Operator needs access to the GitHub API through the platform-approved authentication mechanism.

The platform Service ID should have only the permissions required to access the target repository and publish changes.

Potential implementation steps:

1. Configure the approved Service ID credentials.
2. Retrieve the target repository information.
3. Read the current branch or configuration.
4. Create a deployment branch.
5. Generate the required Helm values.
6. Commit the generated files.
7. Create or update a Pull Request.
8. Track the Pull Request and merge status.

### Security considerations

- Do not hardcode GitHub credentials.
- Use approved workload identity or Secret-based authentication.
- Follow the platform's credential rotation policy.
- Apply least-privilege repository permissions.
- Avoid storing credentials in generated Helm values.

## 10. Argo CD Integration

Argo CD is responsible for applying the desired application configuration from Git to GKE.

There are two possible approaches.

### Approach A: Automated synchronization

Argo CD monitors the configured Git repository and automatically synchronizes changes.

```text
Operator
   |
   v
GitHub Commit
   |
   v
Argo CD Detects Change
   |
   v
Helm Rendering
   |
   v
GKE Deployment
```

This is preferable if the platform already supports automated synchronization.

### Approach B: Explicit Argo CD API integration

If automatic synchronization is not enabled, the Operator may call the Argo CD API to request synchronization.

```text
Operator
   |
   v
GitHub Commit
   |
   v
Argo CD API
   |
   v
Application Sync
   |
   v
GKE Deployment
```

The platform team must confirm which approach is expected.

## 11. Immediate Implementation Plan

The initial milestone is to deploy a basic AI Agent into GKE using the Operator.

The work should be divided into the following tasks.

### Task 1: Understand the existing GKE environment

- Review existing Terraform configuration.
- Identify the GKE cluster and namespaces.
- Understand available service accounts and IAM permissions.
- Identify required Artifact Registry permissions.
- Understand the existing network and DNS setup.
- Identify what permissions the Operator requires.

### Task 2: Obtain standard Helm templates

- Request the existing application Helm chart.
- Request the network Helm chart or templates.
- Obtain example values files.
- Review the deployment configuration used by the disputes Agent.
- Understand environment-specific configuration requirements.
- Identify the expected Git repository and folder structure.

### Task 3: Implement Helm values generation

- Extend the AIAgent CRD.
- Add environment and network configuration fields.
- Implement the Helm values generator in Go.
- Generate values using the platform's existing chart contract.
- Validate generated values.
- Test the generated configuration using Helm.

### Task 4: Implement GitHub integration

- Obtain the approved Service ID integration.
- Configure GitHub authentication.
- Implement repository access.
- Implement branch and file operations.
- Publish generated values.
- Create Pull Requests if required.

### Task 5: Implement Argo CD deployment integration

- Identify the existing Argo CD application configuration.
- Confirm automated synchronization settings.
- Determine whether API-based synchronization is necessary.
- Integrate with the approved deployment workflow.
- Track synchronization and health status.

### Task 6: Deploy the Agent in GKE

- Create the AIAgent resource.
- Verify Helm values generation.
- Verify GitHub publication.
- Verify Argo CD synchronization.
- Verify Deployment, Pods and Services in GKE.
- Verify the Agent API endpoint.

## 12. Questions for Cloud Engineering – Datta

The following questions should be clarified before implementing the GKE integration.

### GKE and Terraform

1. What resources are already provisioned by Terraform?
2. Which namespace should the AI Agent Operator use?
3. What Kubernetes RBAC and cloud IAM permissions are required?
4. Does the Operator need any additional network access?

### Helm

5. Do we have standard application and network Helm templates?
6. What is the expected Helm values structure?
7. Which fields are mandatory for an Agent?
8. How should namespace, DNS, region and environment be represented?
9. Can we use the existing disputes Agent deployment as a reference?

### GitHub

10. Which repository should store the generated Helm values?
11. How should the platform Service ID authenticate?
12. Can the Service ID create branches and Pull Requests?
13. Are there any repository naming or folder conventions?

### Argo CD

14. Does Argo CD automatically synchronize changes from Git?
15. Does the Operator need to call the Argo CD API?
16. How should the Operator obtain deployment status?
17. What is the expected workflow for Pull Request approval and merge?

### CI/CD

18. What changes are required in the existing Agent CI pipeline?
19. Should CI update the AIAgent resource after publishing the container image?
20. Which team owns the GitOps deployment configuration?

## 13. Initial Scope vs Future Scope

### Initial scope

The immediate objective is to demonstrate that the Operator can deploy an AI Agent into GKE using the platform's existing Helm and GitOps workflow.

- Obtain standard Helm templates.
- Generate Agent-specific Helm values.
- Publish the configuration to GitHub.
- Enable Argo CD deployment.
- Deploy and validate a basic AI Agent.

### Future scope

After the basic deployment works, the implementation can be extended to support:

- Automated reconciliation of Helm configuration.
- Agent upgrades and version management.
- Multiple GKE clusters.
- Multiple regions and environments.
- Dynamic network and DNS configuration.
- Automated Pull Request lifecycle management.
- Deployment rollback.
- Agent scaling and resource management.
- Deployment health monitoring.
- Multi-Agent management.

## 14. Current Implementation vs Target Architecture

The current Minikube implementation and the target GKE architecture have different deployment responsibilities.

| Area | Current implementation | Target implementation |
|---|---|---|
| Kubernetes environment | Minikube | GKE |
| Agent configuration | AIAgent CR | AIAgent CR |
| Deployment mechanism | Operator creates Deployment and Service | Operator generates Helm configuration |
| Helm integration | Not implemented | Required |
| GitHub integration | Not implemented | Required |
| Argo CD integration | Not implemented | Required |
| Infrastructure | Local Minikube environment | Terraform-provisioned GKE |
| Deployment validation | Local Pod validation | GKE deployment validation |

The existing implementation provides the foundation for the CRD and controller concepts. It needs to evolve to support the platform's Helm and GitOps deployment model.

## 15. Definition of Done – Initial GKE Milestone

The initial milestone can be considered complete when:

- [ ] Standard platform Helm templates are obtained.
- [ ] AIAgent CRD supports the required deployment configuration.
- [ ] Operator generates valid application Helm values.
- [ ] Operator generates the required network configuration.
- [ ] Generated configuration is published to the designated Git repository.
- [ ] GitHub integration uses the approved Service ID.
- [ ] Argo CD detects and deploys the configuration.
- [ ] AI Agent Deployment is created in GKE.
- [ ] AI Agent Pod reaches Ready state.
- [ ] Agent API endpoint is accessible through the approved network path.
- [ ] AIAgent status accurately reflects deployment progress.

## 16. Final Project Goal

The objective is to build a standardized Kubernetes-native AI Agent deployment platform.

Application developers should not need to understand the platform's Helm chart structure, GitHub repository layout, or Argo CD configuration.

They should only provide an `AIAgent` Custom Resource containing their application and environment configuration.

The AI Agent Operator should generate the required deployment configuration, publish it through the approved GitOps workflow, and coordinate deployment into GKE.

**Target workflow:**

`AIAgent CR → Operator → Helm Values → GitHub → Argo CD → GKE → Running AI Agent`

This approach separates application development from infrastructure and deployment management, allowing the platform team to standardize how AI Agents are deployed across environments.
