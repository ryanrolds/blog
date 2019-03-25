# Onboarding developers quickly

For any new project, especially ones with a fixed timeframe, onboarding developers can require a significant amount time. This article is going to go over a minimalist base that will provide a unified developer enviornment for all major operating systems. We will go over each tool and why it's required, how we compose the tools to setup developer environments quickly, and finally how to use deploy the environment to AWS.

## Setup 

Docker and Docker Compose will make it easy for developers to setup their development environment, including databases and other dependencies. They also allow easy publishing of images to the cloud for deployment.

### Install tools

> Every developer must install these tools

Git, Docker, and Docker Compose.

### Create Docker related files

> This only needs to be done once for the project

`Dockerfile`

`docker-compose.yaml`

## Workflow

Once the the developer has installed the tools they will need to fork the repo in GitHub and clone their fork.

    $ git clone <fork repo url>
    $ cd <repo dir>

They should first look at the project's README and perform any setup steps. Each project usuaully requires install a platform (like Go, Ruby, PHP, Python, Node.js, etc...). If you don't plan on running your application on your host OS, you don't need to install anything as Docker containers will contain everything need to build and run the application.

Setting up the project's environment will generally reflect steps on this post. Most projects will require starting the database, running a command to setup the database, followed by instructions on how to start the application. Those steps will rougly look like:

    $ docker-compose up -d postgres
    ... Output from starting container
    $ cat database.sql | docker-compose exec -T postgres psql -U postgres
    ... Output from running SQL commands
    $ docker-compose up frontend
    ... Output from building and running of container

The container can be exited with `Ctrl-C` or sleeped with `Ctrl-Z`. By passing in `-d` (see the line upping postgres), we can run the container in the background. It's output can be accessed with `docker-compose logs frontend`.

If the application is serving files that you will be iterating on often it's good to either use a tool that restarts the application inside of the container or "bind" the directory in the container to the directory in your repo. See the `volume` lines in the above `docker-compose.yaml` file.

Developers will create a new branch and begin literating on the application. When they complete a chunk of work, they create a PR against the main repo. Other developers can checkout and test their changes by switching branches in git and rebuilding:

    $ docker-compose stop frontend && docker-compose up --build frontend

Once the PR is approved it's merged into `master`. The updated application can the be deployed from the `master` branch.

## Deploying

Deploying the Docker image to AWS requires some step:

  * Create a AWS Container Repository and get credentials required to push container
  * Create database instances needed in RDS (in the `docker-compose.yaml` file)

Once any dependencies (like DBs) are up a new service can be created in Elastic Beanstalk, ECS, EKS, or whatever Docker container hosting service/platform you want. The setup of the container service will require knowing the ID of the container, the DNS, user, password, and db name from the RDS instance. Once the service is running the load balancer may require additional configuration (redirecting HTTP->HTTPS, pointing a domain name to the LB, and ensuring it has SSL/TLS certificates for that domain).






