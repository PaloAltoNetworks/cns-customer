# Tests for APIs

## Running the test 

### Build 

```
    go build 
```

### Run 

```
    ./tests test --api-public https://api.preprod.aporeto.us --token <token> --namespace </somewhere> --encoding json
```
    
    --api-public: API endpoint of service.
    --token:      Authorization to access the service.
    --namespace:  Base namespace where to start.

Useful options while developing tests:
    -S:           Skip teardown.


### Run a specific group of tests 

```
    ./tests test --api-public https://api.preprod.aporeto.us --token <token> --namespace </somewhere> --encoding json -t suite:zone
```