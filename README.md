# Kubernetes AI Agent Operator

## 1. Project Overview

The Kubernetes AI Agent Operator is a Kubernetes-native application that automates the deployment and management of AI Agents inside a Kubernetes cluster.

Instead of manually creating Deployments, Pods, and Services for every AI Agent, users define an `AIAgent` custom resource. The Operator reads this configuration and creates the required Kubernetes resources automatically.

The project consists of two main components:

- **AI Agent:** A Python-based application that communicates with an AI model and provides APIs for interacting with the agent.
- **AI Agent Operator:** A Go-based Kubernetes Operator that manages the lifecycle of AI Agent workloads.

The main objective is to make AI Agent deployment and management similar to deploying any other Kubernetes application.

## 2. Architecture

```text
                   User
                    |
                    |
             kubectl apply
                    |
                    v
             AIAgent YAML
                    |
                    v
          Kubernetes API Server
                    |
                    v
          AI Agent Operator
             (Go Controller)
                    |
          Reads AIAgent Spec
                    |
          Creates Kubernetes Resources
                    |
          +---------+---------+
          |         |         |
          v         v         v
      Deployment  Service   Configuration
          |
          v
         Pod
          |
          v
     Python AI Agent
          |
          v
       OpenAI API
```

## 3. How the Operator Deploys the AI Agent

The Operator acts as the manager of the AI Agent. It automates the deployment and maintenance of the Agent based on the configuration provided by the user.

### Step 1: User creates an AIAgent resource

The user defines the desired AI Agent configuration in a YAML file.

For example:

```yaml
apiVersion: ai.example.com/v1alpha1
kind: AIAgent
metadata:
  name: devops-agent
spec:
  replicas: 1
  image: ai-agent:v1
  model:
    provider: openai
    name: gpt-4.1-mini
  agent:
    role: devops-engineer
```

The user submits this resource to Kubernetes:

```bash
kubectl apply -f manifests/aiagent.yaml
```

### Step 2: Operator detects the resource

The AI Agent Operator continuously watches for `AIAgent` custom resources.

When a new `AIAgent` resource is created or updated, the Operator Controller receives an event and starts its reconciliation process.

### Step 3: Operator creates Kubernetes resources

The Controller reads the desired configuration from the `AIAgent` resource and creates the required Kubernetes objects.

| Kubernetes resource | Purpose |
|---|---|
| AIAgent | Contains the desired AI Agent configuration |
| Deployment | Defines how the AI Agent application should run |
| Pod | Runs the Python AI Agent application |
| Service | Provides network access to the AI Agent |

The Operator creates these resources through the Kubernetes API. It does not directly start containers or create Pods itself.

### Step 4: AI Agent starts running

The Kubernetes Deployment Controller takes responsibility for the Deployment.

It creates the required Pod, and Kubernetes schedules the Pod onto an available node.

The Python FastAPI application starts inside the Pod and exposes its API, including:

- `/` - Application information
- `/health` - Health check
- `/ready` - Readiness check
- `/chat` - AI Agent interaction endpoint

The Service provides a stable network endpoint to access the Agent.

### Step 5: Operator continuously maintains the desired state

The Operator continues watching the resources it manages.

It compares the desired state defined in the `AIAgent` resource with the actual Kubernetes resources.

If the configuration changes, the Operator reconciles the resources to match the new configuration.

## 4. What Happens If the AI Agent Pod Crashes?

The Operator and Kubernetes Deployment Controller have different responsibilities.

Consider a situation where the AI Agent Pod crashes:

1. Kubernetes detects that the Pod has stopped.
2. The Kubernetes Deployment Controller ensures that the required number of Pods is maintained.
3. Kubernetes creates a replacement Pod.
4. The replacement Pod starts the AI Agent application.
5. The Operator observes the updated resource state and updates the `AIAgent` status when appropriate.

This means that a single Pod failure does not require the user to manually deploy the Agent again.

## 5. Operator Controller vs Kubernetes Deployment Controller

There are two different controllers involved in this architecture.

| Controller | Responsibility |
|---|---|
| AI Agent Operator Controller | Manages the desired configuration of the AI Agent and its Kubernetes resources |
| Kubernetes Deployment Controller | Maintains the required number of Pods for the Deployment |

**Example:**

If the user changes the replica count from 1 to 3:

- The AI Agent Operator detects the updated `AIAgent` resource.
- It updates the Deployment replica count to 3.
- The Kubernetes Deployment Controller creates additional Pods.
- Kubernetes schedules the Pods onto available nodes.

The Operator manages the Agent configuration; Kubernetes manages the actual Pod lifecycle.

## 6. What Is Reconciliation?

Reconciliation is the core mechanism of a Kubernetes Operator.

The Controller continuously works to make the actual cluster state match the desired state.

For example:

| Desired state | Actual state | Action |
|---|---|---|
| 1 Agent Pod | 1 Pod running | No action required |
| 2 Agent Pods | 1 Pod running | Update Deployment replicas |
| Agent image v2 | Agent image v1 | Update Deployment image |
| Agent Service exists | Service missing | Recreate Service |

