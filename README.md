# irspamd
Powerful golang anti spam solution for IMAP-Mail accounts using [rspamd](https://github.com/vstakhov/rspamd).
Tested with rspamd version 1.2.8
Use this software at your own risk. I am not responsible for damage or loss of data.

## prerequisites
Install rspamd as described [here](https://rspamd.com/downloads.html). Make sure,
rspamd is running and rspamc is on the PATH.

## install
```
go get -u github.com/denkhaus/irspamd
```
#### global params
* `--host, -H` IMAP host to connect to. Default is localhost.
* `--port, -P` IMAP port to connect to. Default is 993
* `--user, -u` The username. Value is required.
* `--pass, -p` The password. For security reasons prefer ENV IMAP_PASSWORD='yourpassword'

## scan
```
IMAP_PASSWORD='yourpassword' irspamd -u user@name.com -H imap.host.net scan -m Mail -e
```

This scans unseen messages from your inbox 'INBOX', moves spam to 'Spam' and ham to the 'Mail' box. `-m`
If you do not specify `-m or --hambox`, messages remain in inbox. After scanning, processed messages get expunged from inbox.`-e`

#### params

* `--expunge, -e` Expunge all spam messages from inbox after sucessfull scan. Default is false.
* `--spambox, -s` Name of the box to move spam messages to. Default is 'Spam'.
* `--hambox,  -m` Name of the box to move ham messages to. If no hambox is given, ham remains in inbox.
* `--inbox,   -i` Name of the box to be scanned. Default is 'INBOX'.
* `--force,   -f` Forces processing if message was already checked.

## learn
#### subcommands
* `ham` Learn ham from learnbox.
* `spam` Learn spam from learnbox.

#### params

* `--learnbox, -l` Name of the box to be scanned for learning. Required.
* `--markseen, -s` Mark learned messages as seen on learnbox.

####learn ham from learnbox 'Mail'
```
IMAP_PASSWORD='yourpassword' irspamd -u user@name.com -H imap.host.net learn ham -l Mail
```
####learn spam from learnbox 'Spam'
```
IMAP_PASSWORD='yourpassword' irspamd -u user@name.com -H imap.host.net learn spam -l Spam
```
