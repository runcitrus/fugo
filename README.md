# Fugo

A flexible log parsing and processing agent for various log formats.

## Configuration

Fugo uses YAML configuration files to define agents and their behavior:

```yaml
agents:
  - name: nginx-access
    file:
      path: /var/log/nginx/access.log
      format: plain
      regex: '(?P<remote_addr>[^ ]+) - (?P<remote_user>[^ ]+) \[(?P<time>[^\]]+)\] "(?P<method>[^ ]+) (?P<path>[^ ]+) (?P<protocol>[^"]+)" (?P<status>[^ ]+) (?P<body_bytes_sent>[^ ]+) "(?P<http_referer>[^"]*)" "(?P<http_user_agent>[^"]*)"'
      fields:
        - name: time
          time_format: common
        - name: status
        - name: message
          template: "{{.method}} {{.path}}"
  - name: nginx-error
    file:
      path: /var/log/nginx/error.log
      format: plain
      regex: '^(?P<time>[^ ]+ [^ ]+) \[(?P<level>[^\]]+)\] \d+#\d+: (?P<message>.*)'
      fields:
        - name: time
          time_format: '2006/01/02 15:04:05'
        - name: level
        - name: message
```

## File-based Input

File-based input has the following configuration:

- `path`: Path to the log file or regex pattern to match multiple files.
- `format`: The format of the log file. Supported formats are `plain` and `json`.
- `regex`: A regex pattern to match the plain log lines. Named capture groups are used to extract fields.
- `fields`: A list of fields to store in the log records.

### Path Configuration

The `path` can be a single file or a regex pattern. For example:

```yaml
path: '/var/log/nginx/(?P<host>.*)/access\.log'
```

A named capture group can be used in the fields.

### Field Configuration

Each field can be defined with:

- `name`: The name of the field in the output
- `source`: The source field to extract (defaults to the field name)
- `template`: A Go template to transform source fields into the new field
- `time_format`: Format for the time field (following Go's time format)
