# initial-configs-go #

It's a simple library to try to not repeat myself on
configurations on each new project I start.

It configures `Viper` configuration, `slog` structured
(including log keys formatting) logs and maybe another
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