# irspamd
Simple yet powerful anti spam solution for IMAP-Mail accounts using [rspamd](https://github.com/vstakhov/rspamd).

## install
```
go get github.com/denkhaus/irspamd
```

## scan

This scans your inbox 'INBOX' and moves spam to 'Spam' and ham to the 'Mail' box.(-m) 
If you do not specify -m or --hambox, messages remain in inbox.

### parameters 

--expunge, -e Expunge all spam messages from inbox after scan has finished.
--spambox, -s Name of the box to move spam messages to.
--hambox,  -m Name of the box to move ham messages to. If no hambox is given, ham remains in inbox.
--inbox,   -i Name of the box to be scanned.

After scanning, processed messages get expunged. (-e)
```
IMAP_PASSWORD='yourpassword' irspamd -u user@name.com -H imap.host.net scan -m Mail -e
```

