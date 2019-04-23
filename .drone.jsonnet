{
  _config+:: {
    golang: 'golang:1.12',
  },

  kind: 'pipeline',
  name: 'build',
  platform: {
    os: 'linux',
    arch: 'amd64',
  },

  steps: [
    {
      name: 'build',
      image: $._config.golang,
      pull: 'always',
      environment: {
        CGO_ENABLED: '0',
        GO111MODULE: 'on',
      },
      commands: [
        'make',
      ],
      when: {
        event: {
          exclude: ['tag'],
        },
      },
    },
  ],
}
