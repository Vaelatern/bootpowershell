# On Boot, Run Arbitrary Powershell

Usage: `bootpowershell.exe [install] E:`

- Pass install to use `schtasks` to install.
- Set the last parameter to specify where yaml files can be found
- Commands will be a list of strings under raw_ps or raw_cmd, appended in lexical order

Valid files end in `.yml`. So if you do `.yml.no` then that won't be picked up and parsed as yaml.

Example yaml file:

```
raw_ps:
  - Expand-Archive -Force C:\path\to\archive.zip C:\where\to\extract\to
raw_cmd:
  - >
     chdir C:\where\to\extract\to &&
     dir
  # this last command doesn't do anything, I just wanted to demonstrate cmd commands
```

Files are parsed in lexical order. raw_ps and raw_cmd are executed, first the Powershell commands, then the regular commands, executed file by file. If you need to toggle from powershell to cmd and then back, you want multiple files in lexical order.

Remember that Windows doesn't always know about capitalization. Best to make your lexical order clear with numeric starts to your filenames.
