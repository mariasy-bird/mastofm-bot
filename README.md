# mastofm-bot
A small Go daemon for scraping the most recent track posted on a given Last.FM account, and posting it to a Mastodon feed.

## Features
- Queries Last.fm for the most recent track of a specified user
- Maintains a persisted timestamp to deduplicate postings
- Formats and posts "Now listening:" updates to a Mastodon account, with album art and (simple) alt-text
- Outputs journal entries, designed to run as a "set it and forget it" systemd service

---
## Configuration
All configuration is provided via a `config.json` file located in the same directory as the compiled binary.
### Configuring Mastodon
Set the following fields:
- `mastodon_server`
  
  Base URL of your instance, including `https://`
  Example: `"https://mastodon.social"`
- `mastodon_client_id`
- `mastodon_secret`
- `mastodon_token`
  
To obtain these values:
1. Go to **Preferences->Development** in your Mastodon profile
2. Create a new Application, grant it **write:statuses** permission
3. Copy the client ID, client secret, and access token value from the Application

### Configuring Last.fm
- `lfm_username`
  
  Your Last.fm username.
- `lfm_api_key`
  
  Create an API key at: https://www.last.fm/api/account/create.

  Only the API key is needed, no additional tokens.

### Miscellaneous Configurations
- `test_mode`

    Set this to 'true' to 'dry fire,' only log new entries and do not post anything to Mastodon.

- `album_art`

    Set this to 'false' to only post the track name, no album art.

---
## Building
Clone this repository:
```bash
git clone https://github.com/mariasy-bird/mastofm-bot
cd mastofm-bot
```
Build the binary:
```bash
go build
```
Include your edited config.json in the same directory as your mastofm-bot executable.

## Running
You can run the bot directly in terminal:
```bash
./mastofm-bot
```
