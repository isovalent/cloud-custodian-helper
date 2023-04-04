# c7n-helper

The tool helps to work with Cloud Custodian generated reports.

## Installation

To install the latest `c7n-helper` release run:

```console
$ go install github.com/isovalent/cloud-custodian-helper@latest
```

## Lint

To lint `c7n-helper` sources please run the following locally:

```console
$ make lint
```

## Build

To build `c7n-helper` from source please run the following locally:

```console
$ make build
```

## Usage

* Help:

```console
$ c7n-helper --help
```

* Parse C7N output directory into JSON file:

```console
$ c7n-helper parse -d <c7n-report-dir> -p <c7n-policy-name> -t <resource-type> -r <resource-file>
```

* Send Slack notification:

If members YAML file specified it will try to find resource owners Slack IDs and send direct message notifications.
Otherwise, it will send notification in the default SLack channel.

```console
$ c7n-helper slack -r <resource-file> \
                   -a <slack-auth-token> \
                   -m <members-file> \
                   -c <default-slack-channel-id> \
                   -t "<message-title>"
```

YAML file structure example:
```yaml
members:
  <member-name>:
    slackID: <slack-id>
...
```

* Clean resources:

```console
$ c7n-helper clean -r <resource-file>
```

## License

Apache-2.0
