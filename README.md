# narc

**narc** is an activity tracking tool. It logs named activities, tracks idle time, and aggregates work sessions.

## Features

- Start and stop named activities
- Cross-platform idle detection (macOS and Windows)
- Background daemon for continuous tracking
- Aggregate time logs by date or range
- Configurable via command-line

## Installation

If you have Go installed:

```
go install github.com/alexrjones/narc/cmd/narc@latest
```

Make sure `$GOBIN` is in your `PATH`, or move the binary to a directory that is.

## Usage

```
narc <command> [flags]
```

Run `narc <command> --help` to get more detailed usage per command.

### Commands

#### `start <nameparts> ...`

Start a named activity. The name can contain multiple words.

```
narc start writing weekly report
```

#### `end`

End the current activity.

```
narc end
```

#### `status`

Get the current daemon and activity status.

```
narc status
```

#### `aggregate` / `agg` `[<start> [<end>]]`

Summarize logged activity within a date or range.

Examples:

```
narc aggregate today
narc agg 2024-01-01 2024-01-31
```

#### `daemon`

Start the background tracking daemon (if not already running).

```
narc daemon
```

#### `terminate`

Stop the background daemon.

```
narc terminate
```

#### `config show`

Show the current configuration.

```
narc config show
```

#### `config get <name>`

Get a specific config value.

```
narc config get idleTimeout
```

#### `config set <name> <value>`

Set a config option.

```
narc config set idleTimeout 300
```

## Configuration

Config values are stored per user and control things like idle timeout and data paths. You can view and modify them using the `config` subcommands.

## Notes

- The `daemon` must be running to track idle time.
- All logs are stored locally.
- narc uses platform-native APIs for idle detection (no polling loops or elevated privileges required).