This makes the Operator declarative and self-correcting.

## 7. Technology Stack

| Component | Technology |
|---|---|
| Operator language | Go |
| Operator framework | Kubebuilder |
| Kubernetes API | controller-runtime |
| AI Agent backend | Python |
| API framework | FastAPI |
| AI model integration | OpenAI API |
| Containerization | Docker |
| Kubernetes environment | Minikube |
| Infrastructure | AWS EC2 |
| Container runtime in Minikube | containerd |

## 8. Custom Resource Definition (CRD)

The project introduces a custom Kubernetes resource named `AIAgent`.

**API details:**

- API Group: `ai.example.com`
- Version: `v1alpha1`
- Kind: `AIAgent`

The CRD allows Kubernetes to understand and store AI Agent configurations.

The user can define:

- Number of replicas
- Agent container image
- AI model provider
- AI model name
- Agent role
- System prompt
- Agent capabilities
- Memory configuration
- CPU and memory resources

## 9. AI Agent Application

The AI Agent is a Python application built using FastAPI.

Its responsibilities include:

- Accepting user requests through HTTP APIs.
- Sending prompts to the configured AI model.
- Receiving and returning AI-generated responses.
- Providing health and readiness endpoints.
- Running as a container inside Kubernetes.

The Operator manages the Agent's Kubernetes deployment, while the Python application handles the actual AI interactions.

## 10. Project Structure

```text
kubernetes-ai-agent-operator/
│
├── api/
│   └── v1alpha1/
│       ├── aiagent_types.go
│       └── groupversion_info.go
│
├── internal/
│   └── controller/
│       └── aiagent_controller.go
│
├── cmd/
│   └── main.go
│
├── config/
│   ├── crd/
│   ├── default/
│   ├── manager/
│   └── rbac/
│
├── agent/
│   ├── app/
│   │   ├── __init__.py
│   │   └── main.py
│   ├── requirements.txt
│   └── Dockerfile
│
├── manifests/
│   └── aiagent.yaml
│
├── Dockerfile
├── Makefile
├── go.mod
└── README.md
```

## 11. Environment Setup

The development environment uses:

- Ubuntu 24.04 on AWS EC2
- Go
- Kubebuilder
- Docker
- Minikube
- kubectl
- Python 3.12

The Minikube cluster uses the Docker driver, with containerd as its internal container runtime.

## 12. Build and Deployment

### Build the Operator

```bash
make manifests
make generate
make build
```

### Build the AI Agent image

```bash
docker build -t ai-agent:v1 ./agent
```

### Build the Operator image

```bash
make docker-build IMG=ai-agent-operator:v1
```

### Load images into Minikube

```bash
minikube image load ai-agent:v1
minikube image load ai-agent-operator:v1
```

### Install the CRD

```bash
make install
```

### Deploy the Operator

```bash
make deploy IMG=ai-agent-operator:v1
```

### Create the AI Agent

```bash
kubectl apply -f manifests/aiagent.yaml
```

## 13. Verify Deployment

Check the Operator:

```bash
kubectl get pods -n kubernetes-ai-agent-operator-system
```

Check the custom resource:

```bash
kubectl get aiagents
```

Check the Kubernetes resources created by the Operator:

```bash
kubectl get deployment,pod,service
```

Check the Agent status:

```bash
kubectl describe aiagent devops-agent
```

Check the Operator logs:

```bash
kubectl logs -n kubernetes-ai-agent-operator-system \
  deployment/kubernetes-ai-agent-operator-controller-manager \
  -c manager
```

## 14. Security Considerations

The project should follow Kubernetes security best practices:

- Store AI API keys in Kubernetes Secrets.
- Avoid hardcoding API keys in application code or YAML files.
- Use dedicated ServiceAccounts and minimum required RBAC permissions.
- Run containers as non-root users.
- Configure CPU and memory requests and limits.
- Use health and readiness probes.
- Restrict access to Agent APIs using authentication and authorization.

## 15. Future Enhancements

The following features can be added as the project evolves:

- Support for multiple AI model providers.
- Secure API key management through Kubernetes Secrets.
- Agent memory and conversation history.
- Persistent storage for Agent memory.
- Dynamic Agent configuration updates.
- Horizontal scaling of AI Agent workloads.
- Integration with Kubernetes APIs for cluster discovery.
- Controlled access to Kubernetes resources.
- Tool integration for DevOps automation.
- Multiple specialized AI Agents.
- Agent-to-Agent communication.
- Observability, metrics, and logging.

## 16. Project Goal

The goal is to build a Kubernetes-native platform where users can deploy and manage AI Agents using simple YAML configurations.

Instead of manually managing containers, Deployments, Services, and configuration, users define an `AIAgent` resource and let the Operator manage the Kubernetes infrastructure required to run it.

**In simple terms:**

You declare what kind of AI Agent you want. The Operator translates that declaration into Kubernetes resources, deploys the Agent, and manages its lifecycle.
