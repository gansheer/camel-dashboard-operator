FROM scratch

# Tells the pipeline that this is a bundle image and should be delivered via an index image
LABEL com.redhat.delivery.operator.bundle=true

# Tells the pipeline which versions of OpenShift the operator supports.
# This is used to control which index images should include this operator.
LABEL com.redhat.openshift.versions="v4.12-v4.17"

# The rest of these labels are copies of the same content in annotations.yaml and are needed by OLM
# Note the package name and channels which are very important!

# Core bundle labels.
LABEL operators.operatorframework.io.bundle.mediatype.v1=registry+v1
LABEL operators.operatorframework.io.bundle.manifests.v1=manifests/
LABEL operators.operatorframework.io.bundle.metadata.v1=metadata/
LABEL operators.operatorframework.io.bundle.package.v1=red-hat-camel-dashboard
LABEL operators.operatorframework.io.bundle.channels.v1=1.10.10.x,candidate,latest
LABEL operators.operatorframework.io.bundle.channel.default.v1=1.10.10.x
LABEL operators.operatorframework.io.metrics.builder=operator-sdk-v1.13.0+git
LABEL operators.operatorframework.io.metrics.mediatype.v1=metrics+v1
LABEL operators.operatorframework.io.metrics.project_layout=go.kubebuilder.io/v2

# Labels for testing.
LABEL operators.operatorframework.io.test.mediatype.v1=scorecard+v1
LABEL operators.operatorframework.io.test.config.v1=tests/scorecard/

# Copy files to locations specified by labels.
COPY manifests /manifests/
COPY metadata /metadata/
COPY tests/scorecard /tests/scorecard/

# Log the camel-dashboard operator version we built this with
LABEL com.redhat.fuse.camel-dashboard.operatorversion=camelk-1.10.10-openshift-rhel-8
LABEL com.redhat.fuse.camel-dashboard.csvversion=1.10.10

# This last block are standard Red Hat container labels
LABEL \
        com.redhat.component="red-hat-camel-dashboard-bundle-container" \
        version="1.10.10" \
        name="integration/camel-dashboard-rhel8-operator-bundle" \
        License="ASL 2.0" \
        io.k8s.display-name="red-hat-camel-dashboard bundle" \
        io.k8s.description="bundle containing manifests for red-hat-camel-dashboard" \
        summary="bundle containing manifests for red-hat-camel-dashboard" \
        maintainer="Thomas Cunningham <tcunning@redhat.com>"
