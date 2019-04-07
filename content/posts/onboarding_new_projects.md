# Onboarding developers quickly
<div id="published-at">2019-04-06T19:09:00Z</div>

Onboarding developers can require a significant amount of time, but it doesn't have to. This article goes over a minimal set of tools that provide a uniform developer environment and workflow for all major operating systems. We will go over why each tool is valuable, how we use the tools to increase velocity, and a short outline on how to deploy to AWS.

### README.md

Always create a `README.md` in the root of your project. It should include a short description of the project and its setup instructions. It's also useful to list key maintainers/contributors and the project's license. If you only do one thing from this article, please document your setup instruction in the project's `README.md`. Nothing kills a project faster than not having easy to find setup instructions. It's also a good practice to document important environment variables.

### Git & GitHub

Git and the repository hosting service GitHub, and similar version control tools, are universally required at organizations with teams of developers. Open Source projects are no exception. All commonly used software is tracked and organized with version control software. 

Without VC software teams are not able to easily compare previous versions, switch between tickets, or reliably resolve file conflicts. When two developers edit the same file, the last developer to upload wins and the other's changes are lost. This problem is solved by version control; When the two developers commit and push their changes to the repository the last developer to push must pull the other developer's changes and resolve any merge conflict. Resolving merges can be easy or hard depending on the situation, but regardless of the complexity the alternative - losing another developer's work - is unacceptable.

