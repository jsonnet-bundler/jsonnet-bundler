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

  steps: [
    golang() {
      name: 'gomod',
      commands: [
        'go mod vendor',
        'git diff --exit-code',
      ],
    },

    build('1.18'),
    build('1.19'),
    build('1.20'),

    golang() {
      name: 'generate',
      commands: [
        'make check-license',
        'make generate',
        'git diff --exit-code',
      ],
    },
  ],
}
