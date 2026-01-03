# FactoCord 3.0 - Discord Commands Tutorial

This document describes all available Discord commands for FactoCord 3.0 with detailed examples and testing instructions.

## Table of Contents
- [Overview](#overview)
- [Admin Commands](#admin-commands)
  - [server](#server)
  - [save](#save)
  - [kick](#kick)
  - [ban](#ban)
  - [unban](#unban)
  - [config](#config)
  - [mod](#mod)
- [Utility Commands](#utility-commands)
  - [mods](#mods)
  - [version](#version)
  - [info](#info)
  - [online](#online)
  - [help](#help)
- [DM Mod-Settings Manager](#dm-mod-settings-manager)
  - [mods (DM)](#mods-dm)
  - [modshelp](#modshelp)
  - [save (DM)](#save-dm)
  - [cancel](#cancel)

## Overview

FactoCord 3.0 provides various Discord commands to manage and monitor your Factorio server. Commands are divided into three categories:

- **Admin Commands**: Require admin rights (configured in `config.json` under `admin_ids`) or a specific role
- **Utility Commands**: Available to all users
- **DM Mod-Settings**: Manage Factorio mod settings via Discord DMs (requires verification)

The default command prefix is `$` (configurable in `config.json`).

---

## Admin Commands

### server

**Description:** Manages the Factorio server (start, stop, restart, updates).

**Permissions:** 
- Status query: All users
- Control: Admins only

**Usage:**
```
$server
$server stop
$server start
$server restart
$server update [version]
$server install [version]
```

**Subcommands:**

#### $server
Displays the current server status (running or stopped).

**Example:**
```
$server
```
**Expected Output:** `Factorio server is **running**` or `Factorio server is **stopped**`

---

#### $server stop
Stops the running Factorio server.

**Example:**
```
$server stop
```
**Expected Output:** Server shuts down, status message is sent.

**Test:** Run `$server` after stopping to confirm the server is stopped.

---

#### $server start
Starts the Factorio server.

**Example:**
```
$server start
```
**Expected Output:** Server starts, status message is sent.

**Test:** Run `$server` after starting to confirm the server is running.

---

#### $server restart
Restarts the Factorio server (stop + start).

**Example:**
```
$server restart
```
**Expected Output:** Server is stopped and then restarted.

**Test:** Check with `$server` if the server is running again after restart.

---

#### $server update [version]
Updates the Factorio server to the latest or a specific version.

**Note:** Server must be stopped before updating.

**Examples:**
```
$server update
$server update 1.1.100
```

**Expected Output:** 
- Download progress is displayed
- Unpacking message
- Success message with version number

**Test:**
1. Stop the server: `$server stop`
2. Update to latest version: `$server update`
3. Check version after start: `$version`

---

#### $server install [version]
Installs a Factorio server version without version check (similar to update, but without checking current version).

**Examples:**
```
$server install
$server install 1.1.100
```

**Test:** Same as `$server update`

---

### save

**Description:** Saves the current game.

**Permissions:** Admin

**Usage:**
```
$save
```

**Example:**
```
$save
```

**Expected Output:** Save operation is executed on the Factorio server, confirmation message in Discord.

**Test:** 
1. Run `$save`
2. Check Factorio server logs for `/save` command
3. Look for Discord message confirming successful save

---

### kick

**Description:** Kicks a player from the server with a specified reason.

**Permissions:** Admin

**Usage:**
```
$kick <playername> <reason>
```

**Parameters:**
- `<playername>`: Name of the player to kick
- `<reason>`: Reason for the kick (multiple words allowed)

**Example:**
```
$kick PlayerName Rule violation
$kick TestUser Inactive for 30 minutes
```

**Expected Output:** `Player PlayerName kicked with reason Rule violation!`

**Test:**
1. Connect to the server with a test account
2. Run `$kick TestUser Test reason`
3. Confirm that the player was disconnected from the server
4. Check the kick message in Discord and on the Factorio server

---

### ban

**Description:** Bans a player from the server with a specified reason.

**Permissions:** Admin

**Usage:**
```
$ban <playername> <reason>
```

**Parameters:**
- `<playername>`: Name of the player to ban
- `<reason>`: Reason for the ban (multiple words allowed)

**Example:**
```
$ban PlayerName Cheating
$ban GrieferUser Griefing and repeated violations
```

**Expected Output:** `Player PlayerName banned with reason "Cheating"!`

**Test:**
1. Ban a test user: `$ban TestUser Test ban`
2. Try connecting with this account (should fail)
3. Check the ban list on the server
4. Remove the ban with `$unban TestUser` after testing

---

### unban

**Description:** Removes a player from the ban list.

**Permissions:** Admin

**Usage:**
```
$unban <playername>
```

**Parameters:**
- `<playername>`: Name of the player to unban (no spaces in name)

**Example:**
```
$unban PlayerName
```

**Expected Output:** `Player PlayerName unbanned!`

**Test:**
1. First ban a player: `$ban TestUser Test reason`
2. Unban the player: `$unban TestUser`
3. Try connecting with this account (should succeed)
4. Check Discord message and server logs

---

### config

**Description:** Manages the FactoCord configuration directly via Discord.

**Permissions:** Admin

**Usage:**
```
$config save
$config load
$config get [path]
$config set <path> [value]
```

**Subcommands:**

#### $config save
Saves the current configuration from memory to `config.json`.

**Example:**
```
$config save
```

**Expected Output:** `Config saved`

**Test:** 
1. Change a setting: `$config set prefix !`
2. Save: `$config save`
3. Check the `config.json` file on the server

---

#### $config load
Reloads the configuration from `config.json`. Unsaved changes will be lost.

**Example:**
```
$config load
```

**Expected Output:** `Config reloaded`

**Test:**
1. Reload configuration: `$config load`
2. Verify that changes took effect

---

#### $config get [path]
Displays the value of a configuration setting. Path elements are separated by dots.

**Examples:**
```
$config get
$config get prefix
$config get admin_ids
$config get admin_ids.0
$config get command_roles
$config get command_roles.mod
$config get messages
$config get messages.server_save
```

**Expected Output:** JSON-formatted display of the configuration value

**Test:**
1. Show entire config object: `$config get`
2. Show prefix: `$config get prefix`
3. Show admin IDs: `$config get admin_ids`

---

#### $config set <path> [value]
Sets the value of a configuration setting. Changes are only saved permanently after `$config save`.

**Important Notes:**
- To add a value to an array, use `*` as index: `$config set admin_ids.* 123456`
- Put strings with spaces in quotes: `$config set game_name "My Server"`
- Without value, the entry is deleted or set to null

**Examples:**
```
$config set prefix !
$config set game_name "Factorio 1.0"
$config set ingame_discord_user_colors true
$config set admin_ids.0 123456789
$config set admin_ids.* 987654321
$config set command_roles.mod 55555555
$config set messages.server_save "**:mango: Game saved!**"
```

**Expected Output:** `Value set`

**Test Sequence:**
1. Show current prefix: `$config get prefix`
2. Change prefix: `$config set prefix !`
3. Show new prefix: `$config get prefix`
4. Save: `$config save`
5. Test new prefix: `!version`
6. Reset to default: `!config set prefix $` and `!config save`

---

### mod

**Description:** Manages Factorio mods (add, remove, enable, disable, update).

**Permissions:** Admin

**Important Notes:**
- Mod names with spaces must be in quotes: `"Squeak Through"`
- Multiple mods can be processed simultaneously (separated by spaces)
- Versions can be specified with `==` (FactoCord-specific syntax): `FNEI==0.3.4`
- Mods must be compatible with the Factorio version

**Usage:**
```
$mod add <modname>+
$mod update [modname]+
$mod remove <modname>+
$mod enable <modname>+
$mod disable <modname>+
```

**Subcommands:**

#### $mod add <modname>+
Adds mods to mod-list.json and downloads the latest or a specific version.

**Examples:**
```
$mod add FNEI
$mod add "Squeak Through"
$mod add FNEI==0.3.4
$mod add FNEI Bottleneck "Squeak Through"
```

**Expected Output:** 
- List of added mods
- Download progress for each mod
- Dependency check and recommendations

**Test:**
1. Add a single mod: `$mod add FNEI`
2. Check download progress in Discord
3. Verify with `$mods files` that the file was downloaded
4. Check `mod-list.json` on the server

---

#### $mod update [modname]+
Updates the specified mods or all mods to the latest version.

**Examples:**
```
$mod update
$mod update FNEI
$mod update FNEI==0.3.5
$mod update FNEI Bottleneck
```

**Expected Output:**
- List of updated mods with old and new version numbers
- Download progress
- List of already up-to-date mods
- Dependency warnings if any

**Test:**
1. Show installed mods: `$mods files`
2. Update all mods: `$mod update`
3. Verify new versions: `$mods files`

---

#### $mod remove <modname>+
Removes mods from mod-list.json and deletes the mod files.

**Examples:**
```
$mod remove FNEI
$mod remove FNEI Bottleneck
$mod remove "Squeak Through"
```

**Expected Output:**
- List of removed mods
- List of deleted files
- Message if mod not found

**Test:**
1. Remove a mod: `$mod remove FNEI`
2. Confirm with `$mods files` that the file was deleted
3. Check `mod-list.json` on the server

---

#### $mod enable <modname>+
Enables mods in mod-list.json (mods must already be installed).

**Examples:**
```
$mod enable FNEI
$mod enable FNEI Bottleneck
```

**Expected Output:** `Enabled mod "FNEI"` or list of enabled mods

**Test:**
1. Disable a mod: `$mod disable FNEI`
2. Check status: `$mods all`
3. Enable again: `$mod enable FNEI`
4. Confirm with `$mods all`

---

#### $mod disable <modname>+
Disables mods in mod-list.json (files remain).

**Examples:**
```
$mod disable FNEI
$mod disable FNEI Bottleneck
```

**Expected Output:** `Disabled mod "FNEI"` or list of disabled mods

**Test:**
1. Disable a mod: `$mod disable FNEI`
2. Check with `$mods all` that the mod is shown as disabled
3. Restart server and verify that the mod is not loaded

---

## Utility Commands

### mods

**Description:** Shows information about installed mods.

**Permissions:** All users

**Usage:**
```
$mods [on|off|all|files]
```

**Subcommands:**

#### $mods or $mods on
Shows currently enabled mods from mod-list.json.

**Example:**
```
$mods
$mods on
```

**Expected Output:** List of all enabled mods with count

**Test:** Run `$mods` and compare with contents of `mod-list.json`

---

#### $mods off
Shows currently disabled mods from mod-list.json.

**Example:**
```
$mods off
```

**Expected Output:** List of all disabled mods

**Test:**
1. Disable a mod: `$mod disable FNEI`
2. Run: `$mods off`
3. FNEI should appear in the list

---

#### $mods all
Shows all mods from mod-list.json (enabled and disabled).

**Example:**
```
$mods all
```

**Expected Output:** Complete list with status indicator (enabled/disabled)

**Test:** Run `$mods all` and check if the output matches the contents of `mod-list.json`

---

#### $mods files
Shows all downloaded mod files in the mod directory.

**Example:**
```
$mods files
```

**Expected Output:** List of all `.zip` mod files with version numbers

**Test:** 
1. Run `$mods files`
2. Check the mod directory on the server
3. Compare the lists

---

### version

**Description:** Shows the Factorio server version and FactoCord version.

**Permissions:** All users

**Usage:**
```
$version
```

**Example:**
```
$version
```

**Expected Output:** 
```
Server version: **1.1.100**
FactoCord version: **3.0.x**
```

**Test:** Run `$version` and compare the Factorio version with the actual server version (verifiable by directly running `factorio --version` on the server)

---

### info

**Description:** Shows server information from the Factorio lobby (name, description, version, tags, players).

**Permissions:** All users

**Important:** Server must be running and registered with the Factorio lobby.

**Usage:**
```
$info
```

**Example:**
```
$info
```

**Expected Output:** 
- Embed with server name
- Description
- Factorio version
- Tags
- Online players

**Test:**
1. Start the server: `$server start`
2. Wait for server to register with lobby (~30 seconds)
3. Run: `$info`
4. Compare with Factorio multiplayer browser view

---

### online

**Description:** Shows currently online players.

**Permissions:** All users

**Important:** Server must be running and registered with the Factorio lobby.

**Usage:**
```
$online
```

**Example:**
```
$online
```

**Expected Output:** 
- List of online players
- Player count / Maximum players (if set)
- Or: `**No one is online**`

**Test:**
1. Start the server: `$server start`
2. Connect to the server with a player
3. Run: `$online`
4. Your player name should appear in the list

---

### help

**Description:** Shows all available commands or detailed help for a specific command.

**Permissions:** All users

**Usage:**
```
$help
$help <command>
$help <command> <subcommand>
```

**Examples:**

#### $help
Shows all available commands with short description.

**Example:**
```
$help
```

**Expected Output:** Embed with all commands and their descriptions

**Test:** Run `$help` and verify that all commands are listed

---

#### $help <command>
Shows detailed help for a specific command.

**Examples:**
```
$help server
$help mod
$help config
$help mods
```

**Expected Output:** 
- Usage examples
- Documentation
- List of subcommands (if any)

**Test:**
1. Help for server: `$help server`
2. Check if all subcommands are listed

---

#### $help <command> <subcommand>
Shows detailed help for a subcommand.

**Examples:**
```
$help server update
$help mod add
$help config set
```

**Expected Output:** Detailed documentation of the specific subcommand

**Test:** Run `$help mod add` and check the documentation

---

## Test Plan - Complete Command Testing

Here is a recommended order for testing all commands:

### 1. Basic Tests (Utility Commands)
```
$help
$version
$mods
```

### 2. Server Management
```
$server
$server stop
$server
$server start
$server
```

### 3. Server Info (Server must be running)
```
$info
$online
```

### 4. Mod Management
```
$mods all
$mod add FNEI
$mods files
$mod disable FNEI
$mods off
$mod enable FNEI
$mod update FNEI
$mod remove FNEI
```

### 5. Config Management
```
$config get prefix
$config set prefix !
$config get prefix
!config save
!config load
!config set prefix $
$config save
```

### 6. Player Management (requires test account)
```
$kick TestUser Test reason
$ban TestUser Test reason
$unban TestUser
```

### 7. Game Functions
```
$save
```

### 8. Advanced Tests
```
$help server
$help mod add
$server update
$mod update
```

---

## Troubleshooting

### "You are not an admin!"
- Verify that your Discord ID is entered in `config.json` under `admin_ids`
- Reload the config with `$config load` (if recently changed)

### "The server is not running"
- Start the server with `$server start`
- Check server logs in `error.log` and `factorio.log`

### "The server did not register a game on the factorio server"
- Server must be public and registered with the Factorio lobby
- Check `launch_parameters` in `config.json`
- Wait ~30 seconds after startup

### "Error reading mod list"
- Check `mod_list_location` in `config.json`
- Ensure the path is correct and the file exists

### "No token to download mods" / "No username to download mods"
- Set `username` and `mod_portal_token` in `config.json`
- Token available at: https://factorio.com/profile

---

## Factorio API Compatibility

The Factorio events used in `control.lua` are compatible with all Factorio versions from 1.0:

- ✅ `defines.events.on_player_joined_game`
- ✅ `defines.events.on_player_left_game`
- ✅ `defines.events.on_console_chat`
- ✅ `defines.events.on_player_died`
- ✅ `defines.events.on_player_kicked`
- ✅ `defines.events.on_player_unbanned`
- ✅ `defines.events.on_player_unmuted`
- ✅ `defines.events.on_player_banned`
- ✅ `defines.events.on_player_muted`

These events are part of the stable Factorio Lua API and are fully supported in current versions (1.1.x, 2.0.x).

---

## DM Mod-Settings Manager

The Mod-Settings Manager allows verified users to manage Factorio mod settings directly via Discord DMs. Changes are applied with a server restart.

### Requirements
- User must be verified (Discord-Factorio account linked)
- `enable_dm_chat: true` in config.json
- Bot must have DM permissions

### Verification
Before using the Mod-Settings Manager, link your Discord account with your Factorio player name:
1. In-game: Type `$$link` in chat
2. A verification code will be shown
3. DM the bot with the code
4. Your accounts are now linked

---

### mods (DM)

**Description:** Opens the interactive mod settings editor in DMs.

**Usage:**
```
!mods
```

**Expected Output:** 
- Interactive embed with all mods that have configurable settings
- Select menu to choose a mod
- Pagination for many mods

**Features:**
- 📋 Table-style mod listing showing name, status, game/map setting counts
- 🔍 Select menu for quick mod selection
- ◀️ ▶️ Pagination for large mod lists
- 🔄 Refresh button to reload mods

---

### modshelp

**Description:** Shows help for the Mod-Settings Manager.

**Usage:**
```
!modshelp
```

**Expected Output:** Detailed help with available commands and workflow explanation.

---

### save (DM)

**Description:** Shows preview of pending changes and allows saving.

**Usage:**
```
!save
```

**Expected Output:**
- Summary of all pending setting changes
- Confirm/Cancel buttons
- Warning about server restart

**Process:**
1. Review changes in the preview
2. Click "✅ Speichern & Neustarten" to apply
3. Server restarts automatically
4. Changes are backed up before applying

---

### cancel

**Description:** Cancels the current editing session and discards all changes.

**Usage:**
```
!cancel
```

**Expected Output:** Confirmation message, session is cleared.

---

### Editing Workflow

1. **Start**: Send `!mods` in a DM to the bot
2. **Select Mod**: Choose a mod from the dropdown menu
3. **Select Tab**: Choose between Game Settings (🎮) or Map Settings (🗺️)
4. **Edit Setting**: Select a setting from the dropdown to edit
5. **Enter Value**: Enter the new value in the modal dialog
6. **Preview**: Click "💾 Speichern" to see changes preview
7. **Confirm**: Click "✅ Speichern & Neustarten" to apply changes

### Setting Types
- **🔘 Boolean**: Enter `true` or `false`
- **🔢 Number**: Enter integer or decimal number
- **📝 Text**: Enter string value

### Backup System
Before applying changes, the bot automatically:
- Creates a backup of `mod-settings.dat`
- Stores it in `./backups/` with timestamp
- Maximum 10 backups are kept

---

## Additional Notes

### Command Prefix
The default prefix is `$`, but can be changed in `config.json`:
```json
"prefix": "$"
```

### DM Configuration
Enable DM features in `config.json`:
```json
"enable_dm_chat": true,
"verification_data_path": "./verification.json"
```

### Role-Based Permissions
Certain commands can be bound to Discord roles in `config.json`:
```json
"command_roles": {
    "mod": "role_id_here"
}
```

### Customizable Messages
Many bot messages can be customized in `config.json` under `messages`.

---

**FactoCord Version:** 3.0  
**Last Updated:** December 16, 2024  
**Documentation Created For:** Factorio 1.1.x / 2.0.x Compatibility
