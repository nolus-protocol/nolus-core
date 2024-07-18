# Versioning
## Release a new version

1. Tag the source code in git with a version complying to [semantic versioning](https://semver.org/).

        git tag -a major.minor.patch -m "tag message"
        git push origin --tags

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
            --dns "192.168.10.200" --dns "192.168.10.202" \
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

10. Troubleshooting
    - DNS issues - the gitlab host name cannot be resolved by the runner or executor

    The problem stems from the smart way docker daemon setup the host resolution of containers. " When you run a new container on the docker host without any DNS related option in command, it simply copies host’s /etc/resolv.conf into container. While copying it filter’s out all localhost IP addresses from the file. That’s pretty obvious since that won’t be reachable from container network so no point in keeping them. During this filtering, if no nameserver left to add in container’s /etc/resolv.conf the file then Docker daemon smartly adds Google’s public nameservers 8.8.8.8 and 8.8.4.4 in to file and use it within the container" [ref](https://kerneltalks.com/networking/how-docker-container-dns-works/) 

    So if the gitlab is on a VPN and the gitlab runner host's host resolution is set (a) to add VPN dns entries in `/etc/resolv.conf` that leads to intermittent host resolution failures when non-VPN dns is being used, or (b) to use a local resolver, e.g. `systemd-resolved` which sets a local host address as a name server, for example 127.0.0.53, which docker then filters it out and replaces it with the Google's one causing host resolution failures.

    A workaround is to setup explicit `dns` on the GitLab Runner and executors, namely on the helper container that performs source code cloning. The Runner needs `--dns` parameter on the command line. The executors, incl. the helper container may be fixed specifying `dns = [""]` in `[runners.docker]` of the GitLab Runner `config.toml`.

# TBD
- [Build Docker images](https://docs.gitlab.com/ee/ci/docker/using_docker_build.html)
