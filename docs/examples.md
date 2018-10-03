# Usage

```
Usage:
  faq [flags] [filter string] [files...]

Flags:
  -c, --color-output           colorize the output (default true)
  -h, --help                   help for faq
  -f, --input-format string    input format (default "auto")
  -m, --monochrome-output      monochrome (don't colorize the output)
  -o, --output-format string   output format (default "auto")
  -p, --pretty-output          pretty-printed output (default true)
  -r, --raw-output             output raw strings, not JSON texts
```

# Examples

Note: Example files can be found in the /docs/examples/ directory

## Piping to make something legible

Input
```sh
cat docs/examples/unformatted.json | faq
```

Output
```json
{
  "hello": {
    "world": "whats up"
  },
  "with": "you"
}
```

## Reading a raw string value from a YAML file

Input
```sh
faq -r '.company' docs/examples/sample.yaml
```

Output
```
Awesome Code inc.
```
