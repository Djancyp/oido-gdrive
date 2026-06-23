# Oido Google Drive

Read and write Google Drive files. Authenticated with your own Google OAuth
connection (set up in the extension settings → Connect with Google).

## Setup

1. In Google Cloud, create an OAuth client (type: Web application) and add the
   redirect URI shown by your Oido studio: `https://<studio>/api/oauth/callback`.
2. In the extension settings, paste the **Client ID** and **Client secret**, Save.
3. Click **Connect with Google** and grant access.

## Tools

- `drive_list` — list/search files (optional Drive query; empty = recent).
- `drive_read_file` — read a file's text by ID (Google Docs exported to text).
- `drive_create_file` — create a text file (optionally in a folder).
- `drive_update_file` — overwrite a file's content.
- `drive_delete_file` — permanently delete a file.
