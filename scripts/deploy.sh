#!/bin/bash

set -e

# Configuration
DOCKER_IMAGE=${DOCKER_IMAGE:-"pinning-service:latest"}
KUBE_NAMESPACE=${KUBE_NAMESPACE:-"pinning-service"}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

usage() {
    echo "Usage: $0 {docker|k8s|all}"
    echo "  docker - Build and push Docker image"
    echo "  k8s    - Deploy to Kubernetes"
    echo "  all    - Build Docker image and deploy to Kubernetes"
    exit 1
}

build_docker() {
    echo -e "${YELLOW}Building Docker image: $DOCKER_IMAGE${NC}"
    
    docker build -t "$DOCKER_IMAGE" .
    
    echo -e "${GREEN}Docker image built successfully!${NC}"
    
    # Push to registry if DOCKER_REGISTRY is set
    if [ -n "$DOCKER_REGISTRY" ]; then
        echo -e "${YELLOW}Pushing to registry: $DOCKER_REGISTRY${NC}"
        docker tag "$DOCKER_IMAGE" "$DOCKER_REGISTRY/$DOCKER_IMAGE"
        docker push "$DOCKER_REGISTRY/$DOCKER_IMAGE"
        echo -e "${GREEN}Image pushed successfully!${NC}"
    fi
}

deploy_k8s() {
    echo -e "${YELLOW}Deploying to Kubernetes namespace: $KUBE_NAMESPACE${NC}"
    
    # Check if kubectl is available
    if ! command -v kubectl &> /dev/null; then
        echo -e "${RED}Error: kubectl not found${NC}"
        exit 1
    fi
    
    # Apply Kubernetes manifests
    kubectl apply -f deployments/k8s/namespace.yaml
    kubectl apply -f deployments/k8s/secret.yaml
    kubectl apply -f deployments/k8s/configmap.yaml
    kubectl apply -f deployments/k8s/deployment.yaml
    kubectl apply -f deployments/k8s/service.yaml
    kubectl apply -f deployments/k8s/ingress.yaml
    kubectl apply -f deployments/k8s/hpa.yaml
    
    # Wait for deployment to be ready
    echo -e "${YELLOW}Waiting for deployment to be ready...${NC}"
    kubectl wait --for=condition=available --timeout=300s deployment/pinning-service-api -n "$KUBE_NAMESPACE"
    kubectl wait --for=condition=available --timeout=300s deployment/pinning-service-worker -n "$KUBE_NAMESPACE"
    
    echo -e "${GREEN}Kubernetes deployment completed successfully!${NC}"
    
    # Show status
    echo -e "${YELLOW}Deployment status:${NC}"
    kubectl get pods -n "$KUBE_NAMESPACE"
    kubectl get services -n "$KUBE_NAMESPACE"
}

main() {
    if [ $# -eq 0 ]; then
        usage
    fi
    
    case $1 in
        "docker")
            build_docker
            ;;
        "k8s")
            deploy_k8s
            ;;
        "all")
            build_docker
            deploy_k8s
            ;;
        *)
            usage
            ;;
    esac
    
    echo -e "${GREEN}Deployment completed successfully!${NC}"
}

main "$@"
