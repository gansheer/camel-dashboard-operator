# Test OLM Bundle on Local OpenShift (CRC)

## Prerequisites

- CRC running with `oc` logged in as cluster-admin
- `docker` or `podman` available
- `opm` installed (run `make opm` to fetch it)

## Registry addresses

CRC's internal registry has two addresses:

| Address | Used by | Notes |
|---|---|---|
| `image-registry.openshift-image-registry.svc:5000` | Pods inside the cluster | Stable, no TLS issues |
| `default-route-openshift-image-registry.apps-crc.testing` | Your workstation | Goes through CRC's router, self-signed cert |

`operator-sdk run bundle` tries to pull images on your workstation via the external route, which is unreliable on CRC (self-signed cert, 503s). Instead, we build a **file-based catalog (FBC)** image and create the OLM resources manually — everything stays in-cluster.

## Steps

### 1. Log in to your cluster and registry

```bash
oc login -u kubeadmin https://api.crc.testing:6443

# Expose the internal registry (if not already done)
oc patch configs.imageregistry.operator.openshift.io/cluster \
  --type merge -p '{"spec":{"defaultRoute":true}}'

# Log in to the external route
REGISTRY=$(oc get route default-route -n openshift-image-registry -o jsonpath='{.spec.host}')
docker login -u kubeadmin -p $(oc whoami -t) $REGISTRY
```

### 2. Create the target namespace

```bash
oc new-project camel-dashboard
```

### 3. Build and push the operator image

From the repo root:

```bash
INTERNAL_REGISTRY=image-registry.openshift-image-registry.svc:5000

make build-camel-monitor
make image-build \
  IMAGE_NAME=$REGISTRY/camel-dashboard/camel-monitor-operator \
  CUSTOM_VERSION=0.3.0-test
docker push $REGISTRY/camel-dashboard/camel-monitor-operator:0.3.0-test
```

### 4. Generate and push the OLM bundle image

The CSV must reference the **internal** service address so that cluster pods can pull the operator image:

```bash
make bundle \
  CUSTOM_IMAGE=$INTERNAL_REGISTRY/camel-dashboard/camel-monitor-operator \
  CUSTOM_VERSION=0.3.0-test

make bundle-push \
  BUNDLE_IMAGE_NAME=$REGISTRY/camel-dashboard/camel-monitor-operator-bundle \
  CUSTOM_IMAGE=$INTERNAL_REGISTRY/camel-dashboard/camel-monitor-operator \
  CUSTOM_VERSION=0.3.0-test
```

Verify that `bundle/manifests/camel-monitor-operator.clusterserviceversion.yaml` references the internal registry address and includes the console-related RBAC.

### 5. Build and push the FBC catalog image

A bundle image is just metadata — OLM needs a **catalog image** that runs `opm serve`. We use `opm render` to convert the bundle into a file-based catalog, then package it with the upstream `opm` image.

```bash
mkdir -p /tmp/camel-monitor-catalog && cd /tmp/camel-monitor-catalog

# Render the bundle (--skip-tls-verify for CRC's self-signed cert)
opm render \
  $REGISTRY/camel-dashboard/camel-monitor-operator-bundle:0.3.0-test \
  --skip-tls-verify > catalog.json

# Add the package and channel entries
cat >> catalog.json <<'EOF'
{
    "schema": "olm.package",
    "name": "camel-monitor-operator",
    "defaultChannel": "stable-v0"
}
{
    "schema": "olm.channel",
    "name": "stable-v0",
    "package": "camel-monitor-operator",
    "entries": [
        {
            "name": "camel-monitor-operator.v0.3.0-test"
        }
    ]
}
EOF

# Validate
opm validate .

# Build and push the catalog image
cat > Dockerfile <<'DOCKERFILE'
FROM quay.io/operator-framework/opm:v1.24.0
COPY catalog.json /configs/catalog.json
EXPOSE 50051
ENTRYPOINT ["/bin/opm"]
CMD ["serve", "/configs"]
DOCKERFILE

docker build -t $REGISTRY/camel-dashboard/camel-monitor-catalog:latest .
docker push $REGISTRY/camel-dashboard/camel-monitor-catalog:latest
```

### 6. Deploy via OLM

Create the CatalogSource, OperatorGroup, and Subscription manually. The catalog pod pulls the image via the **internal** registry address — no TLS issues. The ConsolePlugin requires **AllNamespaces** install mode (it is a cluster-scoped resource).

```bash
cat <<EOF | oc apply -n camel-dashboard -f -
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: camel-monitor-operator-catalog
spec:
  sourceType: grpc
  image: $INTERNAL_REGISTRY/camel-dashboard/camel-monitor-catalog:latest
  displayName: Camel Monitor Operator (dev)
---
apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  name: camel-dashboard-og
spec: {}
---
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: camel-monitor-operator
spec:
  channel: stable-v0
  name: camel-monitor-operator
  source: camel-monitor-operator-catalog
  sourceNamespace: camel-dashboard
  installPlanApproval: Automatic
EOF
```

### 7. Verify the deployment

```bash
# Wait for the catalog to become READY
oc get catalogsource camel-monitor-operator-catalog -n camel-dashboard \
  -o jsonpath='{.status.connectionState.lastObservedState}'

# Check the CSV status (should be Succeeded)
oc get csv -n camel-dashboard

# Check the operator pod is running
oc get pods -n camel-dashboard -l camel.apache.org/component=operator

# Check CRDs are installed
oc get crd camelmonitors.camel.apache.org

# Check the ConsolePlugin
oc get consoleplugins
```

### 8. Smoke test — create a monitored deployment

```bash
oc create deployment test-app --image=nginx -n camel-dashboard
oc label deployment test-app camel.apache.org/monitor=camel-sample-monitored -n camel-dashboard

# Verify CamelMonitor CR is created
oc get camelmonitors -n camel-dashboard
```

### 9. Cleanup

```bash
oc delete subscription camel-monitor-operator -n camel-dashboard
oc delete csv -n camel-dashboard --all
oc delete catalogsource camel-monitor-operator-catalog -n camel-dashboard
oc delete operatorgroup camel-dashboard-og -n camel-dashboard
oc delete project camel-dashboard
```

## What to watch for

- The bundle CSV should include the new console RBAC resources from `pkg/resources/config/rbac/console/`
- The `ConsolePlugin` is a cluster-scoped resource — the operator only creates it in **AllNamespaces** install mode. In `OwnNamespace` mode the operator logs: `"Operator is not global, skipping console plugin deployment (ConsolePlugin is cluster-scoped)"`
- If `opm render` returns 503, retry — CRC's registry route can be flaky under resource pressure
- The operator deployment inside the cluster pulls via the **internal** address (`svc:5000`), so image pull errors at that level are not TLS-related — check the image stream with `oc get is -n camel-dashboard`
