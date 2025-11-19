Create the project
operator-sdk init --domain example.com --repo github.com/example/nginx-operator

config – A directory that holds YAML definitions of Operator resources.
• hack – A directory that is used by many projects to hold various hack scripts.
These are scripts that can serve a variety of purposes but are often used to generate
or verify changes (often employed as part of a continuous integration process to
ensure code is properly generated before merging).
72 Developing an Operator with the Operator SDK
• .dockerignore / .gitignore – Declarative lists of files to be ignored by
Docker builds and Git, respectively.
• Dockerfile – Container image build definitions.
• Makefile – Operator build definitions.
• PROJECT – File used by Kubebuilder to hold project config information
(https://book.kubebuilder.io/reference/project-config.html).
• go.mod / go.sum – Dependency management lists for go mod (already
populated with various Kubernetes dependencies).
• main.go – The entry point file for the Operator's main functional code.


Create api

operator-sdk create api --group operator --version v1alpha1 --kind NginxOperator --resource --controller