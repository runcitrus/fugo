# Fugo

A flexible log parsing and processing agent for various log formats.

## Configuration

Fugo uses YAML configuration files to define agents and their behavior:

```yaml
agents:
  - name: nginx-access
    fields:
      - name: time
        time_format: common
      - name: status
        type: int
      - name: message
        template: "{{.method}} {{.path}}"
    file:
      path: /var/log/nginx/access.log
      format: plain
      regex: '^(?P<remote_addr>[^ ]+) - (?P<remote_user>[^ ]+) \[(?P<time>[^\]]+)\] "(?P<method>[^ ]+) (?P<path>[^ ]+) (?P<protocol>[^"]+)" (?P<status>[^ ]+)'
  - name: nginx-error
    fields:
      - name: time
        time_format: '2006/01/02 15:04:05'
      - name: level
      - name: message
    file:
      path: /var/log/nginx/error.log
      format: plain
      regex: '^(?P<time>[^ ]+ [^ ]+) \[(?P<level>[^\]]+)\] \d+#\d+: (?P<message>.*)'
```

- `name`: The name of the agent
- `fields`: A list of fields to store in the log records
- `file`: Configuration for file-based input

## Fields Configuration

Each field can be defined with:

- `name`: The name of the field in the output
- `source`: The source field to extract (defaults to the field name)
- `type`: The type of the field (e.g., `int`, `float`, `string`, `time`). Default is `string`, or `time` if `time_format` is specified.
- `template`: A Go template to transform source fields into the new field
- `time_format`: Format for the time field (e.g., `rfc3339`, `common`, `unix` or a custom Go layout)

## File-based Input

File-based input has the following configuration:

- `path`: Path to the log file or regex pattern to match multiple files.
- `format`: The format of the log file. Supported formats are `plain` and `json`.
- `regex`: A regex pattern to match the plain log lines. Named capture groups are used to extract fields.

### Path Configuration

The `path` can be a single file or a regex pattern. For example:

```yaml
path: '/var/log/nginx/access_(?P<host>.*)\.log'
```

A named capture group should be in the file name only and can be used in the fields.
