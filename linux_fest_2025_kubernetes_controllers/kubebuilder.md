# Steps
```bash {name=kubebuilder_mkdir}
mkdir pet-controller
```

```bash {name=kubebuilder_init}
kubebuilder init --domain=example.com --repo=github.com/itzloop/pet-controller
```

```bash {name=kubebuilder_create_api}
kubebuilder create api --group=animals --version=v1 --kind=Pet
```

```bash {name=kubebuilder_make_install}
make install
```

```bash {name=kubebuilder_make_run}
make run
```
