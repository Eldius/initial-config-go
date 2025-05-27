# initial-configs-go #

It's a simple library to try to not repeat myself on
configurations on each new project I start.

It defines `Viper`'s initial configuration,
`slog` structured logs (including log keys
formatting) and maybe some other
things to help application startup.


## Default configuration keys ##

```yaml
---
debug: true
log:
  level: debug
  format: json
  output_to_file: execution.log
  output_to_stdout: false

```
