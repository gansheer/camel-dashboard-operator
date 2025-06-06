# Camel Dashboard Operator

The Camel Dashboard Operator is a project created to simplify the management of any Camel application on a Kubernetes cluster. The tool is in charge to monitor any Camel application and provide a set of basic information, useful to learn how your fleet of Camel (a caravan!?) is behaving.

The project is designed to be as simple and low resource consumption as possible. It only collects the most important Camel application KPI in order to quickly identify what's going on across your Camel applications.

NOTE: as the project is still in an experimental phase, the metrics collected can be changed at each development iteration.

## The Camel custom resource

The operator uses a simple custom resource known as `app` or `capp` (as Camel App) which stores certain metrics around your running applications. The operator detects the Camel applications you're deploying to the cluster, identifying them in a given namespace or a given metadata label that need to be included when deploying your applications (all configurable on the operator side).

## Install the operator

You can use Helm to install the operator resources. You can install it in any namespace (we conventially use `camel-dashboard` namespace, which, has to be created previously). The default configuration is for a cluster scoped operator (use `--set operator.global=\"false\"` for a namespace scoped operator).

```
helm install camel-dashboard https://github.com/camel-tooling/camel-dashboard-operator/raw/refs/heads/main/docs/charts/camel-dashboard-0.0.1-SNAPSHOT.tgz -n camel-dashboard
```

NOTE: the installation procedure is still in experimental phase and uses a snapshot artifacts. It could change in future stable releases.

You can check if the operator is running:

```
kubectl get pods -n camel-dashboard
NAME                                        READY   STATUS    RESTARTS   AGE
camel-dashboard-operator-7c6bcf5576-fwn7s   1/1     Running   0          4m18s
```

## Import a Camel application

The operator is instructed to watch `Deployment` and verify if they are marked as Camel application. You will likely need to update your deployment process and include automatically a `camel.apache.org/app` label for all the applications you want to monitor.

NOTE: you can configure the operator to watch for a different label setting the environment variable `LABEL_SELECTOR` in the operator Pod.

## Collect Camel metrics

The operator is designed to consume the services exposed by [Camel Observability Services component](https://camel.apache.org/components/next/others/observability-services.html).

It will works also when no services are exposed, but it won't be able to collect any meaningful metrics (likely only the status and the number of replicas).

## Run a new Camel application

Let's run some sample Camel application. We have prepared a few available to run some quick demo:

* A Camel main application available at `docker.io/squakez/db-app-main:1.0`
* A Camel Quarkus application available at `docker.io/squakez/db-app-quarkus:1.0`
* A Camel Spring Boot application available at `docker.io/squakez/db-app-sb:1.0`

These applications were created, exported and "containerized" via Camel JBang, which includes by default the aforementioned `camel-observability-services` dependency.

Let's run them in a Kubernetes cluster (it also works in a local cluster such as `Minikube`):

```
kubectl create deployment camel-app-main --image=docker.io/squakez/db-app-main:1.0
```

The application should start, but, since there is no label for the operator, this one cannot discover it.

NOTE: ideally your pipeline process should be the one in charge to include this and any other label to the applications.

Let's include the label via CLI:

```
kubectl label deployment camel-app-main camel.apache.org/app=camel-app-main
```

NOTE: you can test it straight away with any of your existing Camel application by adding the label as well.

The application is immediately imported by the operator. Its metrics are also scraped and available to be monitored:

```
kubectl get apps
NAME             PHASE     IMAGE                               REPLICAS   INFO
camel-app-main   Running   docker.io/squakez/db-app-main:1.0   1          Main - 4.11.0 (4.11.0) [exchanges: total 973, succeeded 876, failed 97, pending 1]
```

NOTE: more information are available inspecting the custom resource (i.e. via `-o yaml`).

## Fine tune the metrics polling

You can watch the metrics evolving as long as the application is running, for example via `-w` parameter:

```
kubectl get apps -w
NAME             PHASE     IMAGE                               REPLICAS   INFO
camel-app-main   Running   docker.io/squakez/db-app-main:1.0   1          Main - 4.11.0 (4.11.0) [exchanges: total 1197, succeeded 1078, failed 119, pending 1]
camel-app-sb     Running   docker.io/squakez/db-app-sb:1.0     1          Spring-Boot - 3.4.3 (4.11.0) [exchanges: total 20, succeeded 18, failed 2, pending 1]
camel-app-sb     Running   docker.io/squakez/db-app-sb:1.0     1          Spring-Boot - 3.4.3 (4.11.0) [exchanges: total 32, succeeded 28, failed 4, pending 1]
camel-app-main   Running   docker.io/squakez/db-app-main:1.0   1          Main - 4.11.0 (4.11.0) [exchanges: total 1207, succeeded 1086, failed 121, pending 1]
camel-app-sb     Running   docker.io/squakez/db-app-sb:1.0     1          Spring-Boot - 3.4.3 (4.11.0) [exchanges: total 44, succeeded 36, failed 8, pending 1]
camel-app-main   Running   docker.io/squakez/db-app-main:1.0   1          Main - 4.11.0 (4.11.0) [exchanges: total 1218, succeeded 1097, failed 121, pending 1]
```

The `App` are polled every minute by default. It should be enough in most cases, as the project is really a dashboard and not a proper monitoring tool. However, you can change this configuration if you want a more or less reactive polling. You can configure this value both at operator level (which would affect all the applications) or at single application level.

### Operator level

You can setup the environment variables `POLL_INTERVAL_SECONDS` with the number of seconds between each metrics polling.

NOTE: this will affect all your applications. Setting it a low value can reduce the performances of both the operator and the same Camel applications which will need to use compute resources to read from the HTTP service.

### Application level

You can add an annotation to the `App` resource, `camel.apache.org/polling-interval-seconds` with the value you want.

NOTE: although this configuration will only affect the single application, consider the right balance to avoid affecting the application performances.

## Openshift plugin

This operator can work standalone and you can use the data exposed in the `App` custom resource accordingly. However it has a great fit with the [Camel Openshift Console Plugin](https://github.com/camel-tooling/camel-openshift-console-plugin?tab=readme-ov-file#deployment-to-openshift), which is a visual representation of the services exposed by the operator.
