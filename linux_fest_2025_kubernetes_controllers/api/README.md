# Kubernetes API

## Setup 

```bash
source ~/p/talks/linux_fest_2025_kubernetes_controllers/setup
```

## See the kubernetes api using curl and kubectl

### See what apis there are
```bash
curl -XGET $KUBE_API_SERVER/api/v1/namespaces/kube-systes/pods
```

### API discovery

```bash
curl -XGET $KUBE_API_SERVER/openapi/v2 | jq -c '                                              
  .host = "192.168.49.2:8443" |
  .schemes = ["https"] |
  .basePath = "/"
' > openapiv2.json

k create token swagger-user -n default

docker run --network host \
  -e SWAGGER_JSON=/swagger.json \
  -e CONFIG_URL=/swagger-config.yaml \
  -v $(pwd)/openapiv2.json:/swagger.json \
  -v $(pwd)/swagger-config.yaml:/swagger-config.yaml \
  swaggerapi/swagger-ui
```

## Extending Kubernetes API


- **Create a new CRD**

```bash
k apply -f crd.yaml
k apply -f fluffy.yaml
```

- **get, describe and delete**

```bash {name=04_gkubectl_get_fluffy}
k get pets fluffy
k describe pets fluffy
k delete pets fluffy
```

- **Run get with watch mode**

```bash
k get pets fluffy -w
k get pets fluffy -w -ojson
```
1. Show all of the above using curl

```bash 
curl $KUBE_API_SERVER/apis/animals.example.com/v1/namespaces/default/pets
```

```bash
curl --no-buffer $KUBE_API_SERVER/apis/animals.example.com/v1/namespaces/default/pets\?watch=true | jq 'del(.object.metadata.managedFields)'
```

```bash
curl -XDELETE $KUBE_API_SERVER/apis/animals.example.com/v1/namespaces/default/pets
```


## Clean up 
```bash
k delete -f crd.yaml -f fluffy.yaml
rm -rf *json
```
