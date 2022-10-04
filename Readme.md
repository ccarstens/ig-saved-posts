# IG Saved Posts

With this tool you can download the saved posts from your instagram account(s).
The download is as fast as the Instagram API allows.
Your credentials are needed once when the tool is first run, after that the existing session is reused (just like when logging in to the Instagram app).
The first sync can take some time because the Instagram API only allows for posts to be retrieved in a paginated way, so the tool has to simulate scrolling through all the posts.
The folowing syncs will be much faster because only the content and albums that have recently changed are synced.


IG Saved Posts uses [Goinsta](https://github.com/Davincible/goinsta) under the hood.

## Usage
Both the macOS and Windows files can be executed by being double clicked.
On the first run you will be asked to specify the folder where your posts will go.
e.g. 
```
~/Pictures/Instagram
```
The tool will create all the required folders for you.

### Full Sync
In case errors have interrupted a sync or local files have been deleted a full sync (without skipping albums that haven't changed recently) can be triggered by passing the `--all` flag:
```bash
$ path/to/executable/IgSavedPosts-macOS-amd64 --all
```
