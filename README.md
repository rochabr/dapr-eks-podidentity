# Setting Up Dapr with AWS EKS Pod Identity and Secrets Manager

This guide walks through setting up Dapr with AWS EKS Pod Identity for accessing AWS Secrets Manager.

## Prerequisites

- AWS CLI configured with appropriate permissions
- kubectl installed
- eksctl installed
- Docker installed and configured
- A Docker Hub account or another container registry

## Step 0: Clone repository

```bash
git clone github.com/rochabr/dapr-eks-podidentity.git
cd dapr-eks-podidentity
```

## Step 1: Create EKS Cluster and install Dapr

Follow the official Dapr documentation for setting up an EKS cluster and installing Dapr:
[Set up an Elastic Kubernetes Service (EKS) cluster](https://docs.dapr.io/operations/hosting/kubernetes/cluster/setup-eks/)

## Step 2: Create IAM Role and Enable Pod Identity

1. Create IAM policy for Secrets Manager access:

```bash
aws iam create-policy \
    --policy-name dapr-secrets-policy \
    --policy-document '{
        "Version": "2012-10-17",
        "Statement": [
            {
                "Effect": "Allow",
                "Action": [
                    "secretsmanager:GetSecretValue",
                    "secretsmanager:DescribeSecret"
                ],
                "Resource": "arn:aws:secretsmanager:YOUR_AWS_REGION:YOUR_ACCOUNT_ID:secret:*"
            }
        ]
    }'
```

2. Create IAM role with Pod Identity trust relationship:

```bash
aws iam create-role \
    --role-name dapr-pod-identity-role \
    --assume-role-policy-document '{
        "Version": "2012-10-17",
        "Statement": [
            {
                "Effect": "Allow",
                "Principal": {
                    "Service": "pods.eks.amazonaws.com"
                },
                "Action": [
                    "sts:AssumeRole",
                    "sts:TagSession"
                ]
            }
        ]
    }'
```

3. Attach the policy to the role:

```bash
aws iam attach-role-policy \
    --role-name dapr-pod-identity-role \
    --policy-arn arn:aws:iam::YOUR_ACCOUNT_ID:policy/dapr-secrets-policy
```

## Step 4: Create Test Resources

1. Create namespace:

```bash
kubectl create namespace dapr-test
```

2. Create service account (`service-account.yaml`):

```bash
kubectl apply -f k8s-config/service-account.yaml
```

3. Create Pod Identity association:

```bash
eksctl create podidentityassociation \
    --cluster [your-cluster-name] \
    --namespace dapr-test \
    --region [your-aws-region] \
    --service-account-name dapr-test-sa \
    --role-arn arn:aws:iam::YOUR_ACCOUNT_ID:role/dapr-pod-identity-role
```

4. Create a test secret in AWS Secrets Manager:

```bash
aws secretsmanager create-secret \
    --name test-secret \
    --secret-string '{"key":"value"}' \
    --region [your-aws-region]
```

5. Create Dapr component for AWS Secrets Manager (`aws-secretstore.yaml`):

```bash
kubectl apply -f components/aws-secretstore.yaml
```

## Step 5: Deploy Test Application

1. Build and push the Docker image:

```bash
cd app
docker build -t your-repository/dapr-secrets-test:latest .
docker push your-repository/dapr-secrets-test:latest
```

2. Apply the deployment:

```bash
kubectl apply -f deploy/app.yaml
```

## Step 6: Testing

1. Check if the pod is running:

```bash
kubectl get pods -n dapr-test
```

2. Port forward to access the application:

```bash
kubectl port-forward -n dapr-test deploy/test-app 8080:8080
```

3. Test secret access:

```bash
curl http://localhost:8080/test-secret
```

## Troubleshooting

### Authentication Issues

If you see "You must be logged in to the server (Unauthorized)", update your kubeconfig:

```bash
aws eks update-kubeconfig --region [your-aws-region] --name [your-cluster-name]
```

### Pod Identity Issues

Verify Pod Identity association:

```bash
eksctl get podidentityassociation --cluster [your-cluster-name] --region [your-aws-region]]
```

### Docker Image Issues

If you see pull access denied errors, ensure you:

1. Built the image correctly
2. Pushed it to your registry
3. Used the correct image name in the deployment YAML

### Dapr Component Issues

Check Dapr sidecar logs:

```bash
kubectl logs -n dapr-test -l app=test-app -c daprd
```

## References

- [EKS Pod Identity Documentation](https://docs.aws.amazon.com/eks/latest/userguide/pod-identities.html)
- [Dapr Secrets Management](https://docs.dapr.io/developing-applications/building-blocks/secrets/)
- [AWS Secrets Manager](https://docs.aws.amazon.com/secretsmanager/)
- [Set up an Elastic Kubernetes Service (EKS) cluster](https://docs.dapr.io/operations/hosting/kubernetes/cluster/setup-eks/)