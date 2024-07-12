# Nextcloud Status Command

Set your Nextcloud status from the command line.

## Usage

### Log in to your Nextcloud server

**Note:** Create an app password first if you are using two-factor authentication.

Run `nsc auth` to log in to your Nextcloud.
Enter your server address (e.g. `https://my.cloud.com`), your username and your password.
Submit the form to save the credentials on your disk.

### Update your Status

Run `nsc` to set your status.
A form will be shown that guides you through the options.
You can update your status, emoji and message.

Exit anytime by pressing `ctrl+c`, `q` or `esc`.

### Clear your status message

Run `nsc clear` to clear your status message.

### Get your current status

Run `nsc get` to print your current status, emoji and message.

## Build

Run `make` or `go build -o nsc cmd/nsc/main.go` to build a binary at `./nsc`.

## Acknowledgements

Emoji data is taken from the [gemoji](https://github.com/github/gemoji/blob/master/db/emoji.json)
library.
The gemoji library is available under the [MIT](https://github.com/github/gemoji/blob/master/LICENSE) license.
