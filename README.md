# tofu-template-tester

A lightweight Go CLI that renders Terraform-compatible templates by replacing ${NAME} placeholders with user-supplied values.

## Command-Line Parameters

*   `-list-params`: List all template variables found in the input template.
*   `-in <path>`: Path to the input template file. If not specified, input is read from STDIN.
*   `-properties <path>`: Path to a properties file (key=value format) containing variable assignments.
*   `-interactive`: Prompt for missing template variables interactively.
*   `-out <path>`: Path to write the output. If not specified, output is written to STDOUT.

## Example Usage

Suppose you have a template file named `template.txt` with the following content:

```
Hello, ${NAME}!
Your favorite color is ${COLOR}.
```

And a properties file named `vars.properties` with:

```
NAME=World
COLOR=Blue
```

You can render the template using the following command:

```bash
tofu-template-tester -in template.txt -properties vars.properties -out output.txt
```

This will produce the following output:

```
Hello, World!
Your favorite color is Blue.
```

Alternatively, you can use interactive mode:

```bash
tofu-template-tester -in template.txt -interactive
```

The tool will then prompt you to enter values for `NAME` and `COLOR`.