The learning curve of Git and GitHub is initially a little steep. It's not uncommon to encounter people that know a little scripting but have never worked on a team. It's also not unusual for lone wolf developers to be resistant to using VC. When starting a project, ensure that everyone is willing to learn and use VC. Developers not familiar with Git or Github can read [GitHub's guide](https://guides.github.com/introduction/git-handbook/) and should receive help from the team when they get stuck. I strongly recommend not working with developers that refused to use VC, they will cause more problems than they solve and the velocity of the team will suffer. 

### Docker & Docker Compose

Ok, everyone can check out the repo. Great! But the project requires some initial setup. Docker and Docker Compose will make it easy for developers to stand up their development environment, including databases and other dependencies. They also allow easy publishing of images to the cloud for deployment. The earlier a team sets up Docker the sooner they will benefit from the improved velocity, standardized developer environments, and retention of maintainers and collaborators. The benefits are cumulative, if you wait until after you burn people out or they struggle to get onboard you've already lost valuable time and energy.   

## Dockerizing a project

First, if you're using a common platform, like Wordpress, someone else has already created the Docker-related files. You only need to find them, confirm they work, check them into your repo, and update your `README.md`. 

From scratch, projects must decide the programming platform they are using (Node.js, Go, PHP, etc...) and go to [Docker Hub](https://hub.docker.com/search?q=&type=image) and find the Docker Image for the platform being used. Create a `Dockerfile` at the root of the repo and fill it out with the setup steps. Here is an example from [Hack4Eugene/SpeedUpYourCity](https://github.com/Hack4Eugene/SpeedUpYourCity):

``` 
FROM ruby:2.3.1-alpine

RUN apk add --no-cache mariadb-dev make g++ linux-headers nodejs tzdata

WORKDIR /suyc
COPY . .

RUN bundle install
RUN rake assets:precompile

EXPOSE 3000
CMD ["rails", "server", "-b", "0.0.0.0"]
```

The SpeedUpYourCity project uses Ruby v2.3.1, requires a handful of OS packages, and is stored in the containers `/suyc` directory. After copying the contents of the repo into `/suyc`, `bundle install` and `rake assets:precompile` are run to install platform libraries and build up-to-date copies of assets (not required for all projects). This app contains an HTTP server listening on port 3000 and is started by running `rails server -b 0.0.0.0`. That's it. This file will vary significantly from project to project. Some projects will use an OS image (`ubuntu`) or another language platform, have different build steps.

Once the `Dockerfile` is complete, it's possible to build a Docker image for your application and run a container, but with a little bit more work we can make the whole process much simpler. Docker provides a tool called Docker Compose that allows developers to concretely define the environment and dependencies (MySQL, PostgreSQL, MongoDB, etc...). 

SpeedUpYourCity uses MySQL to store speed test submissions. We must have the database running for the application to start. We can define and control our entire stack with a `docker-compose.yaml` file:

```
version: '3.2'
services:
  mysql:
    image: mysql:5.7
    environment:
      MYSQL_ROOT_PASSWORD: suyc
      MYSQL_USER: suyc
      MYSQL_PASSWORD: suyc
      MYSQL_DATABASE: suyc
    ports:
      - "3306:3306"
  frontend:
    build: .
    environment:
      DB_HOSTNAME=mysql
      DB_PORT=3306
      DB_USERNAME=suyc
      DB_PASSWORD=suyc
      DB_NAME=suyc
    ports:
      - "3000:3000"
    volumes:
      - .:/suyc
```

Ignore the version, the key bits are the services (mysql and frontend). From the Docker Hub page for the `mysql:5.7` image we know that the image allows us to define a few environment variables. When the database service is started, it will be using the environment variables to create a user and database. We can then pass the same values to the frontend service. The hostname for the DB is the same as it's service name (`mysql`). Docker Compose also allows us to bind our repo's files to a directory in the started container. The defined volume does just that. Now the local repo files and the files in the container's `/suyc` directory are the same files, allowing much faster iterations as the container doesn't need to be rebuilt to test changes. 

With both of these files, we can now start and stop the entire stack with `docker-compose up` and `docker-compose down`. We can also run the containers in the background by adding `-d` to the up command.

One last thing, many platforms require setting up a database schema. This is trivial with Docker Compose. After defining the above files and start the containers we run our platforms DB migration tool inside of the frontend container with `docker-compose run frontend rake db:setup`. When you have all of these files defined and tested make sure to update your `README.md` with the setup instructions.

## Onboarding and workflow

It's time to onboard new developers. The first thing a developer should do is clone the repo:

    $ git clone <fork repo url>
    $ cd <repo dir>

Next, they look at the project's `README.md` and perform any setup steps. Most projects will require starting the database, running a command to create the schema and required data, and  instructions on how to start the application. Those steps will roughly look like:

    $ docker-compose up -d postgres
    ... Output from starting container
    $ cat database.sql | docker-compose exec -T postgres psql -U postgres
    ... Output from running SQL commands
    $ docker-compose up frontend
    ... Output from building and running of container

The container can be exited with `Ctrl-C` or suspended with `Ctrl-Z`. By passing in `-d` (see the line upping postgres), we can run the container in the background. Its output can be accessed with `docker-compose logs frontend`.

Developers will create a new branch and begin iterating on the application. When they complete a chunk of work, they create a Pull Request against the main repo in GitHub. Other developers can check out and test the PR's changes by switching branches with Git and rebuilding the container:

    $ docker-compose stop frontend && docker-compose up --build frontend

Once the PR is approved, it's merged into `master`. The updated application can then be deployed from the `master` branch. Using GitHub add-ons, it's possible to automatically deploy changes to the cloud.

## Deploying

Deploying the Docker image to AWS requires some setup:

  * Create an AWS Container Repository and get credentials needed to push application images
  * Push images built from master to the container image repository
  * Create database instances needed using RDS

Once your database is ready, create a new service in Elastic Beanstalk, ECS, EKS, or whatever service/platform that supports running Docker containers. The setup of the container orchestration service will require knowing the ID of the published image and the details of your RDS instance. Once the service is running, a load balancer may need additional configuration (redirecting HTTP->HTTPS, pointing a domain name to the LB, and ensuring it has SSL/TLS certificates for that domain). Deploying to the cloud requires some system administration knowledge, setting up deployments to the cloud should be done by a senior developer.

> Google Cloud and Azure provide managed database services, similar to AWS RDS, and container orchestration services, similar to Beanstalk/ECS/EKS. Their are may cloud providers of various sizes and you can be mostly cloud provider agnostic using Kubernetes. I use AWS because that's what I'm familiar with and the volume of information available on how to setup Continuous Integration and Deployment. The big providers work rougtly the same, but each has their own idiosyncrasies. 

## Wrap-up

Software projects live and die based on ease of setup. Nobody wants to spend hours setting up codebase or spend additional hours setting up dependencies a developer added without documentation. By using tools like Git, GitHub, Docker, and AWS, you can streamline developer onboarding, workflows, and deployments. Git and GitHub will allow everyone to work across the codebase without worrying about overwriting each other's work. Docker will enable your team to stand-up their local environments in a uniform way and efficiently manage changes to the environment. Once you're able to make Docker images for your project, deployment to the Cloud is much easier, and techniques like Continuous Deployment are within reach. Make your life easier and use these tools from the start.
