{
  _config+:: {
    golang: 'golang:1.13',
  },

  kind: 'pipeline',
  name: 'build',
  platform: {
    os: 'linux',
    arch: 'amd64',
  },

  local golang = {
    name: 'golang',
    image: $._config.golang,
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

  steps: [
    golang {
      name: 'gomod',
      commands: [
        'go mod vendor',
        'git diff --exit-code',
      ],
    },

    golang {
      name: 'build',
      commands: [
        'make build',
        'make test',
        'make test-integration',
      ],
    },

    golang {
      name: 'generate',
      commands: [
        'make check-license',
        'make generate',
        'git diff --exit-code',
      ],
    },
  ],
}
