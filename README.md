# schtab

schtab sets tasks to Windows Task Scheduler from a text in [crontab](https://man7.org/linux/man-pages/man5/crontab.5.html) format.

## Usage

- Edit your schtab file

  ```powershell
  schtab -e
  ```

  To change the default editor, set `EDITOR` or `VISUAL` environment variable.

- Write schtab from a file

  ```powershell
  schtab file.txt
  ```

- Write schtab from stdin

  ```powershell
  schtab - < file.txt
  ```

- List your schtab file

  ```powershell
  schtab -l
  ```

- Remove your schtab file

  ```powershell
  schtab --delete
  ```

## Not supported in schtab file

- Setting environment variables
- Intervals on months (e.g. `* * * */5 *`)
- Intervals on DOW (e.g. `* * * * MON/3`)

## License

[MIT License](LICENSE)
