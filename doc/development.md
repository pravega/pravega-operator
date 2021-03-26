## Development

### Contents

 * [Requirements](#requirements)
 * [Build the operator image](#build-the-operator-image)
 * [Run the Operator locally](#run-the-operator-locally)
 * [Installation on Google Kubernetes Engine](#installation-on-google-kubernetes-engine)
 * [Steps for creating a pull request on Pravega Operator](#steps-for-creating-a-pull-request-on-pravega-operator)
 * [Deploying in Test Mode](#deploying-in-test-mode)

#### Requirements:
  - Go 1.13+

##### Install Go

You can install go directly or use gvm ( go version manager)

Install gvm:

```
bash < <(curl -s -S -L https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer)
```

See all currently installed go versions:
```
gvm list
```

See all available go versions that can be installed using gvm:
```
gvm listall
```

Install a new go version:
```
gvm install go1.4 -B
gvm use go1.4
gvm install go1.13.8 --binary
gvm use go1.13.8 --default
```
Your GOPATH should be be set by now, check using
```
echo $GOPATH
```
should display something like `/home/<userdir>/.gvm/pkgsets/go1.11/global`

Clone operator repo:
```
cd $GOPATH
go get github.com/pravega/pravega-operator
```
This should clone operator code under `$GOPATH/src/github.com/pravega/pravega-operator`

For pulling the dependencies we are using go modules for more details on go modules refer to the link below:-

https://blog.golang.org/using-go-modules
#### Build the operator image

Use the `make` command to build the Pravega operator image, it will also automatically get all the dependencies by using the go.mod file.

```
$ cd $GOPATH/src/github.com/pravega/pravega-operator
$ make build
```
That will generate a Docker image with the format
`<latest_release_tag>-<number_of_commits_after_the_release>` (it will append-dirty if there are uncommitted changes). The image will also be tagged as `latest`.

Example image after running `make build`.

The Pravega Operator image will be available in your Docker environment.

```
$ docker images pravega/pravega-operator

REPOSITORY                  TAG            IMAGE ID      CREATED          SIZE        

pravega/pravega-operator    0.1.1-3-dirty  2b2d5bcbedf5  10 minutes ago   41.7MB    

pravega/pravega-operator    latest         2b2d5bcbedf5  10 minutes ago   41.7MB

```

Optionally push it to a Docker registry.

```
docker tag pravega/pravega-operator [REGISTRY_HOST]:[REGISTRY_PORT]/pravega/pravega-operator
docker push [REGISTRY_HOST]:[REGISTRY_PORT]/pravega/pravega-operator
```

where:

- `[REGISTRY_HOST]` is your registry host or IP (e.g. `registry.example.com`)
- `[REGISTRY_PORT]` is your registry port (e.g. `5000`)

#### Run the Operator locally

You can run the Operator locally to help with development, testing, and debugging tasks.

The following command will run the Operator locally with the default Kubernetes config file present at `$HOME/.kube/config`. Use the `--kubeconfig` flag to provide a different path.

```
$ make run-local
```

#### Installation on Google Kubernetes Engine

The Operator requires elevated privileges in order to watch for the custom resources.

According to Google Container Engine docs:

> Ensure the creation of RoleBinding as it grants all the permissions included in the role that we want to create. Because of the way Container Engine checks permissions when we create a Role or ClusterRole.
>
> An example workaround is to create a RoleBinding that gives your Google identity a cluster-admin role before attempting to create additional Role or ClusterRole permissions.
>
> This is a known issue in the Beta release of Role-Based Access Control in Kubernetes and Container Engine version 1.6.

On GKE, the following command must be run before installing the Operator, replacing the user with your own details.

```
$ kubectl create clusterrolebinding your-user-cluster-admin-binding --clusterrole=cluster-admin --user=your.google.cloud.email@example.org
```

#### Steps for creating a pull request on Pravega Operator

```
$> git clone <operator repo url>
$> git checkout -b <issue-234-unique-branch-name>
```

Commit changes to this branch using:
```
$> git commit --signoff -m "commit message"
```

A PR can be created after the very first commit on the branch from github UI.
Once the PR is created, all subsequent commits to the branch will be added appended to the same PR.

To refresh from master:
```
$> git pull origin master
```

To push changes to branch:
```
$> git push origin issue-234-unique-branch-name
```

#### Deploying in Test Mode
The Operator can be run in `test mode` if we want to deploy pravega on minikube or on a cluster with very limited resources by enabling `testmode: true` in `values.yaml` file of operator charts. Operator running in test mode skips minimum replica requirement checks on Pravega components. Test mode provides a bare minimum setup and is not recommended to be used in production environments.
