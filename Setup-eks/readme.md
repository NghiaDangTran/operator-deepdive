Setup vpc

https://docs.aws.amazon.com/eks/latest/userguide/creating-a-vpc.html

only public 
https://docs.aws.amazon.com/eks/latest/userguide/creating-a-vpc.html#_only_public_subnets

then create with eksctl
eksctl create cluster --name operator-test --region ca-central-1 --version 1.33 --vpc-private-subnets subnet-ExampleID1,subnet-ExampleID2 --without-nodegroup






scale 

eksctl scale nodegroup \
  --cluster operator-test \
  --region ca-central-1 \
  --name operator-ng \
  --nodes 0



make docker-build docker-push deploy IMG=nghiadang23/my-nginx-operator-image:v1

make deploy IMG=nghiadang23/my-nginx-operator-image:v1

mnt/c/Users/BobDang/Documents/GitHub/The-Kubernetes-Operator-Framework-Book/chap_04

k apply -f config/samples/operator_v1alpha1_nginxoperator.yaml


eksctl scale nodegroup \
  --cluster operator-test \
  --region ca-central-1 \
  --name operator-ng \
  --nodes 0


kubectl port-forward -n kube-system svc/kite 8080:8080

cd /mnt/c/Users/BobDang/Documents/GitHub/The-Kubernetes-Operator-Framework-Book/chap_04/nginx-operator


make docker-build docker-push deploy & kubectl rollout restart deployment/nginx-operator-controller-manager -n nginx-operator-system


make generate & make manifests