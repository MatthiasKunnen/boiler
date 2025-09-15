# Steamcmd login
Games can require a logged-in user before allowing downloading.
Steam recommends creating a separate account to log in to steamcmd in order not to expose account credentials on the server. 
However, it seems that to download workshop items, you need to own the respective game.
To that end, boiler will be built from the ground up to securely handle main account credentials.
Credentials will be purged after each invocation depending on a flag.

## Different login paths
This contains the output of steamcmd in different authentication cases.

### Email code prompt
```
Steam>login username
Cached credentials not found.

password: 
Proceeding with login using username/password.
Logging in user 'username' [U:1:0] to Steam Public...
That Steam Guard code was invalid.
Please check your email for the message from Steam, and enter the Steam Guard
 code from that message.
You can also enter this code at any time using 'set_steam_guard_code'
 at the console.
Steam Guard code:
```

### Steam guard prompt
```
Steam>login username
Cached credentials not found.

password: 
Proceeding with login using username/password.
Logging in user 'username' [U:1:0] to Steam Public...This account is protected by a Steam Guard mobile authenticator.
Please confirm the login in the Steam Mobile app on your phone.
```

#### Steam guard prompt without checking _remember on this device_
```
Steam>login username
Cached credentials not found.

password: 
Proceeding with login using username/password.
Logging in user 'username' [U:1:0] to Steam Public...This account is protected by a Steam Guard mobile authenticator.
Please confirm the login in the Steam Mobile app on your phone.

Waiting for confirmation...
Waiting for confirmation...
Waiting for confirmation...
Waiting for confirmation...
Waiting for confirmation...
Waiting for confirmation...
Waiting for confirmation...
Warning: Login token expires at Tue Oct 14 03:59:33 2025
If using cached credentials, this may indicate your client cannot renew its token.
This can occur if steamcmd cannot write the machine auth token to its config.vdf.
OK
Waiting for client config...OK
Waiting for user info...
```

### Cached credentials
```
Steam>login username
Logging in using cached credentials.
Logging in user 'username' [U:1:xxxxxxxxxx] to Steam Public...OK
Waiting for client config...OK
Waiting for user info...OK
```
