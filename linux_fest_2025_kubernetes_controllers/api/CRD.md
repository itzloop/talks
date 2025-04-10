# Steps

1. Setup

```bash {name=00_setup}
source setup
```

1. Add crd

```bash {name=01_kubectl_create_crd}
k apply -f crd.yaml
```

1. Add fluffy

```bash {name=02_kubectl_apply_fluffy}
k apply -f fluffy.yaml
```

1. get, describe and delete

```bash {name=04_gkubectl_get_fluffy}
k get pets fluffy
```

```bash {name=05_kubectl_describe_fluffy}
k describe pets fluffy
```

```bash {name=06_kubectl_delete_fluffy}
k delete pets fluffy
```

1. Run get with watch mode

```bash  {name=07_kubectl_watch_fluffy}
k get pets fluffy -w
```

```bash {name=08_kubectl_watch_fluffy_json}
k get pets fluffy -w -ojson
```

1. Show all of the above using curl

```bash {name=09_curl_get_fluffy}
curl -ks --no-buffer $KUBE_API_SERVER/apis/animals.example.com/v1/namespaces/default/pets'
```

```bash {name=10_curl_watch_fluffy}
curl -ks --no-buffer $KUBE_API_SERVER/apis/animals.example.com/v1/namespaces/default/pets\?watch=true | jq 'del(.object.metadata.managedFields)'
```

```bash {name=11_curl_delete_fluffy}
curl -XDELETE $KUBE_API_SERVER/apis/animals.example.com/v1/namespaces/default/pets
```
