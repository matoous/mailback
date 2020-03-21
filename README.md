# Mailback

Mailback is service that accepts emails and sends them back to the original
sender either after requested time has elapsed or periodically.

## Components

Mailback consists of 3 separate components that need to run in order
for everything to work.

### Receiver

Receiver handles incoming emails. Validates the time interval (or period)
and saves the emails and schedules them for delivery back.

### Sender

Sender retrieves the emails from database when its time for them to be send
back and sends them to the original sender.

### Web Server

Web server is necessary to allow the users to unsubscribe from periodic emails.

## Setup

1. Get a domain - get a domain for your server.
2. Get a server - get a VPS or some other server to host the application.
3. Generate DKIM private and public key
4. Add dkim record for your domain
5. Add SPF record
6. Generate API key on Cloudflare, add it on server
7. Expose port 25

