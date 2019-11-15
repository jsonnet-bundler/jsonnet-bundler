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
    depends_on: ["gomod"]
  },

  steps: [
    golang() {
      name: 'gomod',
      commands: [
        'go mod vendor',
        'git diff --exit-code',
      ],
    },

    build('1.13'),
    build('1.12'),
    build('1.11'),

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
