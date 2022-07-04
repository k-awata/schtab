# schtab

schtab sets tasks to Windows Task Scheduler from a text in [crontab](https://man7.org/linux/man-pages/man5/crontab.5.html) format.

## Installation

If you're using Go:

```bash
go install github.com/k-awata/schtab@latest
```

Otherwise you can download a binary from [Releases](https://github.com/k-awata/schtab/releases).

## Usage

- Edit your schtab file

  ```bash
  schtab e
  ```

  To change the default editor, set `EDITOR` or `VISUAL` environment variable as follows:

  ```bash
  export EDITOR="code --wait"
  ```

- Write schtab from a file

  ```bash
  schtab file.txt
  ```

- Write schtab from stdin

  ```bash
  schtab - < file.txt
  ```

- List your schtab file

  ```bash
  schtab l
  ```

- Remove your schtab file

  ```bash
  schtab remove
  ```

## Not supported in schtab file

- Setting environment variables
- Setting range of values for **minute**, **hour**, and **day of month**

## License

[MIT License](LICENSE)
