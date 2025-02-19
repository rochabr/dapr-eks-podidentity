# Setting Up Dapr with AWS EKS Pod Identity, AWS Secrets Manager, and Amazon S3

This guide walks through setting up Dapr with AWS EKS Pod Identity for accessing AWS Secrets Manager and Amazon S3.

## Prerequisites

- AWS CLI configured with appropriate permissions
- kubectl installed
- eksctl installed
- Docker installed and configured
- A Docker Hub account or another container registry

## Clone repository

```bash
git clone https://github.com/rochabr/dapr-eks-podidentity.git
cd dapr-eks-podidentity
```

## Create EKS Cluster and install Dapr

Follow the official Dapr documentation for setting up an EKS cluster and installing Dapr:
[Set up an Elastic Kubernetes Service (EKS) cluster](https://docs.dapr.io/operations/hosting/kubernetes/cluster/setup-eks/)

## Create IAM Role and Enable Pod Identity

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

2. Create IAM policy for S3 full access:

```bash
aws iam create-policy \
    --policy-name dapr-s3-policy \
    --policy-document '{
        "Version": "2012-10-17",
        "Statement": [
            {
                "Effect": "Allow",
                "Action": [
                    "s3:*",
                    "s3-object-lambda:*"
                ],
                "Resource": "*"
            }
        ]
    }'
```

3. Create IAM role with Pod Identity trust relationship:

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

4. Attach the policies to the role:

```bash
aws iam attach-role-policy \
    --role-name dapr-pod-identity-role \
    --policy-arn arn:aws:iam::YOUR_ACCOUNT_ID:policy/dapr-secrets-policy
```

```bash
aws iam attach-role-policy \
    --role-name dapr-pod-identity-role \
    --policy-arn arn:aws:iam::YOUR_ACCOUNT_ID:policy/dapr-s3-policy
```

## Create Test Resources

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

5. Create S3 bucket:

```bash
aws s3api create-bucket --bucket [your-bucket-name] --region [your-aws-region]
```

6. Create Dapr component for AWS Secrets Manager (`aws-secretstore.yaml`) and AWS S3 (`aws-s3.yaml`):

> Update the necessary values on both component files before running the command below.

```bash
kubectl apply -f components/
```

## Deploy Test Application

1. Build and push the Docker image:

```bash
cd app
docker build -t your-repository/dapr-secrets-test:latest .
docker push your-repository/dapr-secrets-test:latest
```

2. Apply the deployment:

> Update the `image` attribute by adding your repository name.

```bash
kubectl apply -f deploy/app.yaml
```

## Testing

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

4. Test S3 access:

```bash
curl -X POST -H "Content-Type: application/json" \
     -d '{"data": "Hello World"}' \
     http://localhost:8080/create-s3
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
eksctl get podidentityassociation --cluster [your-cluster-name] --region [your-aws-region]
```

### Dapr Component Issues

Check Dapr sidecar logs:

```bash
kubectl logs -n dapr-test -l app=test-app -c daprd
```

## References

- [EKS Pod Identity Documentation](https://docs.aws.amazon.com/eks/latest/userguide/pod-identities.html)
- [AWS Secrets Manager](https://docs.aws.amazon.com/secretsmanager/)
- [Set up an Elastic Kubernetes Service (EKS) cluster](https://docs.dapr.io/operations/hosting/kubernetes/cluster/setup-eks/)
