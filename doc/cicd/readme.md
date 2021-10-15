# Setup Gitlab Runner

1. Install Docker

    - install
    - make sure no root access is required
    - start the service

2. Download Gitlab Runner Image

        docker pull gitlab/gitlab-runner:alpine

3. Register

        docker run --rm -it -v ~/.gitlab-runner:/etc/gitlab-runner gitlab/gitlab-runner:alpine register \
        --executor docker

    Dump the server certificate in PEM format file in ~/.gitlab-runner/certs/gitlab-nomo.credissimo.net.crt from the output of

        echo |     openssl s_client -connect  gitlab-nomo.credissimo.net:443 2>/dev/null |     openssl x509 -text

4. Configure the cache storage directory as persistent

    Define `volumes = ["/path/to/host/dir/.gitlab-runner-cache:/cache"]` under the `[runners.docker]` section in `~/.gitlab-runner/config.toml`

5. Make sure all relevant directories and files are marked as cache

    Use `<job>:cache:` in your Gitlab pipeline yaml descriptor. [Ref](https://docs.gitlab.com/ee/ci/yaml/#cache)

6. Start

    [Source](https://docs.gitlab.com/runner/install/docker.html#option-1-use-local-system-volume-mounts-to-start-the-runner-container)

        docker run -d --name gitlab-runner --restart always \
            -v ~/.gitlab-runner:/etc/gitlab-runner \
            -v /var/run/docker.sock:/var/run/docker.sock \
            gitlab/gitlab-runner:alpine

7. Logs

        docker logs gitlab-runner

8. Housekeeping

        docker system prune

9. References

    - [Docker images to run into](https://docs.gitlab.com/ee/ci/docker/using_docker_images.html)
    - [Gitlab CI/CD pipeline reference](https://docs.gitlab.com/ee/ci/yaml/)
    - [Docker runner executor](https://docs.gitlab.com/runner/executors/docker.html)
    - [Docker runner advanced configuration](https://docs.gitlab.com/runner/configuration/advanced-configuration.html#the-runnersdocker-section)
    - [Caching](https://docs.gitlab.com/ee/ci/caching/index.html)

# TBD
- [Build Docker images](https://docs.gitlab.com/ee/ci/docker/using_docker_build.html)
