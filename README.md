# taco

Create Slack messages for [HeyTaco](https://heytaco.com/) so you can rotate through teammates, copy the message, and post it yourself.

## Quickstart

Install from source:

```sh
go install github.com/rodneyxr/taco@latest
```

Create `taco.yaml` in your current directory, or `~/.taco.yaml`:

```yaml
tacos: 5
teams:
  default:
    - alice
    - bob
    - charles
    - don
    - eric
```

Generate a message:

```sh
taco --config taco.yaml --number-of-tacos 5 --message "thanks for helping"
```

Output:

```text
@alice @bob @charles @don @eric thanks for helping 🌮
```

The next run advances the rotation, so larger teams get spread out over time. Use plain usernames in YAML; if a quoted value includes `@`, taco strips it before adding the mention prefix.

Rotation details are documented in [doc/rotation.md](doc/rotation.md).

## Flags

```text
-c, --config string           config file (default "taco.yaml")
-n, --number-of-tacos int     number of teammates to include (default 5)
-m, --message string          optional message
    --emoji string            emoji to include (default "🌮")
    --template string         Go message template (default "{{ team \"default\" }} {{ .Message }} {{ .Emoji }}")
-t, --team string             team to use with the default template (default "default")
    --state string            rotation state file
    --no-rotate               preview without advancing rotation
    --version                 print build version
```

Use another team:

```yaml
teams:
  platform:
    - ada
    - grace
    - linus
```

```sh
taco --team platform --message "ship it"
```

Use a custom Go template:

```sh
taco --template '{{ team "platform" }} {{ .Message }} {{ .Emoji }}' --message "ship it"
```

## Releases

This repo uses conventional commits, Release Please for semantic version tags, and GoReleaser for binaries. Merging conventional commits to `main` creates releases for `amd64` and `arm64`.
