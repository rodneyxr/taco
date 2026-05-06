# Rotation

Taco keeps rotation state in a separate YAML file so `taco.yaml` can stay focused on the team roster.

By default, the state file is written to:

```text
<user config dir>/taco/state.yaml
```

On macOS, that is usually:

```text
~/Library/Application Support/taco/state.yaml
```

You can choose a different file with `--state`:

```sh
taco --state ./state.yaml --message "thanks for helping"
```

## What Gets Stored

State is stored per team. A state file looks like this:

```yaml
teams:
  default:
    index: 2
    order:
      - @bob
      - @charles
      - @alice
```

Each team has:

- `order`: the current shuffled rotation order for that team
- `index`: the next position to read from that order

The usernames are stored as Slack mentions because taco normalizes every configured username before selection. For example, `alice` and `"@alice"` both become `@alice`.

## How Selection Works

When a template calls a team:

```gotemplate
{{ team "default" }}
```

taco does this:

1. Reads the members for `default` from `taco.yaml`.
2. Normalizes usernames by stripping any leading `@` and adding one back.
3. Removes duplicate mentions while preserving the roster order from `taco.yaml`.
4. Loads the saved `order` and `index` for that team from the state file.
5. Removes any saved names that are no longer in `taco.yaml`.
6. Adds any new names from `taco.yaml` into the saved order after shuffling them.
7. Selects up to `--number-of-tacos` unique members from the saved order.
8. Saves the updated `index` and `order` back to the state file.

If the selection reaches the end of the saved order, taco starts a new shuffled cycle.

## Example

Given this config:

```yaml
tacos: 2
teams:
  default:
    - alice
    - bob
    - charles
```

The first run may create state like:

```yaml
teams:
  default:
    index: 2
    order:
      - @bob
      - @charles
      - @alice
```

That means the first run selected:

```text
@bob @charles
```

The next run starts at `index: 2`, so it selects `@alice` and then starts a new shuffled cycle for the second taco.

## No Rotate

Use `--no-rotate` to preview a message without reading or writing the state file:

```sh
taco --no-rotate --message "preview"
```

This is useful when testing a template or checking output before posting in Slack.

## Multiple Teams

State is tracked independently for each team:

```yaml
teams:
  default:
    index: 2
    order:
      - @bob
      - @charles
      - @alice
  siem:
    index: 1
    order:
      - @justin
      - @charles
      - @eboyer
```

A template can use a team with:

```gotemplate
{{ team "siem" }} {{ .Message }} {{ .Emoji }}
```

Each team advances only when it is used in the rendered template.

## Changing The Roster

It is safe to edit `taco.yaml`.

When a name is removed from a team, taco drops it from that team's saved rotation order the next time the team is used.

When a name is added to a team, taco adds it into that team's saved rotation order the next time the team is used.

If you want to start fresh, delete the state file or pass a new path with `--state`.
