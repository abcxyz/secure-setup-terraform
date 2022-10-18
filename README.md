# Secure Setup Terraform


## Usage


```yaml
```

### Vendor and sign project source

```sh

```

### In a CI/CD workflow

```yaml
jobs:
  my_job:
    runs-on: 'ubuntu-latest'
    name: 'osprey'
    steps:
      - uses: 'actions/checkout@v3'

      # Example of building from secured source
      - uses: 'docker://ghcr.io/abcxyz/secure-setup-terraform:0.1.4'
        with:
          args: "build examples/terraform/osprey_v1.2.2.yaml"
```

## Installation


### Building 

```sh
```
