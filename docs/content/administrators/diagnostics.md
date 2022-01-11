---
title: Collect Diagnostics from Porter
description: How to configure Porter to generate logs and telemetry data for diagnostic purposes
---

When the [structured-logs experimental feature][structured-logs] is enabled, Porter generates two types of data to assist with diagnostics and troubleshooting:

* [Logs](#logs)
* [Telemetry](#telemetry)

## Logs

Porter can be configured to write logs to the PORTER_HOME/logs directory, for example ~/.porter/logs.
Each time a Porter command is executed, a new log file is created, formatted in json, containing all the logs output by the command.
The log lines are filtered by the current log level.

The name of each log file is the command's _correlationId_ which can be used to find the trace for the command executed in the configured [telemetry](#telemetry) backend.
See [Log Settings] for details on how to configure logging.

## Telemetry

Porter is compatible with the [OpenTelemetry] specification and generates trace data that can be sent to a [compatible services][compat].
When the [structured-logs experimental feature][structured-logs] and [telemetry] is enabled, Porter automatically uses the standard [OpenTelemetry environment variables] to configure the trace exporter.

Below is an example trace from running the porter upgrade command. You can see timings for each part of the command, and relevant variables used.

![Screen shot of the Jaeger UI showing that porter upgrade was run](/administrators/jaeger-trace-example.png)

If you are running a local grpc OpenTelemetry collector, for example with the [otel-jaeger bundle], you can set the following environment variables to have Porter send telemetry data to it. This turns on the [structured-logs experimental feature][structured-logs], enables telemetry, and uses standard OpenTelemetry environment variables to point to an unsecured grpc OpenTelemetry collector.

* PORTER_EXPERIMENTAL: structured-logs
* PORTER_TELEMETRY_ENABLED: true
* OTEL_EXPORTER_OTLP_PROTOCOL: grpc
* OTEL_EXPORTER_OTLP_ENDPOINT: 127.0.0.1:4317
* OTEL_EXPORTER_OTLP_INSECURE: true

See [Telemetry Settings][telemetry] for all the supported configuration settings.

[compat]: https://opentelemetry.io/vendors/
[OpenTelemetry environment variables]: https://github.com/open-telemetry/opentelemetry-specification/blob/v1.8.0/specification/protocol/exporter.md
[telemetry]: /configuration/#telemetry
[Log Settings]: /configuration/#logs
[structured-logs]: /configuration/#structured-logs
[OpenTelemetry]: https://opentelemetry.io
[otel-jaeger bundle]: https://github.com/getporter/example-bundles/tree/main/otel-jaeger
