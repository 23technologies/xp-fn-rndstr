# function-dummy

A [Crossplane] Composition Function that returns what you tell it to.

## What is this?

This [Composition Function][function-design] generates random strings and inserts them into specified paths of a composition's resources. This is useful, when one needs to generate e.g. credentials for foreign infrastructure.

Here's an example:

```yaml
apiVersion: apiextensions.crossplane.io/v1
kind: Composition
metadata:
  name: test-xp-rndstr
spec:
  compositeTypeRef:
    apiVersion: database.example.com/v1alpha1
    kind: NoSQL
  pipeline:
    - step: patch-and-transform
      functionRef:
        name: function-patch-and-transform
      input:
        apiVersion: pt.fn.crossplane.io/v1beta1
        kind: Resources
        resources:
          - name: some-secret-for-password 
            base:
              apiVersion: kubernetes.crossplane.io/v1alpha1
              kind: Object
              spec:
                providerConfigRef:
                  name: provider-config-in-cluster
                forProvider:
                  manifest:
                    apiVersion: v1
                    kind: Secret
                    metadata:
                      namespace: default
                    stringData:
                      password: patchme
    - step: generate-openstack-password
      functionRef:
        name: xp-rndstr-func 
      input:
        apiVersion: randomstring.fn.23technologies.cloud
        kind: randString
        config:
          objects:
            - name: some-secret-for-password 
              fieldPath: "spec.forProvider.manifest.stringData.password"
          randomString:
            length: 10
```

## Developing

This Function doesn't use the typical Crossplane build submodule and Makefile,
since we'd like Functions to have a less heavyweight developer experience.
It mostly relies on regular old Go tools:

```shell
# Run code generation - see input/generate.go
$ go generate ./...

# Run tests
$ go test -cover ./...
?       github.com/crossplane-contrib/function-dummy/input/v1beta1      [no test files]
ok      github.com/crossplane-contrib/function-dummy    0.006s  coverage: 25.8% of statements

# Lint the code
$ docker run --rm -v $(pwd):/app -v ~/.cache/golangci-lint/v1.54.2:/root/.cache -w /app golangci/golangci-lint:v1.54.2 golangci-lint run

# Build a Docker image - see Dockerfile
$ docker build .
```

This Function is pushed to `xpkg.upbound.io/crossplane-contrib/function-dummy`.
At the time of writing it's pushed manually via `docker push` using
`docker-credential-up` from https://github.com/upbound/up/.

[Crossplane]: https://crossplane.io
[function-design]: https://github.com/crossplane/crossplane/blob/3996f20/design/design-doc-composition-functions.md
[function-pr]: https://github.com/crossplane/crossplane/pull/4500
[docs-composition]: https://docs.crossplane.io/v1.13/getting-started/provider-aws-part-2/#create-a-deployment-template
[#2581]: https://github.com/crossplane/crossplane/issues/2581
[merge]: https://pkg.go.dev/github.com/golang/protobuf/proto#Merge
