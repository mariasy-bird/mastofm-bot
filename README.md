# mastofm-bot
A Go bot for scraping the most recent track posted on a given Last.FM account, and posting it to a Mastodon feed.

# Configuring, Building

## Configuring
Edit the config.json file. 

For the mastodon_server value, use the base URL of your instance, leading with https://, e.g. "https://mastodon.social".

To get the client ID, secret, and token, on your Mastodon account, create an app in Preferences -> Development. 

You will need a client ID, a secret, and an access token. The application will only need write:statuses permissions.


For the lfmUsername field, substitute the placeholder with your last.fm username. 

For the lfmApiKey field, create an API account at https://www.last.fm/api/account/create. You will only need API_KEY, no special tokens.

## Building
Clone this repository:
```
$ git clone https://github.com/mariasy-bird/mastofm-bot
```
Run the following commands to build the program:
```
$ go get mastofm-bot

$ go build
```
Include your edited config.json in the same directory as your mastofm-bot executable.
