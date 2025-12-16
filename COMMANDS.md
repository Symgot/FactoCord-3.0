# FactoCord 3.0 - Discord Commands Tutorial

Dieses Dokument beschreibt alle verfügbaren Discord-Commands für FactoCord 3.0 mit detaillierten Beispielen und Testanweisungen.

## Inhaltsverzeichnis
- [Übersicht](#übersicht)
- [Admin-Commands](#admin-commands)
  - [server](#server)
  - [save](#save)
  - [kick](#kick)
  - [ban](#ban)
  - [unban](#unban)
  - [config](#config)
  - [mod](#mod)
- [Utility-Commands](#utility-commands)
  - [mods](#mods)
  - [version](#version)
  - [info](#info)
  - [online](#online)
  - [help](#help)

## Übersicht

FactoCord 3.0 bietet verschiedene Discord-Commands zur Verwaltung und Überwachung deines Factorio-Servers. Commands sind in zwei Kategorien unterteilt:

- **Admin-Commands**: Erfordern Admin-Rechte (konfiguriert in `config.json` unter `admin_ids`) oder eine spezifische Rolle
- **Utility-Commands**: Für alle Benutzer verfügbar

Standardmäßig ist das Command-Prefix `$` (konfigurierbar in `config.json`).

---

## Admin-Commands

### server

**Beschreibung:** Verwaltet den Factorio-Server (Start, Stop, Neustart, Updates).

**Berechtigungen:** 
- Status-Abfrage: Alle Benutzer
- Steuerung: Nur Admins

**Verwendung:**
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
Zeigt den aktuellen Status des Servers an (läuft oder gestoppt).

**Beispiel:**
```
$server
```
**Erwartete Ausgabe:** `Factorio server is **running**` oder `Factorio server is **stopped**`

---

#### $server stop
Stoppt den laufenden Factorio-Server.

**Beispiel:**
```
$server stop
```
**Erwartete Ausgabe:** Server wird heruntergefahren, Status-Nachricht wird gesendet.

**Test:** Führe nach dem Stoppen `$server` aus, um zu bestätigen, dass der Server gestoppt ist.

---

#### $server start
Startet den Factorio-Server.

**Beispiel:**
```
$server start
```
**Erwartete Ausgabe:** Server wird gestartet, Status-Nachricht wird gesendet.

**Test:** Führe nach dem Start `$server` aus, um zu bestätigen, dass der Server läuft.

---

#### $server restart
Startet den Factorio-Server neu (Stop + Start).

**Beispiel:**
```
$server restart
```
**Erwartete Ausgabe:** Server wird gestoppt und dann neu gestartet.

**Test:** Überprüfe mit `$server`, ob der Server nach dem Neustart wieder läuft.

---

#### $server update [version]
Aktualisiert den Factorio-Server auf die neueste oder eine spezifische Version.

**Hinweis:** Server muss vor dem Update gestoppt sein.

**Beispiele:**
```
$server update
$server update 1.1.100
```

**Erwartete Ausgabe:** 
- Download-Fortschritt wird angezeigt
- Entpacken-Nachricht
- Erfolgsmeldung mit Versionsnummer

**Test:**
1. Stoppe den Server: `$server stop`
2. Update auf neueste Version: `$server update`
3. Überprüfe Version nach dem Start: `$version`

---

#### $server install [version]
Installiert eine Factorio-Server-Version ohne Versionsprüfung (ähnlich wie update, aber ohne Check der aktuellen Version).

**Beispiele:**
```
$server install
$server install 1.1.100
```

**Test:** Gleich wie bei `$server update`

---

### save

**Beschreibung:** Speichert das aktuelle Spiel.

**Berechtigungen:** Admin

**Verwendung:**
```
$save
```

**Beispiel:**
```
$save
```

**Erwartete Ausgabe:** Speichervorgang wird auf dem Factorio-Server ausgeführt, Bestätigungsnachricht in Discord.

**Test:** 
1. Führe `$save` aus
2. Überprüfe die Factorio-Server-Logs auf `/save` Command
3. Achte auf die Discord-Nachricht, die den erfolgreichen Speichervorgang bestätigt

---

### kick

**Beschreibung:** Wirft einen Spieler vom Server mit einem angegebenen Grund.

**Berechtigungen:** Admin

**Verwendung:**
```
$kick <spielername> <grund>
```

**Parameter:**
- `<spielername>`: Der Name des Spielers, der gekickt werden soll
- `<grund>`: Der Grund für den Kick (mehrere Wörter erlaubt)

**Beispiel:**
```
$kick PlayerName Regelverstoß
$kick TestUser Inaktivität seit 30 Minuten
```

**Erwartete Ausgabe:** `Player PlayerName kicked with reason Regelverstoß!`

**Test:**
1. Verbinde dich mit dem Server mit einem Test-Account
2. Führe `$kick TestUser Testgrund` aus
3. Bestätige, dass der Spieler vom Server getrennt wurde
4. Überprüfe die Kick-Nachricht in Discord und im Factorio-Server

---

### ban

**Beschreibung:** Bannt einen Spieler vom Server mit einem angegebenen Grund.

**Berechtigungen:** Admin

**Verwendung:**
```
$ban <spielername> <grund>
```

**Parameter:**
- `<spielername>`: Der Name des Spielers, der gebannt werden soll
- `<grund>`: Der Grund für den Bann (mehrere Wörter erlaubt)

**Beispiel:**
```
$ban PlayerName Cheating
$ban GrieferUser Griefing und wiederholte Verstöße
```

**Erwartete Ausgabe:** `Player PlayerName banned with reason "Cheating"!`

**Test:**
1. Banne einen Test-Benutzer: `$ban TestUser Testbann`
2. Versuche dich mit diesem Account zu verbinden (sollte fehlschlagen)
3. Überprüfe die Ban-Liste auf dem Server
4. Entferne den Bann mit `$unban TestUser` nach dem Test

---

### unban

**Beschreibung:** Entfernt einen Spieler von der Ban-Liste.

**Berechtigungen:** Admin

**Verwendung:**
```
$unban <spielername>
```

**Parameter:**
- `<spielername>`: Der Name des Spielers, der entbannt werden soll (keine Leerzeichen im Namen)

**Beispiel:**
```
$unban PlayerName
```

**Erwartete Ausgabe:** `Player PlayerName unbanned!`

**Test:**
1. Banne zuerst einen Spieler: `$ban TestUser Testgrund`
2. Entbanne den Spieler: `$unban TestUser`
3. Versuche dich mit diesem Account zu verbinden (sollte erfolgreich sein)
4. Überprüfe die Discord-Nachricht und Server-Logs

---

### config

**Beschreibung:** Verwaltet die FactoCord-Konfiguration direkt über Discord.

**Berechtigungen:** Admin

**Verwendung:**
```
$config save
$config load
$config get [pfad]
$config set <pfad> [wert]
```

**Subcommands:**

#### $config save
Speichert die aktuelle Konfiguration aus dem Speicher in `config.json`.

**Beispiel:**
```
$config save
```

**Erwartete Ausgabe:** `Config saved`

**Test:** 
1. Ändere eine Einstellung: `$config set prefix !`
2. Speichere: `$config save`
3. Überprüfe die `config.json` Datei auf dem Server

---

#### $config load
Lädt die Konfiguration aus `config.json` neu. Nicht gespeicherte Änderungen gehen verloren.

**Beispiel:**
```
$config load
```

**Erwartete Ausgabe:** `Config reloaded`

**Test:**
1. Lade die Konfiguration neu: `$config load`
2. Überprüfe, ob Änderungen wirksam wurden

---

#### $config get [pfad]
Zeigt den Wert einer Konfigurationseinstellung an. Pfad-Elemente werden durch Punkte getrennt.

**Beispiele:**
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

**Erwartete Ausgabe:** JSON-formatierte Anzeige des Konfigurationswertes

**Test:**
1. Zeige das gesamte Config-Objekt: `$config get`
2. Zeige den Prefix: `$config get prefix`
3. Zeige Admin-IDs: `$config get admin_ids`

---

#### $config set <pfad> [wert]
Setzt den Wert einer Konfigurationseinstellung. Änderungen werden erst nach `$config save` dauerhaft gespeichert.

**Wichtige Hinweise:**
- Um einen Wert zu einem Array hinzuzufügen, verwende `*` als Index: `$config set admin_ids.* 123456`
- Strings mit Leerzeichen in Anführungszeichen setzen: `$config set game_name "Mein Server"`
- Ohne Wert wird der Eintrag gelöscht oder auf Null gesetzt

**Beispiele:**
```
$config set prefix !
$config set game_name "Factorio 1.0"
$config set ingame_discord_user_colors true
$config set admin_ids.0 123456789
$config set admin_ids.* 987654321
$config set command_roles.mod 55555555
$config set messages.server_save "**:mango: Game saved!**"
```

**Erwartete Ausgabe:** `Value set`

**Test-Sequenz:**
1. Zeige aktuellen Prefix: `$config get prefix`
2. Ändere Prefix: `$config set prefix !`
3. Zeige neuen Prefix: `$config get prefix`
4. Speichere: `$config save`
5. Teste neuen Prefix: `!version`
6. Setze zurück auf Standard: `!config set prefix $` und `!config save`

---

### mod

**Beschreibung:** Verwaltet Factorio-Mods (Hinzufügen, Entfernen, Aktivieren, Deaktivieren, Aktualisieren).

**Berechtigungen:** Admin

**Wichtige Hinweise:**
- Mod-Namen mit Leerzeichen müssen in Anführungszeichen gesetzt werden: `"Squeak Through"`
- Mehrere Mods können gleichzeitig verarbeitet werden (durch Leerzeichen getrennt)
- Versionen können mit `==` angegeben werden: `FNEI==0.3.4`
- Mods müssen mit der Factorio-Version kompatibel sein

**Verwendung:**
```
$mod add <modname>+
$mod update [modname]+
$mod remove <modname>+
$mod enable <modname>+
$mod disable <modname>+
```

**Subcommands:**

#### $mod add <modname>+
Fügt Mods zu mod-list.json hinzu und lädt die neueste oder eine spezifische Version herunter.

**Beispiele:**
```
$mod add FNEI
$mod add "Squeak Through"
$mod add FNEI==0.3.4
$mod add FNEI Bottleneck "Squeak Through"
```

**Erwartete Ausgabe:** 
- Liste der hinzugefügten Mods
- Download-Fortschritt für jeden Mod
- Abhängigkeitsprüfung und Empfehlungen

**Test:**
1. Füge einen einzelnen Mod hinzu: `$mod add FNEI`
2. Überprüfe Download-Fortschritt in Discord
3. Verifiziere mit `$mods files`, dass die Datei heruntergeladen wurde
4. Überprüfe `mod-list.json` auf dem Server

---

#### $mod update [modname]+
Aktualisiert die angegebenen Mods oder alle Mods auf die neueste Version.

**Beispiele:**
```
$mod update
$mod update FNEI
$mod update FNEI==0.3.5
$mod update FNEI Bottleneck
```

**Erwartete Ausgabe:**
- Liste der aktualisierten Mods mit alten und neuen Versionsnummern
- Download-Fortschritt
- Liste bereits aktueller Mods
- Abhängigkeitswarnungen falls vorhanden

**Test:**
1. Zeige installierte Mods: `$mods files`
2. Update alle Mods: `$mod update`
3. Verifiziere neue Versionen: `$mods files`

---

#### $mod remove <modname>+
Entfernt Mods aus mod-list.json und löscht die Mod-Dateien.

**Beispiele:**
```
$mod remove FNEI
$mod remove FNEI Bottleneck
$mod remove "Squeak Through"
```

**Erwartete Ausgabe:**
- Liste der entfernten Mods
- Liste der gelöschten Dateien
- Meldung falls Mod nicht gefunden

**Test:**
1. Entferne einen Mod: `$mod remove FNEI`
2. Bestätige mit `$mods files`, dass die Datei gelöscht wurde
3. Überprüfe `mod-list.json` auf dem Server

---

#### $mod enable <modname>+
Aktiviert Mods in mod-list.json (Mods müssen bereits installiert sein).

**Beispiele:**
```
$mod enable FNEI
$mod enable FNEI Bottleneck
```

**Erwartete Ausgabe:** `Enabled mod "FNEI"` oder Liste aktivierter Mods

**Test:**
1. Deaktiviere einen Mod: `$mod disable FNEI`
2. Überprüfe Status: `$mods all`
3. Aktiviere wieder: `$mod enable FNEI`
4. Bestätige mit `$mods all`

---

#### $mod disable <modname>+
Deaktiviert Mods in mod-list.json (Dateien bleiben erhalten).

**Beispiele:**
```
$mod disable FNEI
$mod disable FNEI Bottleneck
```

**Erwartete Ausgabe:** `Disabled mod "FNEI"` oder Liste deaktivierter Mods

**Test:**
1. Deaktiviere einen Mod: `$mod disable FNEI`
2. Überprüfe mit `$mods all`, dass der Mod als deaktiviert angezeigt wird
3. Starte Server neu und verifiziere, dass der Mod nicht geladen wird

---

## Utility-Commands

### mods

**Beschreibung:** Zeigt Informationen über installierte Mods an.

**Berechtigungen:** Alle Benutzer

**Verwendung:**
```
$mods [on|off|all|files]
```

**Subcommands:**

#### $mods oder $mods on
Zeigt aktuell aktivierte Mods aus mod-list.json.

**Beispiel:**
```
$mods
$mods on
```

**Erwartete Ausgabe:** Liste aller aktivierten Mods mit Zählung

**Test:** Führe `$mods` aus und vergleiche mit dem Inhalt von `mod-list.json`

---

#### $mods off
Zeigt aktuell deaktivierte Mods aus mod-list.json.

**Beispiel:**
```
$mods off
```

**Erwartete Ausgabe:** Liste aller deaktivierten Mods

**Test:**
1. Deaktiviere einen Mod: `$mod disable FNEI`
2. Führe aus: `$mods off`
3. FNEI sollte in der Liste erscheinen

---

#### $mods all
Zeigt alle Mods aus mod-list.json (aktiviert und deaktiviert).

**Beispiel:**
```
$mods all
```

**Erwartete Ausgabe:** Vollständige Liste mit Markierung des Status (aktiviert/deaktiviert)

**Test:** Führe `$mods all` aus und überprüfe, ob die Ausgabe dem Inhalt von `mod-list.json` entspricht

---

#### $mods files
Zeigt alle heruntergeladenen Mod-Dateien im Mod-Verzeichnis.

**Beispiel:**
```
$mods files
```

**Erwartete Ausgabe:** Liste aller `.zip` Mod-Dateien mit Versionsnummern

**Test:** 
1. Führe `$mods files` aus
2. Überprüfe das Mod-Verzeichnis auf dem Server
3. Vergleiche die Listen

---

### version

**Beschreibung:** Zeigt die Factorio-Server-Version und die FactoCord-Version an.

**Berechtigungen:** Alle Benutzer

**Verwendung:**
```
$version
```

**Beispiel:**
```
$version
```

**Erwartete Ausgabe:** 
```
Server version: **1.1.100**
FactoCord version: **3.0.x**
```

**Test:** Führe `$version` aus und vergleiche die Factorio-Version mit der tatsächlichen Server-Version (prüfbar durch direktes Ausführen von `factorio --version` auf dem Server)

---

### info

**Beschreibung:** Zeigt Serverinformationen aus der Factorio-Lobby an (Name, Beschreibung, Version, Tags, Spieler).

**Berechtigungen:** Alle Benutzer

**Wichtig:** Server muss laufen und bei der Factorio-Lobby registriert sein.

**Verwendung:**
```
$info
```

**Beispiel:**
```
$info
```

**Erwartete Ausgabe:** 
- Embed mit Servernamen
- Beschreibung
- Factorio-Version
- Tags
- Online-Spieler

**Test:**
1. Starte den Server: `$server start`
2. Warte, bis Server bei der Lobby registriert ist (~30 Sekunden)
3. Führe aus: `$info`
4. Vergleiche mit der Factorio-Multiplayer-Browser-Ansicht

---

### online

**Beschreibung:** Zeigt die aktuell online Spieler an.

**Berechtigungen:** Alle Benutzer

**Wichtig:** Server muss laufen und bei der Factorio-Lobby registriert sein.

**Verwendung:**
```
$online
```

**Beispiel:**
```
$online
```

**Erwartete Ausgabe:** 
- Liste der Online-Spieler
- Anzahl Spieler / Maximale Spieler (falls gesetzt)
- Oder: `**No one is online**`

**Test:**
1. Starte den Server: `$server start`
2. Verbinde dich mit dem Server mit einem Spieler
3. Führe aus: `$online`
4. Dein Spielername sollte in der Liste erscheinen

---

### help

**Beschreibung:** Zeigt alle verfügbaren Commands oder detaillierte Hilfe zu einem spezifischen Command an.

**Berechtigungen:** Alle Benutzer

**Verwendung:**
```
$help
$help <command>
$help <command> <subcommand>
```

**Beispiele:**

#### $help
Zeigt alle verfügbaren Commands mit kurzer Beschreibung.

**Beispiel:**
```
$help
```

**Erwartete Ausgabe:** Embed mit allen Commands und deren Beschreibungen

**Test:** Führe `$help` aus und überprüfe, ob alle Commands aufgelistet sind

---

#### $help <command>
Zeigt detaillierte Hilfe zu einem spezifischen Command.

**Beispiele:**
```
$help server
$help mod
$help config
$help mods
```

**Erwartete Ausgabe:** 
- Verwendungsbeispiele
- Dokumentation
- Liste der Subcommands (falls vorhanden)

**Test:**
1. Hilfe für server: `$help server`
2. Überprüfe, ob alle Subcommands aufgelistet sind

---

#### $help <command> <subcommand>
Zeigt detaillierte Hilfe zu einem Subcommand.

**Beispiele:**
```
$help server update
$help mod add
$help config set
```

**Erwartete Ausgabe:** Detaillierte Dokumentation des spezifischen Subcommands

**Test:** Führe `$help mod add` aus und überprüfe die Dokumentation

---

## Testplan - Vollständiger Command-Test

Hier ist eine empfohlene Reihenfolge zum Testen aller Commands:

### 1. Basis-Tests (Utility Commands)
```
$help
$version
$mods
```

### 2. Server-Management
```
$server
$server stop
$server
$server start
$server
```

### 3. Server-Info (Server muss laufen)
```
$info
$online
```

### 4. Mod-Management
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

### 5. Config-Management
```
$config get prefix
$config set prefix !
$config get prefix
!config save
!config load
!config set prefix $
$config save
```

### 6. Spieler-Management (erfordert Test-Account)
```
$kick TestUser Testgrund
$ban TestUser Testgrund
$unban TestUser
```

### 7. Spiel-Funktionen
```
$save
```

### 8. Erweiterte Tests
```
$help server
$help mod add
$server update
$mod update
```

---

## Fehlerbehebung

### "You are not an admin!"
- Überprüfe, dass deine Discord-ID in `config.json` unter `admin_ids` eingetragen ist
- Reload die Config mit `$config load` (falls kürzlich geändert)

### "The server is not running"
- Starte den Server mit `$server start`
- Überprüfe Server-Logs in `error.log` und `factorio.log`

### "The server did not register a game on the factorio server"
- Server muss öffentlich sein und bei der Factorio-Lobby registriert
- Überprüfe `launch_parameters` in `config.json`
- Warte ~30 Sekunden nach dem Start

### "Error reading mod list"
- Überprüfe `mod_list_location` in `config.json`
- Stelle sicher, dass der Pfad korrekt ist und die Datei existiert

### "No token to download mods" / "No username to download mods"
- Setze `username` und `mod_portal_token` in `config.json`
- Token erhältlich auf: https://factorio.com/profile

---

## Factorio API-Kompatibilität

Die in `control.lua` verwendeten Factorio-Events sind kompatibel mit allen Factorio-Versionen ab 1.0:

- ✅ `defines.events.on_player_joined_game`
- ✅ `defines.events.on_player_left_game`
- ✅ `defines.events.on_console_chat`
- ✅ `defines.events.on_player_died`
- ✅ `defines.events.on_player_kicked`
- ✅ `defines.events.on_player_unbanned`
- ✅ `defines.events.on_player_unmuted`
- ✅ `defines.events.on_player_banned`
- ✅ `defines.events.on_player_muted`

Diese Events sind Teil der stabilen Factorio Lua API und werden in aktuellen Versionen (1.1.x, 2.0.x) vollständig unterstützt.

---

## Zusätzliche Hinweise

### Befehlspräfix
Der Standard-Prefix ist `$`, kann aber in `config.json` geändert werden:
```json
"prefix": "$"
```

### Rollen-basierte Berechtigungen
Bestimmte Commands können an Discord-Rollen gebunden werden in `config.json`:
```json
"command_roles": {
    "mod": "role_id_here"
}
```

### Anpassbare Nachrichten
Viele Bot-Nachrichten können in `config.json` unter `messages` angepasst werden.

---

**FactoCord Version:** 3.0  
**Zuletzt aktualisiert:** 2024  
**Dokumentation erstellt für:** Factorio 1.1.x / 2.0.x Kompatibilität
