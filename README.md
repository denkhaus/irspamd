# irspamd
Simple yet powerful anti spam solution for IMAP-Mail accounts using [rspamd](https://github.com/vstakhov/rspamd).
Use this software at your own risk. I am not responsible for damage or loss of data.

## prerequisites
Install rspamd as described [here](https://rspamd.com/downloads.html). Make sure, rspamd is on the PATH.
## install
```
go get github.com/denkhaus/irspamd
```
#### global params
* `--host, -H` Host to connect to. Default is localhost.
* `--port, -P` Port number to connect to. Default is 993
* `--user, -u` The username. Value is required.
* `--pass, -p` The password. For security reasons prefer ENV IMAP_PASSWORD='yourpassword'
* `--reset -r` Reset the database before processing messages.

## scan
```
IMAP_PASSWORD='yourpassword' irspamd -u user@name.com -H imap.host.net scan -m Mail -e
```

This scans your inbox 'INBOX' and moves spam to 'Spam' and ham to the 'Mail' box. `-m`
If you do not specify `-m or --hambox`, messages remain in inbox. After scanning, processed messages get expunged.`-e`

#### params

* `--expunge, -e` Expunge all spam messages from inbox after scan has finished.
* `--spambox, -s` Name of the box to move spam messages to.
* `--hambox,  -m` Name of the box to move ham messages to. If no hambox is given, ham remains in inbox.
* `--inbox,   -i` Name of the box to be scanned.


## learn
#### subcommands
* `ham` Learn ham from learnbox.
* `spam` Learn spam from learnbox.

#### parameters

* `--learnbox, -l` Name of the box to be scanned for learning. Required.

####learn ham
```
IMAP_PASSWORD='yourpassword' irspamd -u user@name.com -H imap.host.net learn ham -l Mail
```
####learn spam
```
IMAP_PASSWORD='yourpassword' irspamd -u user@name.com -H imap.host.net learn spam -l Spam
```
