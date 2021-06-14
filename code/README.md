# Folder Organization 

- `ac`

  Code for a binary and library.
  
- `config` 

  Config for a few tenants.
  
# Demo 

### Building binary 

```cd ac && go install```

### Executing test cases

```ac -config <path-to-config.json> -scenario <scenario>```
  
### Description of Scenarios 

- zone-create
- zone-delete
- tenant-create
- tenant-disable
- tenant-delete
- service-create
- service-delete
- exception-create
- exception-delete

# Library Usage

### Golang

```import "github.com/PaloAltoNetworks/cns-customer/ac/api"```
