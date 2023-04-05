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

Uses `owner` resource tag that can be:
 * Email: then `GetUserByEmail` slack API will be called to get slack user ID and send direct notification.
 * GitHub nickname: then `members` file will be used (if provided) to lookup slack user ID and send direct notification.

Members YAML file structure:
```yaml
members:
  <member-name>:
    slackID: <slack-id>
...
```

If `owner` resource tag is empty or invalid slack notification will be sent in the default Slack channel.

```console
$ c7n-helper slack -r <resource-file> \
                   -a <slack-auth-token> \
                   -m <members-file> \
                   -c <default-slack-channel-id> \
                   -t "<message-title>"
```

* Clean resources:

```console
$ c7n-helper clean -r <resource-file>
```

## License

Apache-2.0
