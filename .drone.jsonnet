{
  kind: 'pipeline',
  name: 'default',
  platform: {
    os: 'linux',
    arch: 'amd64',
  },

  local golang(version='latest') = {
    name: 'golang',
    image: 'golang:%s' % version,
    pull: 'always',
    environment: {
      CGO_ENABLED: '0',
      GO111MODULE: 'on',
    },
    when: {
      event: {
        exclude: ['tag'],
      },
    },
  },

  local build(version) = golang(version) {
    name: 'build-%s' % version,
    commands: [
      'make build',
      'make test',
      'make test-integration',
    ],
  },

  local goreleaser(version) = golang(version) {
    commands: ['curl -sL https://git.io/goreleaser | bash'],
    environment+: { GITHUB_TOKEN: { from_secret: 'github_token' } },
    name: 'goreleaser',
    when: { event: 'tag' },
  },

  local docker_release() = {
    image: 'plugins/docker',
    pull: 'always',
    name: 'docker',
    settings: {
      auto_tag: true,
      password: { from_secret: 'docker_password' },
      repo: 'jsonnet/bundler',
      target: 'production',
      username: { from_secret: 'docker_username' },
    },
    when: { event: 'tag' },
  },

  steps: [
    golang() {
      name: 'gomod',
      commands: [
        'go mod vendor',
        'git diff --exit-code',
      ],
    },

    build('1.11'),
    build('1.12'),
    build('1.13'),
    build('1.14'),

    golang() {
      name: 'generate',
      commands: [
        'make check-license',
        'make generate',
        'git diff --exit-code',
      ],
    },

    goreleaser('1.14'),

    docker_release(),
  ],
}
