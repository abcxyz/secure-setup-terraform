# Secure Setup Terraform

This repository contains a composite GitHub action + two linters that are built to meet the requirements set out to lightly secure the usage of Hashicorp's Terraform product from a GitHub action.

## Linters

'lint-local-exec' is a linter built to find calls to the 'local-exec' provider in a set of Terraform files

'lint-setup-terraform' is a linter built to find calls to the 'hashicorp/setup-terraform' action from a GitHub workflow

## Composite Action

The 'secure-setup-terraform' composite action does 2 primary things. 

- Downloads and runs both of the linters against the files in the repository it is run from.

- Verifies that the binary installed by 'hashicorp/setup-terraform' matches the provided checksum which can be precalculated and stored as a GitHub secret. 

## Checksum generating workflow

The workflow in '.github/workflows/generate-terraform-checksum.yml' can be used to securely verify a terraform binary and produce a checksum to be used in the above composite action.


## In a GitHub workflow

```yaml
env:
  terraform_version: '1.3.2'
  terraform_checksum: '${{ secrets.TERRAFORM_CHECKSUM }}'

jobs:
  secure-setup-terraform:
    runs-on: 'ubuntu-latest'
    steps:
    - name: 'checkout'
      uses: 'actions/checkout@v3'
    -
      name: 'secure-setup-terraform'
      uses: 'bradegler/secure-setup-terraform@v0.0.10'
      with:
        terraform_version: ${{env.terraform_version}}
        terraform_checksum: ${{env.terraform_checksum}}
    ## Use terraform normally
```

## Building the linters
```sh
# Linter to find calls to the 'local-exec' terraform provider
go build -o lint-local-exec ./cmd/lint-local-exec

# Linter to find calls to the 'setup-terraform' GitHub
# action from Hashicopr
go build -o lint-setup-terraform ./cmd/lint-setup-terraform
```
