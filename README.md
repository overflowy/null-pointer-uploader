# npu

# [![npu](static/screenshot.png)](https://github.com/overflowy/null-pointer-uploader)

Simple GO client for [The Null Pointer](https://0x0.st/) file-sharing service.

## How it works

When executed, the program reads the specified file and creates a multipart/form-data POST request to send to the 0x0.st API. The uploaded file is included as a form file with the name file. The expiration and secret options, if specified, are also included as form fields.

Upon a successful upload, the program outputs the URL to the uploaded file to the console and also **writes it to the clipboard**.

### Uploading multiple files

When using `-dir` mode, the program will loop through all files in the directory provided and upload each one. Upon completing all uploads, the program will output a table with the files uploaded and the URL generated. This output will also be saved to a file in the source directory naned `upload.xxxxx.log`.

The same functionality is available if more than one arg is provided (e.g. `npu file1.jpg file2.jpg`) but the log file will be stored in the executable's root directory.

## To do

- [x] Add support for multiple files
- [x] Add support for directories
- [ ] Add support for archiving files (with encryption)
- [ ] Add configuration file

## License

The code in this repository is licensed under the MIT License. See LICENSE for more information.
