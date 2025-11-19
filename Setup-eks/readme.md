Setup vpc

https://docs.aws.amazon.com/eks/latest/userguide/creating-a-vpc.html

only public 
https://docs.aws.amazon.com/eks/latest/userguide/creating-a-vpc.html#_only_public_subnets

then create with eksctl
eksctl create cluster --name operator-test --region ca-central-1 --version 1.33 --vpc-private-subnets subnet-ExampleID1,subnet-ExampleID2 --without-nodegroup




